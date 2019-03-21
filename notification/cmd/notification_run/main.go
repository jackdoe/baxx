package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jackdoe/baxx/api/helpers"
	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/notification"
	"github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func main() {
	defer notification.SlackPanic("notification rules")
	var pdebug = flag.Bool("debug", false, "debug")
	flag.Parse()
	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()

	runRules(db)
}
func runRules(db *gorm.DB) {
	tx := db.Begin()
	users := []*user.User{}
	if err := tx.Find(&users).Error; err != nil {
		log.Panic(err)
	}

	for _, u := range users {
		status, err := helpers.GetUserStatus(tx, u)
		if err != nil {
			log.Panic(err)
		}
		tokens := []*file.Token{}
		if err := tx.Where("user_id = ?", u.ID).Find(&tokens).Error; err != nil {
			log.Panic(err)
		}
		sendQuotaNotification := false
		for _, t := range tokens {
			// if the quota is over, just send an email about it
			leftSize, leftInodes, err := file.GetQuotaLeft(tx, t)
			if err != nil {
				log.Panic(err)
			}

			if leftInodes < 1 || leftSize <= 0 {
				alreadySentForToken := &notification.NotificationForQuota{}
				if tx.Where("token_id = ?", t.ID).First(&alreadySentForToken).RecordNotFound() {
					sendQuotaNotification = true
					if err := tx.Create(&notification.NotificationForQuota{TokenID: t.ID}).Error; err != nil {
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
			err := notification.EnqueueMail(
				tx,
				u.ID,
				"[ baxx.dev ] quota limit reached",
				help.Render(help.HelpObject{
					Template: help.EmailQuotaLeft,
					Email:    u.Email,
					Status:   status,
				}))
			if err != nil {
				log.Panic(err)
			}
		}

		count := 0
		grouped := []common.PerRuleGroup{}
	TOKEN:
		for _, t := range tokens {
			rules := []*notification.NotificationRule{}
			if err := tx.Where("user_id = ? AND token_id = ?", u.ID, t.ID).Find(&rules).Error; err != nil {
				log.Panic(err)
			}
			if len(rules) == 0 {
				continue TOKEN
			}
			files, err := file.ListFilesInPath(tx, t, "", false)
			if err != nil {
				log.Panic(err)
			}

			for _, rule := range rules {
				// get all files in token
				t := &file.Token{}
				if err := tx.First(&t, rule.TokenID).Error; err != nil {
					log.Panic(err)
				}

				u := &user.User{}
				if err := tx.First(&u, rule.UserID).Error; err != nil {
					log.Panic(err)
				}

				pf, err := notification.ExecuteRule(rule, files)
				if err != nil {
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
			err := notification.EnqueueMail(
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
				log.Panic(err)
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		log.Panic(err)
	}
}
