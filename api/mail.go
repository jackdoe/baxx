package main

import (
	"time"

	. "github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func sendVerificationLink(db *gorm.DB, verificationLink *user.VerificationLink) error {
	if err := db.Save(verificationLink).Error; err != nil {
		return err
	}
	err := Sendmail(CONFIG.SendGridKey, Message{
		From:    "jack@baxx.dev",
		To:      []string{verificationLink.Email},
		Subject: "Please verify your email",
		Body:    help.Render(help.EMAIL_VALIDATION, verificationLink),
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
	err := Sendmail(CONFIG.SendGridKey, Message{
		From:    "jack@baxx.dev",
		To:      []string{email},
		Subject: "Thanks for subscribing!",
		Body:    help.Render(help.EMAIL_PAYMENT_THANKS, map[string]string{"Email": email, "PaymentID": paymentid}),
	})
	if err != nil {
		log.Warnf("failed to send: %s", err.Error())
	}
	return err
}

func sendPaymentCancelMail(email string, paymentID string) error {
	err := Sendmail(CONFIG.SendGridKey, Message{
		From:    "jack@baxx.dev",
		To:      []string{email},
		Subject: "Subscription cancelled!",
		Body:    help.Render(help.EMAIL_PAYMENT_CANCEL, map[string]string{"PaymentID": paymentID, "Email": email}),
	})

	if err != nil {
		log.Warnf("failed to send: %s", err.Error())
	}
	return err
}

func sendRegistrationHelp(status *UserStatusOutput) error {
	err := Sendmail(CONFIG.SendGridKey, Message{
		From:    "jack@baxx.dev",
		To:      []string{status.Email},
		Subject: "Welcome to baxx.dev!",
		Body:    help.Render(help.EMAIL_AFTER_REGISTRATION, status),
	})
	return err
}
