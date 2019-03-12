package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/ipn"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func getUserStatus(db *gorm.DB, u *User) (*common.UserStatusOutput, error) {
	tokens, err := u.ListTokens(db)
	if err != nil {
		return nil, err
	}
	tokensTransformed := []*common.TokenOutput{}
	for _, t := range tokens {
		tokensTransformed = append(tokensTransformed, &common.TokenOutput{ID: t.ID, UUID: t.UUID, Name: t.Name, WriteOnly: t.WriteOnly, NumberOfArchives: t.NumberOfArchives, CreatedAt: t.CreatedAt})
	}
	used := uint64(0)
	for _, t := range tokens {
		used += t.SizeUsed
	}

	vl := &VerificationLink{}
	db.Where("email = ?", u.Email).Last(vl)

	return &common.UserStatusOutput{
		EmailVerified:         u.EmailVerified,
		StartedSubscription:   u.StartedSubscription,
		CancelledSubscription: u.CancelledSubscription,
		Tokens:                tokensTransformed,
		Quota:                 u.Quota,
		LastVerificationID:    vl.ID,
		QuotaUsed:             used,
		Paid:                  u.Paid(),
		PaymentID:             u.PaymentID,
		Email:                 u.Email,
	}, nil
}

func registerUser(store *file.Store, db *gorm.DB, json common.CreateUserInput) (*common.UserStatusOutput, *User, error) {
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

	u := &User{Email: json.Email, Quota: CONFIG.DefaultQuota, QuotaInode: CONFIG.DefaultInodeQuota}
	u.SetPassword(json.Password)
	if err := tx.Create(u).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// XXX: should we throw if we fail to send verification email?
	if err := sendVerificationLink(tx, u.GenerateVerificationLink()); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	_, err = CreateTokenAndBucket(store, tx, u, false, 7, "generic-read-write-7")
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	status, err := getUserStatus(tx, u)
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

	return status, u, nil
}

func setupACC(srv *server) {
	r := srv.r
	store := srv.store
	db := srv.db
	authorized := srv.authorized
	r.POST("/register", func(c *gin.Context) {
		var json common.CreateUserInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		out, u, err := registerUser(store, db, json)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		actionLog(db, u.ID, "user", "create", c.Request)
		c.JSON(http.StatusOK, out)
	})

	authorized.POST("/status", func(c *gin.Context) {
		u := c.MustGet("user").(*User)
		status, err := getUserStatus(db, u)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, status)
	})

	authorized.POST("/replace/password", func(c *gin.Context) {
		u := c.MustGet("user").(*User)
		var json common.ChangePasswordInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := ValidatePassword(json.NewPassword); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		u.SetPassword(json.NewPassword)
		if err := db.Save(u).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &common.Success{Success: true})
	})

	authorized.POST("/replace/verification", func(c *gin.Context) {
		u := c.MustGet("user").(*User)
		verificationLink := u.GenerateVerificationLink()
		if err := db.Save(verificationLink).Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &common.Success{Success: true})
	})

	authorized.POST("/replace/email", func(c *gin.Context) {
		u := c.MustGet("user").(*User)

		var json common.ChangeEmailInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := ValidateEmail(json.NewEmail); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx := db.Begin()
		if json.NewEmail != u.Email || u.EmailVerified == nil {
			u.Email = json.NewEmail
			u.EmailVerified = nil

			verificationLink := u.GenerateVerificationLink()
			if err := sendVerificationLink(tx, verificationLink); err != nil {
				tx.Rollback()
				warnErr(c, err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		if err := tx.Save(u).Error; err != nil {
			tx.Rollback()
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := tx.Commit().Error; err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, &common.Success{Success: true})
	})

	authorized.POST("/create/token", func(c *gin.Context) {
		u := c.MustGet("user").(*User)
		var json common.CreateTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := CreateTokenAndBucket(store, db, u, json.WriteOnly, json.NumberOfArchives, json.Name)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, u.ID, "token", "create", c.Request)
		out := &common.TokenOutput{ID: token.ID, Name: token.Name, UUID: token.UUID, WriteOnly: token.WriteOnly, NumberOfArchives: token.NumberOfArchives, CreatedAt: token.CreatedAt, SizeUsed: token.SizeUsed}
		c.JSON(http.StatusOK, out)
	})

	authorized.POST("/change/token", func(c *gin.Context) {
		u := c.MustGet("user").(*User)
		var json common.ModifyTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := FindTokenForUser(db, json.UUID, u)
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

		actionLog(db, u.ID, "token", "change", c.Request)
		out := &common.TokenOutput{Name: token.Name, UUID: token.UUID, WriteOnly: token.WriteOnly, NumberOfArchives: token.NumberOfArchives, CreatedAt: token.CreatedAt, SizeUsed: token.SizeUsed}
		c.JSON(http.StatusOK, out)
	})

	authorized.POST("/delete/token", func(c *gin.Context) {
		u := c.MustGet("user").(*User)
		var json common.DeleteTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := FindTokenForUser(db, json.UUID, u)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := file.DeleteToken(store, db, token); err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, u.ID, "token", "delete", c.Request)

		c.JSON(http.StatusOK, &common.Success{Success: true})
	})
	ipn.Listener(r, "/ipn/:paymentID", func(c *gin.Context, err error, body string, n *ipn.Notification) error {
		if err != nil {
			warnErr(c, err)
			return nil
		}
		if n.TestIPN {
			warnErr(c, errors.New("testIPN received"))
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
			return nil
			// not sure what to do, just ignore
		}

		if err := db.Save(u).Error; err != nil {
			warnErr(c, err)
			return err
		}

		if cancel {
			go sendPaymentCancelMail(u.Email, u.PaymentID)
		} else if subscribe {
			go sendPaymentThanks(u.Email, u.PaymentID)
		}
		return nil
	})

	r.GET("/sub/:paymentID", func(c *gin.Context) {
		prefix := "https://www.paypal.com/cgi-bin/webscr"
		url := prefix + "?cmd=_xclick-subscriptions&business=jack%40baxx.dev&a3=5&p3=1&t3=M&item_name=baxx.dev+-+backup+as+a+service&return=https%3A%2F%2Fbaxx.dev%2Fthanks_for_paying&a1=0.1&p1=1&t1=M&src=1&sra=1&no_note=1&no_note=1&currency_code=EUR&lc=GB&notify_url=https%3A%2F%2Fbaxx.dev%2Fipn%2F" + c.Param("paymentID")
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/unsub/:paymentID", func(c *gin.Context) {
		prefix := "https://www.paypal.com/cgi-bin/webscr"
		url := prefix + "?cmd=_subscr-find&alias=2KG9SK2ZXX2A4"
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/help", func(c *gin.Context) {
		c.String(http.StatusOK, help.Render(help.EMAIL_AFTER_REGISTRATION, common.EMPTY_STATUS))
	})

	r.GET("/thanks_for_paying", func(c *gin.Context) {
		c.String(http.StatusOK, help.Render(help.EMAIL_WAIT_PAYPAL, common.EMPTY_STATUS))
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
}
