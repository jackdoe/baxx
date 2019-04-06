package main

import (
	"flag"
	"fmt"
	"os"

	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/jackdoe/baxx/api/file"
	"github.com/jackdoe/baxx/api/helpers"
	"github.com/jackdoe/baxx/api/notification_rules"
	notification "github.com/jackdoe/baxx/api/notification_rules"
	"github.com/jackdoe/baxx/api/user"
	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/message"
	"github.com/jackdoe/baxx/monitoring"
)

const KIND = "notification_run"

func main() {
	message.MustHavePanic()
	defer message.SlackPanic("notification rules")
	var pdebug = flag.Bool("debug", false, "debug")
	flag.Parse()
	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()
	monitoring.MustInitNode(db, KIND, "notification run not working for 1+Minute", (2 * time.Minute).Seconds())
	for {
		runRules(db)
		monitoring.MustTick(db, KIND)
		time.Sleep(1 * time.Minute)
	}
}

func runRules(db *gorm.DB) {
	tx := db.Begin()
	users := []*user.User{}
	if err := tx.Find(&users).Error; err != nil {
		tx.Rollback()
		log.Panic(err)
	}

	for _, u := range users {
		status, err := helpers.GetUserStatus(tx, u)
		if err != nil {
			tx.Rollback()
			log.Panic(err)
		}

		sendQuotaNotification := false
		if status.QuotaUsed >= uint64(float64(u.Quota)*0.9) || status.QuotaInodeUsed >= uint64(float64(u.QuotaInode)*0.9) {
			alreadySentForToken := &notification.NotificationForUserQuota{}
			if tx.Where("user_id = ?", u.ID).First(&alreadySentForToken).RecordNotFound() {
				sendQuotaNotification = true
				if err := tx.Create(&notification.NotificationForUserQuota{UserID: u.ID}).Error; err != nil {
					tx.Rollback()
					log.Panic(err)
				}
			}
		} else {
			// always delete the fact that we sent
			// so we can send again if space gets cleared and then full again
			tx.Where("user_id = ?", u.ID).Delete(&notification.NotificationForUserQuota{})
		}

		if sendQuotaNotification {
			err := message.EnqueueMail(
				tx,
				u.ID,
				"[ baxx.dev ] quota limit reached",
				help.Render(help.HelpObject{
					Template: help.EmailQuotaLeft,
					Email:    u.Email,
					Status:   status,
				}))
			if err != nil {
				tx.Rollback()
				log.Panic(err)
			}
		}

		count := 0
		grouped := []common.FileNotification{}

		tokens := []*file.Token{}
		if err := tx.Where("user_id = ?", u.ID).Find(&tokens).Error; err != nil {
			tx.Rollback()
			log.Panic(err)
		}

		for _, t := range tokens {
			files, err := file.ListFilesInPath(tx, t, "", false)
			if err != nil {
				tx.Rollback()
				log.Panic(err)
			}

			pf := notification_rules.ExecuteRule(files)
			unseen, err := notification_rules.IgnoreAndMarkAlreadyNotified(tx, pf)
			if err != nil {
				tx.Rollback()
				log.Panic(err)
			}
			grouped = append(grouped, unseen...)
		}

		if len(grouped) != 0 {
			err := message.EnqueueMail(
				tx,
				u.ID,
				fmt.Sprintf("[ baxx.dev ] backup issues for %d files", count),
				help.Render(help.HelpObject{
					Template:      help.EmailNotification,
					Email:         u.Email,
					Notifications: grouped,
					Status:        status,
				}))
			if err != nil {
				tx.Rollback()
				log.Panic(err)
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		log.Panic(err)
	}
}
