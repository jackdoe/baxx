package main

import (
	"time"

	. "github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/help"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type sendMailConfig struct {
	from        string
	to          []string
	subject     string
	body        string
	contentType string
}

func sendmail(sm sendMailConfig) error {
	if CONFIG.SendGridKey == "" {
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", sm.from)
	user := "apikey"
	pass := CONFIG.SendGridKey

	m.SetHeader("To", sm.to...)
	m.SetHeader("Bcc", "jack@sofialondonmoskva.com")
	m.SetHeader("Subject", sm.subject)
	if sm.contentType == "" {
		sm.contentType = "text/plain"
	}
	m.SetBody(sm.contentType, sm.body)

	d := gomail.NewDialer("smtp.sendgrid.net", 465, user, pass)

	return d.DialAndSend(m)
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

func sendPaymentThanks(email string, paymentid string) error {
	err := sendmail(sendMailConfig{
		from:    "jack@baxx.dev",
		to:      []string{email},
		subject: "Thanks for subscribing!",
		body:    help.Render(help.EMAIL_PAYMENT_THANKS, map[string]string{"Email": email, "PaymentID": paymentid}),
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
