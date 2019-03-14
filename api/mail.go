package main

import (
	. "github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/help"
	log "github.com/sirupsen/logrus"
)

func sendVerificationLink(status *UserStatusOutput) error {
	err := Sendmail(CONFIG.SendGridKey, Message{
		From:    "jack@baxx.dev",
		To:      []string{status.Email},
		Subject: "Please verify your email",
		Body:    help.Render(help.HelpObject{Template: help.EmailValidation, Email: status.Email, Status: status}),
	})
	if err != nil {
		return err
	}
	// FIXME: sentAt, use queue
	return nil
}

func sendPaymentThanksMail(status *UserStatusOutput) error {
	err := Sendmail(CONFIG.SendGridKey, Message{
		From:    "jack@baxx.dev",
		To:      []string{status.Email},
		Subject: "Thanks for subscribing!",
		Body:    help.Render(help.HelpObject{Template: help.EmailPaymentThanks, Email: status.Email, Status: status}),
	})
	if err != nil {
		log.Warnf("failed to send: %s", err.Error())
	}
	return err
}

func sendPaymentCancelMail(status *UserStatusOutput) error {
	err := Sendmail(CONFIG.SendGridKey, Message{
		From:    "jack@baxx.dev",
		To:      []string{status.Email},
		Subject: "Subscription cancelled!",
		Body:    help.Render(help.HelpObject{Template: help.EmailPaymentCancel, Email: status.Email, Status: status}),
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
		Body:    help.Render(help.HelpObject{Template: help.EmailAfterRegistration, Email: status.Email, Status: status}),
	})
	return err
}
