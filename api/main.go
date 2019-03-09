package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/erikdubbelboer/gspt"
	"github.com/gin-gonic/gin"
	. "github.com/jackdoe/baxx/baxxio"
	. "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/config"
	. "github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/ipn"
	. "github.com/jackdoe/baxx/user"
	auth "github.com/jackdoe/gin-basic-auth-dynamic"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"io"
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
	if err := db.AutoMigrate(&User{}, &VerificationLink{}, &Token{}, &FileMetadata{}, &FileVersion{}, &ActionLog{}, &PaymentHistory{}).Error; err != nil {
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

	if err := db.Model(&Token{}).AddUniqueIndex("idx_token", "uuid").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&User{}).AddUniqueIndex("idx_payment_id", "payment_id").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&FileVersion{}).AddIndex("idx_token_sha", "token_id", "sha256").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&FileMetadata{}).AddUniqueIndex("idx_fm_token_id_path_2", "token_id", "path", "filename").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&FileVersion{}).AddIndex("idx_fv_metadata_updated", "file_metadata_id", "updated_at_ns").Error; err != nil {
		log.Fatal(err)
	}

}

func getUserStatus(db *gorm.DB, user *User) (*UserStatusOutput, error) {
	tokens, err := user.ListTokens(db)
	if err != nil {
		return nil, err
	}
	tokensTransformed := []*TokenOutput{}
	for _, t := range tokens {
		tokensTransformed = append(tokensTransformed, &TokenOutput{UUID: t.UUID, Name: t.Name, WriteOnly: t.WriteOnly, NumberOfArchives: t.NumberOfArchives, CreatedAt: t.CreatedAt})
	}
	used := uint64(0)
	for _, t := range tokens {
		used += t.SizeUsed
	}

	vl := &VerificationLink{}
	db.Where("email = ?", user.Email).Last(vl)

	return &UserStatusOutput{
		EmailVerified:         user.EmailVerified,
		StartedSubscription:   user.StartedSubscription,
		CancelledSubscription: user.CancelledSubscription,
		Tokens:                tokensTransformed,
		Quota:                 user.Quota,
		LastVerificationID:    vl.ID,
		QuotaUsed:             used,
		Paid:                  user.Paid(),
		PaymentID:             user.PaymentID,
		Email:                 user.Email,
	}, nil
}

func sendVerificationLink(db *gorm.DB, verificationLink *VerificationLink) error {
	if err := db.Save(verificationLink).Error; err != nil {
		return err
	}
	err := sendmail(sendMailConfig{
		from:    "jack@baxx.dev",
		to:      []string{verificationLink.Email},
		subject: "Please verify your email",
		body:    help.Render(help.EMAIL_VALIDATION, verificationLink),
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

func sendPaymentThanks(email string) error {
	err := sendmail(sendMailConfig{
		from:    "jack@baxx.dev",
		to:      []string{email},
		subject: "Thanks for subscribing!",
		body:    help.Render(help.EMAIL_PAYMENT_THANKS, map[string]string{"Email": email}),
	})
	if err != nil {
		log.Warnf("failed to send: %s", err.Error())
	}
	return err
}

func sendPaymentCancelMail(email string, paymentID string) error {
	err := sendmail(sendMailConfig{
		from:    "jack@baxx.dev",
		to:      []string{email},
		subject: "Subscription cancelled!",
		body:    help.Render(help.EMAIL_PAYMENT_CANCEL, map[string]string{"PaymentID": paymentID, "Email": email}),
	})

	if err != nil {
		log.Warnf("failed to send: %s", err.Error())
	}
	return err
}

func sendRegistrationHelp(status *UserStatusOutput) error {
	err := sendmail(sendMailConfig{
		from:    "jack@baxx.dev",
		to:      []string{status.Email},
		subject: "Welcome to baxx.dev!",
		body:    help.Render(help.EMAIL_AFTER_REGISTRATION, status),
	})
	return err
}

func registerUser(db *gorm.DB, json CreateUserInput) (*UserStatusOutput, *User, error) {
	if err := ValidatePassword(json.Password); err != nil {
		return nil, nil, err
	}

	if err := ValidateEmail(json.Email); err != nil {
		return nil, nil, err
	}
	tx := db.Begin()
	_, exists, err := FindUser(tx, json.Email, json.Password)
	if err == nil || exists {
		tx.Rollback()
		return nil, nil, errors.New("user already exists")
	}

	user := &User{Email: json.Email, Quota: CONFIG.DefaultQuota, QuotaInode: CONFIG.DefaultInodeQuota}
	user.SetPassword(json.Password)
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// XXX: should we throw if we fail to send verification email?
	if err := sendVerificationLink(tx, user.GenerateVerificationLink()); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	_, err = user.CreateToken(tx, true, 7, "generic-write-only-7")
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	_, err = user.CreateToken(tx, false, 7, "generic-read-write-7")
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	status, err := getUserStatus(tx, user)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := sendRegistrationHelp(status); err != nil {
		log.Warnf("failed to send email, ignoring error and moving on,  %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	return status, user, nil
}

func main() {
	var pbind = flag.String("bind", "127.0.0.1:9123", "bind")
	var proot = flag.String("root", "/tmp", "temporary file root")
	var ps3region = flag.String("s3-region", "ams3", "s3 region")
	var ps3endpoint = flag.String("s3-endpoint", "ams3.digitaloceanspaces.com", "s3 endpoint")
	var ps3bucket = flag.String("s3-bucket", "baxx", "s3 bucket")
	var ps3keyid = flag.String("s3-key-id", "", "s3 key id")
	var ps3secret = flag.String("s3-secret", "", "s3 secret")
	var ps3token = flag.String("s3-token", "", "s3 token")
	var pdbtype = flag.String("db-type", "sqlite3", "database type, sqlite3 or mysql")
	var pdburl = flag.String("db-url", "/tmp/gorm.db", "database url")
	var psendgridkey = flag.String("sendgrid", "", "sendgrid api key")
	var pdebug = flag.Bool("debug", false, "debug")
	var psandbox = flag.Bool("sandbox", false, "sandbox")
	var prelease = flag.Bool("release", false, "release")
	flag.Parse()

	CONFIG.MaxTokens = 100
	CONFIG.SendGridKey = *psendgridkey
	CONFIG.TemporaryRoot = *proot
	store := NewStore(&StoreConfig{
		Endpoint:        *ps3endpoint,
		Region:          *ps3region,
		Bucket:          *ps3bucket,
		AccessKeyID:     *ps3keyid,
		SecretAccessKey: *ps3secret,
		SessionToken:    *ps3token,
	})
	title := []string{}
	if *prelease {
		gin.SetMode(gin.ReleaseMode)
		title = append(title, "release")
	} else {
		title = append(title, "dev")
	}

	if *pdebug {
		title = append(title, "debug")
	}

	if *psandbox {
		title = append(title, "sandbox")
	}
	title = append(title, *pdbtype)
	title = append(title, *proot)
	title = append(title, *ps3region)
	title = append(title, *ps3bucket)
	title = append(title, *ps3endpoint)

	// because of the passwords in the parameters
	gspt.SetProcTitle(fmt.Sprintf("[baxx.dev %s]", strings.Join(title, " ")))

	db, err := gorm.Open(*pdbtype, *pdburl)
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(*pdebug)
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

	r.POST("/register", func(c *gin.Context) {
		var json CreateUserInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		out, user, err := registerUser(db, json)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		actionLog(db, user.ID, "user", "create", c.Request)
		c.JSON(http.StatusOK, out)
	})

	authorized.POST("/status", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		status, err := getUserStatus(db, user)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, status)
	})

	authorized.POST("/replace/password", func(c *gin.Context) {
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

		c.JSON(http.StatusOK, &Success{Success: true})
	})

	authorized.POST("/replace/verification", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		verificationLink := user.GenerateVerificationLink()
		if err := db.Save(verificationLink).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &Success{Success: true})
	})

	authorized.POST("/replace/email", func(c *gin.Context) {
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

		c.JSON(http.StatusOK, &Success{Success: true})
	})

	authorized.POST("/create/token", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		var json CreateTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := user.CreateToken(db, json.WriteOnly, json.NumberOfArchives, json.Name)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, user.ID, "token", "create", c.Request)
		out := &TokenOutput{Name: token.Name, UUID: token.UUID, WriteOnly: token.WriteOnly, NumberOfArchives: token.NumberOfArchives, CreatedAt: token.CreatedAt, SizeUsed: token.SizeUsed}
		c.JSON(http.StatusOK, out)
	})

	authorized.POST("/change/token", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		var json ModifyTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := FindTokenForUser(db, json.UUID, user)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if json.NumberOfArchives > 0 {
			token.NumberOfArchives = json.NumberOfArchives
		}
		if json.WriteOnly != nil {
			token.WriteOnly = *json.WriteOnly
		}
		token.Name = json.Name
		if err := db.Save(token).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, user.ID, "token", "change", c.Request)
		out := &TokenOutput{Name: token.Name, UUID: token.UUID, WriteOnly: token.WriteOnly, NumberOfArchives: token.NumberOfArchives, CreatedAt: token.CreatedAt, SizeUsed: token.SizeUsed}
		c.JSON(http.StatusOK, out)
	})

	authorized.POST("/delete/token", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		var json DeleteTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := FindTokenForUser(db, json.UUID, user)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := DeleteToken(store, db, token); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, user.ID, "token", "delete", c.Request)

		c.JSON(http.StatusOK, &Success{Success: true})
	})

	getViewTokenLoggedOrNot := func(c *gin.Context) (*Token, *User, error) {
		token := c.Param("token")
		var t *Token
		var u *User
		x, isLoggedIn := c.Get("user")
		if isLoggedIn {
			u = x.(*User)
			t, err = FindTokenForUser(db, token, u)
			if err != nil {
				return nil, nil, err
			}
		} else {
			t, u, err = FindToken(db, token)
			if err != nil {
				return nil, nil, err
			}
			if t.WriteOnly && c.Request.Method != "POST" {
				return nil, nil, errors.New("write only token, use /protected/io/$TOKEN/*path")
			}
		}

		if u.EmailVerified == nil {
			err = errors.New("email not verified yet")
			return nil, nil, err
		}

		if !u.Paid() {
			err = fmt.Errorf("payment not received yet or subscription is cancelled, go to https://baxx.dev/sub/%s or if you already did, contact me at jack@baxx.dev", u.PaymentID)
			return nil, nil, err
		}

		return t, u, nil
	}

	download := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		localFile, err := CreateTemporaryFile(t.ID, c.Param("path"))
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		defer func() {
			localFile.File.Close()
			os.Remove(localFile.TempName)
		}()

		fo, reader, err := FindAndDecodeFile(store, db, t, localFile)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Header("Content-Length", fmt.Sprintf("%d", fo.Size))
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+fo.SHA256+".sha") // make sure people dont use it for loading js
		c.Header("Content-Type", "application/octet-stream")
		c.DataFromReader(http.StatusOK, int64(fo.Size), "octet/stream", reader, map[string]string{})
	}

	lookupSHA := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fv, fm, err := FindFileBySHA(db, t, c.Param("sha256"))
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		accepted := c.NegotiateFormat("application/json")
		if accepted == "application/json" {
			c.JSON(http.StatusOK, gin.H{"sha": fv.SHA256, "path": fm.Path, "name": fm.Filename})
			return
		}

		c.String(http.StatusOK, FileLine(fm, fv))
	}

	upload := func(c *gin.Context) {
		body := c.Request.Body
		defer body.Close()

		t, user, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p := c.Param("path")
		fv, fm, err := SaveFileProcess(store, db, user, t, body, p)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// check if over quota

		actionLog(db, t.UserID, "file", "upload", c.Request, fmt.Sprintf("FileVersion: %d/%d", fv.ID, fv.FileMetadataID))
		accepted := c.NegotiateFormat("application/json")
		if accepted == "application/json" {
			c.JSON(http.StatusOK, fv)
			return
		}
		c.String(http.StatusOK, FileLine(fm, fv))
	}

	deleteFile := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := DeleteFile(store, db, t, c.Param("path")); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, t.UserID, "file", "delete", c.Request, "")
		c.JSON(http.StatusOK, &Success{Success: true})
	}

	listFiles := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		p := c.Param("path")
		if !strings.HasSuffix(p, "/") {
			p = p + "/"
		}

		files, err := ListFilesInPath(db, t, p)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		accepted := c.NegotiateFormat("application/json")
		if accepted == "application/json" {
			c.JSON(http.StatusOK, files)
			return
		}
		c.String(http.StatusOK, LSAL(files))
	}

	mutateSinglePATH := "/io/:token/*path"

	authorized.GET(mutateSinglePATH, download)
	r.GET(mutateSinglePATH, download)

	authorized.POST(mutateSinglePATH, upload)
	r.POST(mutateSinglePATH, upload)

	authorized.DELETE(mutateSinglePATH, deleteFile)
	r.DELETE(mutateSinglePATH, deleteFile)

	for _, a := range []string{"dir", "ls"} {
		mutateManyPATH := "/" + a + "/:token/*path"
		authorized.GET(mutateManyPATH, listFiles)
		r.GET(mutateManyPATH, listFiles)
	}

	mutateManyPATH := "/sha256/:token/:sha256"
	authorized.GET(mutateManyPATH, lookupSHA)
	r.GET(mutateManyPATH, lookupSHA)

	ipn.Listener(r, "/ipn/:paymentID", func(c *gin.Context, err error, body string, n *ipn.Notification) error {
		if err != nil {
			warnErr(c, err)
			return nil
		}
		if n.TestIPN && !*psandbox {
			// received testipn request while not in sandbox mode
			warnErr(c, errors.New("testIPN received while not in sandbox mode"))
		}

		// check currency and amount and etc
		// otherwise anyone can create ipn request with wrong amount :D

		u := &User{}
		if err := db.Where("payment_id = ?", c.Param("paymentID")).Take(u).Error; err != nil {
			return err
		}
		encoded, err := n.JSON()
		if err != nil {
			warnErr(c, err)
			encoded = "{}"
		}

		ph := &PaymentHistory{
			UserID: u.ID,
			IPN:    encoded,
			IPNRAW: body,
		}

		if err := db.Create(ph).Error; err != nil {
			warnErr(c, err)
			return err
		}

		now := time.Now()
		cancel := false
		subscribe := false
		if n.TxnType == "subscr_signup" {
			u.StartedSubscription = &now
			u.CancelledSubscription = nil
			subscribe = true
		} else if n.TxnType == "subscr_cancel" {
			u.CancelledSubscription = &now
			cancel = true
		} else {
			log.Warnf("unknown txn type, ignoring: %s", n.TxnType)
			// not sure what to do, just ignore
		}

		if err := db.Save(u).Error; err != nil {
			warnErr(c, err)
			return err
		}

		if cancel {
			go sendPaymentCancelMail(u.Email, u.PaymentID)
		} else if subscribe {
			go sendPaymentThanks(u.Email)
		}
		return nil
	})

	r.GET("/sub/:paymentID", func(c *gin.Context) {
		prefix := "https://www.paypal.com/cgi-bin/webscr"
		if *psandbox {
			prefix = "https://ipnpb.sandbox.paypal.com/cgi-bin/webscr"
		}
		url := prefix + "?cmd=_xclick-subscriptions&business=jack%40baxx.dev&a3=5&p3=1&t3=M&item_name=baxx.dev+-+backup+as+a+service&return=https%3A%2F%2Fbaxx.dev%2Fthanks_for_paying&a1=0.1&p1=1&t1=M&src=1&sra=1&no_note=1&no_note=1&currency_code=EUR&lc=GB&charset=UTF%2d8¬notify_url=https%3A%2F%2Fbaxx.dev%2Fipn%2F" + c.Param("paymentID")
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/help", func(c *gin.Context) {
		c.String(http.StatusOK, help.Render(help.EMAIL_AFTER_REGISTRATION, EMPTY_STATUS))
	})

	r.GET("/thanks_for_paying", func(c *gin.Context) {
		c.String(http.StatusOK, help.Render(help.EMAIL_WAIT_PAYPAL, EMPTY_STATUS))
	})

	r.GET("/verify/:id", func(c *gin.Context) {
		v := &VerificationLink{}
		now := time.Now()

		wrong := func(err error) {
			c.String(http.StatusInternalServerError, help.Render(help.HTML_LINK_ERROR, err.Error()))
		}
		tx := db.Begin()

		query := tx.Where("id = ?", c.Param("id")).Take(v)
		if query.RecordNotFound() {
			tx.Rollback()
			warnErr(c, errors.New("verification link not found"))
			c.String(http.StatusNotFound, "Oops, verification link not found!\n")
			return
		}
		v.VerifiedAt = &now
		if err := tx.Save(v).Error; err != nil {
			tx.Rollback()
			warnErr(c, err)
			wrong(err)
			return
		}

		if time.Now().Unix()-int64(v.SentAt) > (24 * 3600) {
			tx.Rollback()
			warnErr(c, fmt.Errorf("verification link expired %#v", v))
			c.String(http.StatusOK, help.Render(help.HTML_LINK_EXPIRED, v))
			return
		}

		u := &User{}
		if err := tx.Where("id = ?", v.UserID).Take(u).Error; err != nil {
			warnErr(c, fmt.Errorf("weird state, verification's user not found %#v", v))
			tx.Rollback()
			c.String(http.StatusInternalServerError, help.Render(help.HTML_LINK_ERROR, "verification link's user not found, this is very weird"))
			return
		}

		if u.Email != v.Email {
			tx.Rollback()
			warnErr(c, fmt.Errorf("weird state, user changed email %#v %#v", v, u))
			c.String(http.StatusInternalServerError, help.Render(help.HTML_LINK_ERROR, "user email already changed, this is very weird"))
			return
		}

		u.EmailVerified = v.VerifiedAt
		if err := tx.Save(u).Error; err != nil {
			tx.Rollback()
			warnErr(c, err)
			wrong(err)
			return
		}
		if err := tx.Commit().Error; err != nil {
			warnErr(c, err)
			wrong(err)
		}

		c.String(http.StatusOK, help.Render(help.HTML_VERIFICATION_OK, v))
	})

	log.Fatal(r.Run(*pbind))
}

func SaveFileProcess(s *Store, db *gorm.DB, user *User, t *Token, body io.Reader, p string) (*FileVersion, *FileMetadata, error) {
	localFile, err := CreateTemporaryFile(t.ID, p)
	if err != nil {
		return nil, nil, err
	}
	defer os.Remove(localFile.TempName)
	defer localFile.File.Close()

	sha, size, err := SaveUploadedFile(t.Salt, localFile.File, body)
	if err != nil {
		return nil, nil, err
	}

	leftSize, leftInodes, err := user.GetQuotaLeft(db)
	if err != nil {
		return nil, nil, err
	}
	if leftSize < size {
		return nil, nil, errors.New("quota limit reached")
	}

	if leftInodes < 1 {
		return nil, nil, errors.New("inode quota limit reached")
	}

	localFile.SHA = sha
	localFile.Size = uint64(size)
	_, err = localFile.File.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, err
	}

	fv, fm, err := SaveFile(s, db, t, localFile)

	return fv, fm, err
}
