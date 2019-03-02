package main

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"os"
)

type sendMailConfig struct {
	from        string
	to          []string
	subject     string
	body        string
	contentType string
}

func sendmail(sm sendMailConfig) error {
	m := gomail.NewMessage()
	m.SetHeader("From", sm.from)
	user := "apikey"
	pass := os.Getenv("BAXX_SENDGRID_KEY")

	m.SetHeader("To", sm.to...)
	m.SetHeader("Bcc", "info@baxx.dev")
	m.SetHeader("Subject", sm.subject)
	if sm.contentType == "" {
		sm.contentType = "text/plain"
	}
	m.SetBody(sm.contentType, sm.body)

	d := gomail.NewDialer("smtp.sendgrid.net", 465, user, pass)

	return d.DialAndSend(m)
}
