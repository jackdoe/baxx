package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackdoe/baxx/api/helpers"
	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/ipn"
	"github.com/jackdoe/baxx/notification"
	"github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func registerUser(store *file.Store, db *gorm.DB, json common.CreateUserInput) (*common.UserStatusOutput, *user.User, error) {
	if err := ValidatePassword(json.Password); err != nil {
		return nil, nil, err
	}

	if err := ValidateEmail(json.Email); err != nil {
		return nil, nil, err
	}
	tx := db.Begin()

	if user.Exists(tx, json.Email) {
		// user already exists
		tx.Rollback()
		return nil, nil, errors.New("user already exists")
	}

	u := &user.User{Email: json.Email}
	u.SetPassword(json.Password)
	if err := tx.Create(u).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// XXX: should we throw if we fail to send verification email?
	verificationLink := u.GenerateVerificationLink()
	if err := tx.Save(verificationLink).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	status, err := helpers.GetUserStatus(tx, u)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := sendVerificationLink(tx, status); err != nil {
		tx.Rollback()
		return nil, nil, err
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
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		out, _, err := registerUser(store, db, json)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, out)
	})

	statusFn := func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)
		status, err := helpers.GetUserStatus(db, u)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, status)
	}

	authorized.POST("/status", statusFn)
	authorized.GET("/status", statusFn)

	authorized.POST("/replace/password", func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)
		var json common.ChangePasswordInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := ValidatePassword(json.NewPassword); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		u.SetPassword(json.NewPassword)
		if err := db.Save(u).Error; err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, &common.Success{Success: true})
	})

	authorized.POST("/replace/verification", func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)
		verificationLink := u.GenerateVerificationLink()
		if err := db.Save(verificationLink).Error; err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, &common.Success{Success: true})
	})

	authorized.POST("/replace/email", func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)

		var json common.ChangeEmailInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := ValidateEmail(json.NewEmail); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx := db.Begin()
		var verificationLink *user.VerificationLink
		if json.NewEmail != u.Email || u.EmailVerified == nil {
			u.Email = json.NewEmail
			u.EmailVerified = nil

			verificationLink = u.GenerateVerificationLink()
			if err := tx.Save(verificationLink).Error; err != nil {
				tx.Rollback()
				warnErr(c, err)
				c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

		}

		if err := tx.Save(u).Error; err != nil {
			tx.Rollback()
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		status, err := helpers.GetUserStatus(tx, u)
		if err != nil {
			tx.Rollback()
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if verificationLink != nil {
			if err := sendVerificationLink(tx, status); err != nil {
				tx.Rollback()
				warnErr(c, err)
				c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.IndentedJSON(http.StatusOK, &common.Success{Success: true})
	})

	authorized.POST("/create/notification", func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)
		var json *common.CreateNotificationInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		n, err := helpers.CreateNotificationRule(db, u, json)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, notification.TransformRuleToOutput(n))
	})

	authorized.POST("/delete/notification", func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)
		var json *common.DeleteNotificationInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		n := &notification.NotificationRule{}
		if err := db.Where("uuid = ? AND user_id = ?", json.UUID, u.ID).Error; err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Delete(n).Error; err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		}
		c.IndentedJSON(http.StatusOK, &common.Success{Success: true})
	})

	authorized.POST("/change/notification", func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)
		var json *common.ModifyNotificationInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		n, err := helpers.ChangeNotificationRule(db, u, json)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, notification.TransformRuleToOutput(n))
	})

	authorized.POST("/create/token", func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)
		var json common.CreateTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := helpers.CreateTokenAndNotification(store, db, u, CONFIG.Bucket, json.WriteOnly, json.NumberOfArchives, json.Name, CONFIG.DefaultQuota, CONFIG.DefaultInodeQuota, CONFIG.MaxTokens)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		rules, err := helpers.ListNotifications(db, token)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		}

		c.IndentedJSON(http.StatusOK, helpers.TransformTokenForSending(token, 0, 0, rules))
	})

	authorized.POST("/change/token", func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)
		var json common.ModifyTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := helpers.FindTokenForUser(db, json.UUID, u)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rules, err := helpers.ListNotifications(db, token)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, helpers.TransformTokenForSending(token, 0, 0, rules))
	})

	authorized.POST("/delete/token", func(c *gin.Context) {
		u := c.MustGet("user").(*user.User)
		var json common.DeleteTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := helpers.FindTokenForUser(db, json.UUID, u)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := file.DeleteToken(store, db, token); err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, &common.Success{Success: true})
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
		tx := db.Begin()
		u := &user.User{}
		if err := tx.Where("payment_id = ?", c.Param("paymentID")).Take(u).Error; err != nil {
			tx.Rollback()
			return err
		}
		encoded, err := n.JSON()
		if err != nil {
			warnErr(c, err)
			encoded = "{}"
		}

		ph := &user.PaymentHistory{
			UserID: u.ID,
			IPN:    encoded,
			IPNRAW: body,
		}

		if err := tx.Create(ph).Error; err != nil {
			tx.Rollback()
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
			tx.Rollback()
			return nil
			// not sure what to do, just ignore
		}

		if err := tx.Save(u).Error; err != nil {
			tx.Rollback()
			warnErr(c, err)
			return err
		}
		status, err := helpers.GetUserStatus(tx, u)
		if err != nil {
			tx.Rollback()
			warnErr(c, err)
			return err
		}

		if cancel {
			err := sendPaymentCancelMail(tx, status)
			if err != nil {
				tx.Rollback()
				warnErr(c, err)
				return err

			}
		} else if subscribe {
			if len(status.Tokens) == 0 {
				// in case someone re-subscribes, dont make new token for them
				_, err := helpers.CreateTokenAndNotification(store, tx, u, CONFIG.Bucket, false, 7, "generic-read-write-7", CONFIG.DefaultQuota, CONFIG.DefaultInodeQuota, CONFIG.MaxTokens)
				if err != nil {
					tx.Rollback()
					warnErr(c, err)
					return err
				}
			}

			status, err = helpers.GetUserStatus(tx, u)
			if err != nil {
				tx.Rollback()
				warnErr(c, err)
				return err
			}

			if err = sendRegistrationHelp(tx, status); err != nil {
				tx.Rollback()
				warnErr(c, err)
				return err
			}
		}
		if err := tx.Commit().Error; err != nil {
			warnErr(c, err)
			return err
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
		c.String(http.StatusOK, help.Render(help.HelpObject{Template: help.AllHelp, Email: common.EMPTY_STATUS.Email, Status: common.EMPTY_STATUS}))
	})

	r.GET("/thanks_for_paying", func(c *gin.Context) {
		c.String(http.StatusOK, help.Render(help.HelpObject{Template: help.HtmlWaitPaypal, Email: common.EMPTY_STATUS.Email}))
	})

	r.GET("/verify/:id", func(c *gin.Context) {
		v := &user.VerificationLink{}
		now := time.Now()

		wrong := func(err error) {
			c.String(http.StatusInternalServerError, help.Render(help.HelpObject{Template: help.HtmlLinkError, Err: err}))
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
			c.String(http.StatusOK, help.Render(help.HelpObject{Template: help.HtmlLinkExpired}))
			return
		}

		u := &user.User{}
		if err := tx.Where("id = ?", v.UserID).Take(u).Error; err != nil {
			warnErr(c, fmt.Errorf("weird state, verification's user not found %#v", v))
			tx.Rollback()
			wrong(fmt.Errorf("verification link's user not found, this is very weird"))
			return
		}

		if u.Email != v.Email {
			tx.Rollback()
			warnErr(c, fmt.Errorf("weird state, user changed email %#v %#v", v, u))
			wrong(fmt.Errorf("user email already changed, this is very weird"))
			return
		}

		u.EmailVerified = v.VerifiedAt
		if err := tx.Save(u).Error; err != nil {
			tx.Rollback()
			warnErr(c, err)
			wrong(err)
			return
		}

		if !u.Paid() {
			status, err := helpers.GetUserStatus(tx, u)
			if err != nil {
				tx.Rollback()
				warnErr(c, err)
				wrong(err)
			}

			if err := sendPaymentPlease(tx, status); err != nil {
				tx.Rollback()
				warnErr(c, err)
				wrong(err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			warnErr(c, err)
			wrong(err)
		}

		c.String(http.StatusOK, help.Render(help.HelpObject{
			Template: help.HtmlVerificationOk,
		}))
	})

	srv.registerHelp(false, help.HelpObject{Template: help.Profile, Status: common.EMPTY_STATUS}, "/register")
	srv.registerHelp(false, help.HelpObject{Template: help.GuiTos, Status: common.EMPTY_STATUS}, "/register/tos")
	srv.registerHelp(false, help.HelpObject{Template: help.TokenMeta, Status: common.EMPTY_STATUS}, "/protected/token", "/protected/token/*path", "/token", "/tokens")
	srv.registerHelp(false, help.HelpObject{Template: help.NotificationMeta, Status: common.EMPTY_STATUS}, "/protected/notification", "/protected/notification/*path", "/notification", "/notifications")
}
