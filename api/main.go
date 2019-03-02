package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	. "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
	. "github.com/jackdoe/baxx/user"
	auth "github.com/jackdoe/gin-basic-auth-dynamic"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

func warnErr(c *gin.Context, err error) {
	x, isLoggedIn := c.Get("user")
	user := &User{}
	if isLoggedIn {
		user = x.(*User)
	}
	_, fn, line, _ := runtime.Caller(1)
	log.Warnf("uid: %d, uri: %s, err: >> %s << [%s:%d]", user.ID, c.Request.RequestURI, err.Error(), fn, line)
}

func initDatabase(db *gorm.DB) {
	if err := db.AutoMigrate(&User{}, &VerificationLink{}, &Token{}, &FileOrigin{}, &FileMetadata{}, &FileVersion{}, &ActionLog{}).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&VerificationLink{}).AddUniqueIndex("idx_user_sent_at", "user_id", "sent_at").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&User{}).AddUniqueIndex("idx_email", "email").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&Token{}).AddIndex("idx_token_user_id", "user_id").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&FileOrigin{}).AddUniqueIndex("idx_sha", "sha256").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&FileMetadata{}).AddUniqueIndex("idx_fm_user_id_token_id_path", "user_id", "token_id", "path", "filename").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&FileVersion{}).AddUniqueIndex("idx_fv_metadata_origin", "file_metadata_id", "file_origin_id").Error; err != nil {
		log.Fatal(err)
	}
}

func sendVerificationLink(db *gorm.DB, verificationLink *VerificationLink) error {
	if err := db.Save(verificationLink).Error; err != nil {
		return err
	}
	err := sendmail(sendMailConfig{
		from:    "info@baxx.dev",
		to:      []string{verificationLink.Email},
		subject: "Please verify your email",
		body: fmt.Sprintf(
			`Hi,
this is the verification link: 

  https://baxx.dev/v1/verify/%s

you can check the account status with:

  curl -u %s -XPOST https://baxx.dev/protected/v1/status | json_pp

--
baxx.dev

`, verificationLink.ID, verificationLink.Email),
	})

	if err != nil {
		return err
	}
	verificationLink.SentAt = uint64(time.Now().Unix())
	if err := db.Save(verificationLink).Error; err != nil {
		return err
	}
	return nil
}

func sendRegistrationHelp(email, secret, tokenrw, tokenwo string) error {
	err := sendmail(sendMailConfig{
		from:    "info@baxx.dev",
		to:      []string{email},
		subject: "Welcome to baxx.dev!",
		body:    help.AfterRegistration(secret, tokenrw, tokenwo),
	})
	return err
}

func main() {
	var pbind = flag.String("bind", "127.0.0.1:9123", "bind")
	var proot = flag.String("root", "/tmp", "root")
	flag.Parse()

	dbType := os.Getenv("BAXX_DB")
	dbURL := os.Getenv("BAXX_DB_URL")
	debug := true
	ROOT = *proot
	if dbType == "" {
		dbType = "sqlite3"
		dbURL = "/tmp/gorm.db"
	} else {
		debug = false
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := gorm.Open(dbType, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(debug)
	defer db.Close()

	initDatabase(db)

	r := gin.Default()

	authorized := r.Group("/protected")
	authorized.Use(auth.BasicAuth(func(context *gin.Context, realm, user, pass string) auth.AuthResult {
		u, _, err := FindUser(db, user, pass)
		if err != nil {
			warnErr(context, err)
			return auth.AuthResult{Success: false, Text: `{"error":"not authorized"}`}
		}
		context.Set("user", u)

		return auth.AuthResult{Success: true}
	}))

	r.POST("/v1/register", func(c *gin.Context) {
		var json CreateUserInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := ValidatePassword(json.Password); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := ValidateEmail(json.Email); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// if the user is created again with current password just return it
		// useful for scripts i guess (probably wrong)
		u, exists, err := FindUser(db, json.Email, json.Password)
		if err == nil {
			c.JSON(http.StatusOK, &CreateUserOutput{Secret: u.SemiSecretID, TokenWO: "", TokenRW: ""})
			return
		}

		if exists {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
			return
		}

		user := &User{Email: json.Email}
		user.SetPassword(json.Password)
		if err := db.Create(user).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := sendVerificationLink(db, user.GenerateVerificationLink()); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, user.ID, "user", "create", c.Request)

		tokenWO := &Token{
			UserID:           user.ID,
			Salt:             strings.Replace(fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), "-", "", -1),
			NumberOfArchives: 7,
			WriteOnly:        true,
		}
		if err := db.Create(tokenWO).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tokenRW := &Token{
			UserID:           user.ID,
			Salt:             strings.Replace(fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), "-", "", -1),
			NumberOfArchives: 7,
			WriteOnly:        false,
		}
		if err := db.Create(tokenRW).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := sendRegistrationHelp(user.Email, user.SemiSecretID, tokenRW.ID, tokenWO.ID); err != nil {
			warnErr(c, err)
		}

		c.JSON(http.StatusOK, &CreateUserOutput{
			Secret:  user.SemiSecretID,
			TokenWO: tokenWO.ID,
			TokenRW: tokenRW.ID,
			Help:    help.AfterRegistration(user.SemiSecretID, tokenRW.ID, tokenWO.ID),
		})
	})

	authorized.POST("/v1/replace/secret", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		user.SetSemiSecretID()
		if err := db.Save(user).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &ChangeSecretOutput{
			Secret: user.SemiSecretID,
		})
	})

	authorized.POST("/v1/status", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		c.JSON(http.StatusOK, &UserStatusOutput{
			EmailVerified: user.EmailVerified,
		})
	})

	authorized.POST("/v1/replace/password", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		var json ChangePasswordInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err != ValidatePassword(json.NewPassword) {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user.SetPassword(json.NewPassword)
		if err := db.Save(user).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &Success{true})
	})

	authorized.POST("/v1/replace/verification", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		verificationLink := user.GenerateVerificationLink()
		if err := db.Save(verificationLink).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &Success{true})
	})

	authorized.POST("/v1/replace/email", func(c *gin.Context) {
		user := c.MustGet("user").(*User)

		var json ChangeEmailInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err != ValidateEmail(json.NewEmail) {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx := db.Begin()
		if json.NewEmail != user.Email || user.EmailVerified == nil {
			user.Email = json.NewEmail
			user.EmailVerified = nil

			verificationLink := user.GenerateVerificationLink()
			if err := sendVerificationLink(tx, verificationLink); err != nil {
				tx.Rollback()
				warnErr(c, err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		if err := tx.Save(user).Error; err != nil {
			tx.Rollback()
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := tx.Commit().Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, &Success{true})
	})

	authorized.POST("/v1/create/token", func(c *gin.Context) {
		user := c.MustGet("user").(*User)

		var json CreateTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token := &Token{
			UserID:           user.ID,
			Salt:             strings.Replace(fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), "-", "", -1),
			NumberOfArchives: json.NumberOfArchives,
			WriteOnly:        json.WriteOnly,
		}

		if err := db.Create(token).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, user.ID, "token", "create", c.Request)
		c.JSON(http.StatusOK, token)
	})

	getViewTokenLoggedOrNot := func(c *gin.Context) (*Token, error) {
		user := c.Param("user_semi_secret_id")
		x, isLoggedIn := c.Get("user")
		if isLoggedIn {
			if user != x.(*User).SemiSecretID {
				return nil, errors.New("wrong token/user combination")
			}
		}

		token := c.Param("token")

		t, err := FindToken(db, user, token)
		if err != nil {
			return nil, err
		}

		if !isLoggedIn {
			if t.WriteOnly && c.Request.Method != "POST" {
				return nil, errors.New("write only token, use /v1/protected/io/:secret/:token/*path")
			}
		}
		return t, nil
	}

	download := func(c *gin.Context) {
		t, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fo, file, reader, err := FindAndOpenFile(db, t, c.Param("path"))
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer file.Close()
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		c.DataFromReader(http.StatusOK, int64(fo.Size), "octet/stream", reader, map[string]string{})
	}

	upload := func(c *gin.Context) {
		t, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		body := c.Request.Body
		defer body.Close()
		fv, err := SaveFile(db, t, body, c.Param("path"))
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, t.UserID, "file", "upload", c.Request, fmt.Sprintf("FileVersion: %d/%d/%d", fv.ID, fv.FileMetadataID, fv.FileOriginID))
		c.JSON(http.StatusOK, fv)
	}

	mutatePATH := "/v1/io/:user_semi_secret_id/:token/*path"

	authorized.GET(mutatePATH, download)
	r.GET(mutatePATH, download)

	authorized.POST(mutatePATH, upload)
	r.POST(mutatePATH, upload)

	authorized.GET("/v1/subscription", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		c.String(http.StatusOK, "https://www.paypal.com/cgi-bin/webscr?cmd=_xclick-subscriptions&business=jack%40sofialondonmoskva.com&a3=5&p3=1&t3=M&item_name=baxx.dev%20backups%20as%20a%20service&item_number=1&return=https%3A%2F%2Fbax.dev%2Fthanks&p1=1&t1=M&src=1&sra=1&no_note=1&no_note=1&currency_code=EUR&lc=US&ify_url=https%3A%2F%2Fbaxx.dev%2Fipn%2F"+
			user.Seed)
	})

	r.GET("/v1/verify/:id", func(c *gin.Context) {
		v := &VerificationLink{}
		now := time.Now()
		v.VerifiedAt = &now
		wrong := func(err error) {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Oops, something went wrong!\n%s\n\n, if persists please send it to help@baxx.dev\n", err.Error()))
		}
		tx := db.Begin()

		query := tx.Where("id = ?", c.Param("id")).Take(v)
		if query.RecordNotFound() {
			tx.Rollback()
			warnErr(c, errors.New("verification link not found"))
			c.String(http.StatusNotFound, "Oops, verification link not found!\n")
			return
		}
		if err := tx.Save(v).Error; err != nil {
			tx.Rollback()
			warnErr(c, err)
			wrong(err)
			return
		}

		if time.Now().Unix()-int64(v.SentAt) > (24 * 3600) {
			tx.Rollback()
			warnErr(c, errors.New(fmt.Sprintf("verification link expired %#v", v)))
			c.String(http.StatusOK, `Oops, verification link has expired!

You can generate new one with:

 curl -u your.current.email@example.com \
  -XPOST -d'{"new_email": "your.current.email@example.com"}' \
  https://baxx.dev/protected/v1/replace/email

The verification links are valid for 24 hours,
You can check your account status at:

  curl -u your.current.email@example.com -XPOST https://baxx.dev/protected/v1/status

If something is wrong, please contact me at help@baxx.dev.

Thanks!
`)
			return

		}

		u := &User{}
		if err := tx.Where("id = ?", v.UserID).Take(u).Error; err != nil {
			warnErr(c, errors.New(fmt.Sprintf("weird state, verification's user not found %#v", v)))
			tx.Rollback()
			c.String(http.StatusOK, `
Oops, verification link's user not found,
this is very weird

please contact me at help@baxx.dev!
`)

			return
		}

		if u.Email != v.Email {
			tx.Rollback()
			warnErr(c, errors.New(fmt.Sprintf("weird state, user changed email %#v %#v", v, u)))
			c.String(http.StatusOK,
				`Oops, the user's email already changed,
this is the old verification link.

If you don't receive new link please contact me at help@baxx.dev!
`)
			return
		}

		u.EmailVerified = v.VerifiedAt
		if err := tx.Save(v).Error; err != nil {
			tx.Rollback()
			warnErr(c, err)
			wrong(err)
			return
		}
		if err := tx.Commit().Error; err != nil {
			warnErr(c, err)
			wrong(err)
		}

		c.String(http.StatusOK, `Thanks!
The email is verified now!

You can check your account status at:

  curl -u your.current.email@example.com -XPOST https://baxx.dev/protected/v1/status

`)
	})

	r.Run(*pbind)
}
