package main

import (
	. "github.com/jackdoe/baxx/config"
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
