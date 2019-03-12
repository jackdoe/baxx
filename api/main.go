package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackdoe/baxx/file"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func initDatabase(db *gorm.DB) {
	if err := db.AutoMigrate(&User{}, &VerificationLink{}, &file.Token{}, &file.FileMetadata{}, &file.FileVersion{}, &ActionLog{}, &PaymentHistory{}).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&VerificationLink{}).AddUniqueIndex("idx_user_sent_at", "user_id", "sent_at").Error; err != nil {
		log.Fatal(err)
	}

	// not unique index, we can have many links for same email, they could expire
	if err := db.Model(&VerificationLink{}).AddIndex("idx_vl_email", "email").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&User{}).AddUniqueIndex("idx_email", "email").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&file.Token{}).AddUniqueIndex("idx_token", "uuid").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&User{}).AddUniqueIndex("idx_payment_id", "payment_id").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&file.FileVersion{}).AddIndex("idx_token_sha", "token_id", "sha256").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&file.FileVersion{}).AddIndex("idx_fv_metadata", "file_metadata_id").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&file.FileMetadata{}).AddUniqueIndex("idx_fm_token_id_path_2", "token_id", "path", "filename").Error; err != nil {
		log.Fatal(err)
	}
}

type server struct {
	db         *gorm.DB
	store      *file.Store
	r          *gin.Engine
	authorized *gin.RouterGroup
}

func (s *server) getViewTokenLoggedOrNot(c *gin.Context) (*file.Token, *User, error) {
	token := c.Param("token")
	var t *file.Token
	var u *User
	var err error
	x, isLoggedIn := c.Get("user")
	if isLoggedIn {
		u = x.(*User)
		t, err = FindTokenForUser(s.db, token, u)
		if err != nil {
			return nil, nil, err
		}
	} else {
		t, u, err = FindToken(s.db, token)
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

func main() {
	var pbind = flag.String("bind", ":9123", "bind")
	var pdebug = flag.Bool("debug", false, "debug")
	var pinit = flag.Bool("create-tables", false, "create tables")
	var prelease = flag.Bool("release", false, "release")
	flag.Parse()

	CONFIG.MaxTokens = 100
	CONFIG.SendGridKey = os.Getenv("BAXX_SENDGRID_KEY")
	store, err := file.NewStore(os.Getenv("BAXX_S3_ENDPOINT"), os.Getenv("BAXX_S3_ACCESS_KEY"), os.Getenv("BAXX_S3_SECRET"), false)

	if err != nil {
		log.Fatal(err)
	}

	if *prelease {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()

	if *pinit {
		initDatabase(db)
	}

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		su, pass := BasicAuthDecode(c)
		if su != "" {
			u, _, err := FindUser(db, su, pass)
			if err == nil {
				c.Set("user", u)
			}
		}
	})
	authorized := r.Group("/protected")
	authorized.Use(func(c *gin.Context) {
		_, loggedIn := c.Get("user")
		if !loggedIn {
			c.Header("WWW-Authenticate", "Authorization Required")
			c.String(401, `{"error": "Not Authorized (auth required, or wrong password)"}`)
		}
	})

	srv := &server{db: db, r: r, store: store, authorized: authorized}
	setupIO(srv)
	setupACC(srv)
	setupSYNC(srv)

	log.Fatal(r.Run(*pbind))
}
