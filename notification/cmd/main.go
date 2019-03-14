package main

import (
	"flag"
	"fmt"
	"os"

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
	var pdebug = flag.Bool("debug", false, "debug")
	flag.Parse()
	//	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	db, err := gorm.Open("postgres", "host=localhost user=baxx dbname=baxx password=baxx")
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()

	runRules(db)
	//	sendEmails(db)
}

func sendNotificationEmails(db *gorm.DB) {
	sendgrid := os.Getenv("BAXX_SENDGRID_KEY")
	if sendgrid == "" {
		log.Fatalf("empty BAXX_SENDGRID_KEY")
	}

	for {

	}
}

type PerRuleGroup struct {
	PerFile []notification.FileNotification
	Rule    *notification.NotificationRule
}

func runRules(db *gorm.DB) {
	tx := db.Begin()
	users := []*user.User{}
	if err := tx.Find(&users).Error; err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		tokens := []*file.Token{}
		if err := tx.Where("user_id = ?", u.ID).Find(&tokens).Error; err != nil {
			log.Fatal(err)
		}
		grouped := []PerRuleGroup{}
	TOKEN:
		for _, t := range tokens {
			rules := []*notification.NotificationRule{}
			if err := tx.Where("user_id = ? AND token_id = ?", u.ID, t.ID).Find(&rules).Error; err != nil {
				log.Fatal(err)
			}
			if len(rules) == 0 {
				continue TOKEN
			}
			files, err := file.ListFilesInPath(tx, t, "", false)
			if err != nil {
				log.Fatal(err)
			}

			for _, rule := range rules {
				// get all files in token
				t := &file.Token{}
				if err := tx.First(&t, rule.TokenID).Error; err != nil {
					log.Fatal(err)
				}

				u := &user.User{}
				if err := tx.First(&u, rule.UserID).Error; err != nil {
					log.Fatal(err)
				}

				pf, err := notification.ExecuteRule(rule, files)
				if err != nil {
					log.Fatal(err)
				}
				unseen := []notification.FileNotification{}
				for _, p := range pf {
					nfv := &notification.NotificationForFileVersion{
						NotificationRuleID: rule.ID,
						FileVersionID:      p.FileVersion.ID,
					}
					res := tx.Where(nfv).First(&nfv)
					if res.RecordNotFound() {
						unseen = append(unseen, p)
					}
					nfv.Count++
					if err := tx.Save(nfv).Error; err != nil {
						log.Fatal(err)
					}
				}

				if len(unseen) > 0 {
					grouped = append(grouped, PerRuleGroup{
						Rule:    rule,
						PerFile: unseen,
					})
				}
			}
		}

		if len(grouped) != 0 {
			uuid := common.GetUUID()

			n := &common.EmailQueueItem{
				UUID:   uuid,
				UserID: u.ID,

				EmailSubject: fmt.Sprintf("[ baxx.dev ] backup issues from %d notification rules", len(grouped)),
				EmailText: help.Render(help.EMAIL_NOTIFICATION, map[string]interface{}{
					"Email":   u.Email,
					"Grouped": grouped,
					"UUID":    uuid,
				}),
			}

			if err := tx.Save(n).Error; err != nil {
				log.Fatal(err)
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		log.Fatal(err)
	}
}
