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
	notification "github.com/jackdoe/baxx/api/notification_rules"
	"github.com/jackdoe/baxx/api/user"
	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/message"
)

func main() {
	defer message.SlackPanic("notification rules")
	var pdebug = flag.Bool("debug", false, "debug")
	flag.Parse()
	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()

	for {
		runRules(db)
		time.Sleep(1 * time.Hour)
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
		tokens := []*file.Token{}
		if err := tx.Where("user_id = ?", u.ID).Find(&tokens).Error; err != nil {
			tx.Rollback()
			log.Panic(err)
		}
		sendQuotaNotification := false
		for _, t := range tokens {
			// if the quota is over, just send an email about it
			leftSize, leftInodes, err := file.GetQuotaLeft(tx, t)
			if err != nil {
				tx.Rollback()
				log.Panic(err)
			}

			if leftInodes < 1 || leftSize <= 0 {
				alreadySentForToken := &notification.NotificationForQuota{}
				if tx.Where("token_id = ?", t.ID).First(&alreadySentForToken).RecordNotFound() {
					sendQuotaNotification = true
					if err := tx.Create(&notification.NotificationForQuota{TokenID: t.ID}).Error; err != nil {
						tx.Rollback()
						log.Panic(err)
					}
				}
			} else {
				// always delete the fact that we sent
				// so we can send again if space gets cleared and then full again
				tx.Where("token_id = ?", t.ID).Delete(&notification.NotificationForQuota{})
			}
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
		grouped := []common.PerRuleGroup{}
	TOKEN:
		for _, t := range tokens {
			rules := []*notification.NotificationRule{}
			if err := tx.Where("user_id = ? AND token_id = ?", u.ID, t.ID).Find(&rules).Error; err != nil {
				tx.Rollback()
				log.Panic(err)
			}
			if len(rules) == 0 {
				continue TOKEN
			}
			files, err := file.ListFilesInPath(tx, t, "", false)
			if err != nil {
				tx.Rollback()
				log.Panic(err)
			}

			for _, rule := range rules {
				// get all files in token
				t := &file.Token{}
				if err := tx.First(&t, rule.TokenID).Error; err != nil {
					tx.Rollback()
					log.Panic(err)
				}

				u := &user.User{}
				if err := tx.First(&u, rule.UserID).Error; err != nil {
					tx.Rollback()
					log.Panic(err)
				}

				pf, err := notification.ExecuteRule(rule, files)
				if err != nil {
					tx.Rollback()
					log.Panic(err)
				}
				unseen := []common.FileNotification{}
				for _, p := range pf {
					nfv := &notification.NotificationForFileVersion{
						NotificationRuleID: rule.ID,
						FileVersionID:      p.FileVersionID,
					}
					res := tx.Where(nfv).First(&nfv)
					if res.RecordNotFound() {
						unseen = append(unseen, p)
					}
					count++
					nfv.Count++
					if err := tx.Save(nfv).Error; err != nil {
						tx.Rollback()
						log.Panic(err)
					}
				}

				if len(unseen) > 0 {
					grouped = append(grouped, common.PerRuleGroup{
						Rule:    notification.TransformRuleToOutput(rule),
						PerFile: unseen,
					})
				}
			}
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
