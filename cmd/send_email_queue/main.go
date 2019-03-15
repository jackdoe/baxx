package main

import (
	"flag"
	"os"
	"time"

	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func main() {
	var pdebug = flag.Bool("debug", false, "debug")
	flag.Parse()
	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()

	sendEmails(db)
}

func sendEmails(db *gorm.DB) {
	sendgrid := os.Getenv("BAXX_SENDGRID_KEY")
	if sendgrid == "" {
		log.Fatalf("empty BAXX_SENDGRID_KEY")
	}
	toSend := []*common.EmailQueueItem{}
	if err := db.Where("sent = ?", false).Order("asc").Find(&toSend).Error; err != nil {
		log.Fatal(err)
	}
	for _, m := range toSend {
		u := &user.User{}
		if err := db.Where("user_id = ?", m.UserID).First(&u).Error; err != nil {
			log.Fatal(err)
		}
		if u.EmailVerified == nil {
			log.Infof("skipping notification for %s, unverified email", u.Email)
			continue
		}

		err := common.Sendmail(sendgrid, common.Message{
			From:    "jack@baxx.dev",
			To:      []string{u.Email},
			Subject: m.EmailSubject,
			Body:    m.EmailText,
		})
		if err != nil {
			log.Fatal(err)
		}

		m.SentAt = time.Now()
		m.Sent = true
		if err := db.Save(m).Error; err != nil {
			log.Fatal(err)
		}
	}
}
