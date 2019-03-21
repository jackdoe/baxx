package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackdoe/baxx/api/helpers"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/notification"
	"github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func initDatabase(db *gorm.DB) {
	if err := db.AutoMigrate(
		&user.User{},
		&user.VerificationLink{},
		&file.Token{},
		&file.FileMetadata{},
		&file.FileVersion{},
		&ActionLog{},
		&user.PaymentHistory{},
		&notification.NotificationRule{},
		&notification.NotificationForFileVersion{},
		&notification.NotificationForQuota{},
		&notification.EmailQueueItem{},
	).Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&notification.EmailQueueItem{}).AddIndex("idx_email_sent", "sent").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&user.VerificationLink{}).AddUniqueIndex("idx_user_sent_at", "user_id", "sent_at").Error; err != nil {
		log.Panic(err)
	}

	// not unique index, we can have many links for same email, they could expire
	if err := db.Model(&user.VerificationLink{}).AddIndex("idx_vl_email", "email").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&user.User{}).AddUniqueIndex("idx_payment_id", "payment_id").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&file.FileVersion{}).AddIndex("idx_token_sha", "token_id", "sha256").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&file.FileVersion{}).AddIndex("idx_fv_metadata", "file_metadata_id").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&file.FileMetadata{}).AddUniqueIndex("idx_fm_token_id_path_2", "token_id", "path", "filename").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&notification.NotificationRule{}).AddIndex("idx_n_user_id_token_id", "user_id", "token_id").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&notification.NotificationForFileVersion{}).AddUniqueIndex("idx_nfv_rule_fv", "file_version_id", "notification_rule_id").Error; err != nil {
		log.Panic(err)
	}

	if err := db.Model(&notification.NotificationForQuota{}).AddUniqueIndex("idx_nfq_token", "token_id").Error; err != nil {
		log.Panic(err)
	}

}

type server struct {
	db         *gorm.DB
	store      *file.Store
	r          *gin.Engine
	authorized *gin.RouterGroup
}

func (s *server) registerHelp(protected bool, helpobj help.HelpObject, path ...string) {
	for _, p := range path {
		if protected {
			s.authorized.GET("/help"+p, func(c *gin.Context) {
				c.String(http.StatusOK, help.Render(helpobj))
			})
		} else {
			s.r.GET("/help"+p, func(c *gin.Context) {
				c.String(http.StatusOK, help.Render(helpobj))
			})
		}
	}
}

func (s *server) getViewTokenLoggedOrNot(c *gin.Context) (*file.Token, *user.User, error) {
	token := c.Param("token")
	var t *file.Token
	var u *user.User
	var err error
	x, isLoggedIn := c.Get("user")
	if isLoggedIn {
		u = x.(*user.User)
		t, err = helpers.FindTokenForUser(s.db, token, u)
		if err != nil {
			return nil, nil, err
		}
	} else {
		t, u, err = helpers.FindTokenAndUser(s.db, token)
		if err != nil {
			return nil, nil, err
		}
		writing := c.Request.Method == "POST" || c.Request.Method == "PUT"
		if t.WriteOnly && !writing {
			return nil, nil, fmt.Errorf("write only token, use basic auth curl -X%s-u your.email https://baxx.dev/{io,ls}/$TOKEN/*path", c.Request.Method)
		}
	}

	if u.EmailVerified == nil {
		err := errors.New("email not verified yet")
		return nil, nil, err
	}

	if !u.Paid() {
		err := fmt.Errorf("payment not received yet or subscription is cancelled, go to https://baxx.dev/sub/%s or if you already did, contact me at jack@baxx.dev", u.PaymentID)
		return nil, nil, err
	}

	return t, u, nil
}

func setupAPI(db *gorm.DB, bind string) {
	store, err := file.NewStore(os.Getenv("BAXX_S3_ENDPOINT"), os.Getenv("BAXX_S3_ACCESS_KEY"), os.Getenv("BAXX_S3_SECRET"), os.Getenv("BAXX_S3_DISABLE_SSL") == "true")

	if err != nil {
		log.Panic(err)
	}

	r := gin.Default()
	r.Use(SlackRecovery())
	r.Use(func(c *gin.Context) {
		su, pass := BasicAuthDecode(c)
		if su != "" {
			u, err := user.FindUser(db, su, pass)
			if err == nil {
				if u.MatchPassword(pass) {
					c.Set("user", u)
					actionLog(db, u.ID, c.Request.Method, c.Request.RequestURI, c.Request)
				}
			}
		}
	})

	authorized := r.Group("/protected") // FIXME rename that? its too long /p/ should be fine
	authorized.Use(func(c *gin.Context) {
		_, loggedIn := c.Get("user")
		if !loggedIn {
			c.Header("WWW-Authenticate", "Authorization Required")
			c.String(401, `{"error": "Not Authorized (auth required, or wrong password)"}`)
			c.Abort()
		}
	})

	r.GET("/digitalocean", func(c *gin.Context) {
		c.String(200, "ok")
	})

	srv := &server{db: db, r: r, store: store, authorized: authorized}
	setupIO(srv)
	setupACC(srv)
	setupSYNC(srv)

	log.Panic(r.Run(bind))
}

func main() {
	defer notification.SlackPanic("main api")

	var pbind = flag.String("bind", ":9123", "bind")
	var pdebug = flag.Bool("debug", false, "debug")
	var pinit = flag.Bool("create-tables", false, "create tables")
	var prelease = flag.Bool("release", false, "release")
	flag.Parse()

	CONFIG.MaxTokens = 100

	if *prelease {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()

	if *pinit {
		initDatabase(db)
	}

	setupAPI(db, *pbind)
}
