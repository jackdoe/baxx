package main

import (
	"database/sql"
	"flag"
	"os"

	"time"

	"github.com/jackdoe/baxx/api/user"
	"github.com/jackdoe/baxx/message"
	"github.com/jackdoe/baxx/monitoring"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const KIND = "send_email_queue"

func main() {
	message.MustHavePanic()
	defer message.SlackPanic("email send")
	var pdebug = flag.Bool("debug", false, "debug")
	flag.Parse()
	sendgrid := os.Getenv("BAXX_SENDGRID_KEY")
	if sendgrid == "" {
		log.Fatalf("empty BAXX_SENDGRID_KEY")
	}

	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()
	monitoring.MustInitNode(db, KIND, "send email queue run not working for 5 seconds", (5 * time.Second).Seconds())
	for {

		for {
			more := sendSingleEmail(db, sendgrid)
			if !more {
				break
			}
		}
		monitoring.MustTick(db, KIND)
		time.Sleep(1 * time.Second)
	}
}

func sendSingleEmail(db *gorm.DB, sendgrid string) bool {
	tx := db.Begin()
	id := int64(0)

	// dont send messages, besides the verification message, unless the user is verified
	if err := tx.Raw(`
                            SELECT
                              e.id
                            FROM
                              email_queue_items e, users u
                            WHERE
                              1 = 1
                              AND e.user_id = u.id
                              AND e.sent = false
                              AND (e.is_verification_message = true OR u.email_verified IS NOT NULL)
                            LIMIT 1
                            FOR UPDATE NOWAIT`).Row().Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		tx.Rollback()
		log.Panic(err)
	}

	m := &message.EmailQueueItem{}
	if err := tx.Where("id = ? ", id).Take(&m).Error; err != nil {
		tx.Rollback()
		log.Panic(err)
	}
	u := &user.User{}
	if err := tx.Where("id = ?", m.UserID).Take(&u).Error; err != nil {
		log.Panic(err)
	}

	err := message.Sendmail(sendgrid, message.Message{
		From:    "jack@baxx.dev",
		To:      []string{u.Email},
		Subject: m.EmailSubject,
		Body:    m.EmailText,
	})
	if err != nil {
		log.Panic(err)
	}

	if err := tx.Where("id = ?", id).Update(&message.EmailQueueItem{Sent: true, SentAt: time.Now()}).Error; err != nil {
		log.Panic(err)
	}

	if err := tx.Commit().Error; err != nil {
		log.Panic(err)
	}
	return true
}
