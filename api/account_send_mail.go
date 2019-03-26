package main

import (
	. "github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/notification"
	"github.com/jinzhu/gorm"
)

func sendPaymentPlease(db *gorm.DB, status *UserStatusOutput) error {
	err := notification.EnqueueVerificationMail(
		db,
		status.UserID,
		"Subscription info",
		help.Render(help.HelpObject{Template: help.EmailPaymentPlease, Email: status.Email, Status: status}))
	return err
}

func sendVerificationLink(db *gorm.DB, status *UserStatusOutput) error {
	err := notification.EnqueueVerificationMail(
		db,
		status.UserID,
		"Please verify your email",
		help.Render(help.HelpObject{Template: help.EmailValidation, Email: status.Email, Status: status}))
	return err
}

func sendPaymentCancelMail(db *gorm.DB, status *UserStatusOutput) error {
	err := notification.EnqueueMail(
		db,
		status.UserID,
		"Subscription cancelled!",
		help.Render(help.HelpObject{Template: help.EmailPaymentCancel, Email: status.Email, Status: status}))
	return err
}

func sendRegistrationHelp(db *gorm.DB, status *UserStatusOutput) error {
	err := notification.EnqueueMail(
		db,
		status.UserID,
		"Welcome to baxx.dev!",
		help.Render(help.HelpObject{Template: help.EmailAfterRegistration, Email: status.Email, Status: status}))
	return err
}
