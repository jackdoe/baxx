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
	users := []*user.User{}
	if err := db.Find(&users).Error; err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		tokens := []*file.Token{}
		if err := db.Where("user_id = ?", u.ID).Find(&tokens).Error; err != nil {
			log.Fatal(err)
		}
		grouped := []PerRuleGroup{}
	TOKEN:
		for _, t := range tokens {
			rules := []*notification.NotificationRule{}
			if err := db.Where("user_id = ? AND token_id = ?", u.ID, t.ID).Find(&rules).Error; err != nil {
				log.Fatal(err)
			}
			if len(rules) == 0 {
				continue TOKEN
			}
			files, err := file.ListFilesInPath(db, t, "", false)
			if err != nil {
				log.Fatal(err)
			}

			for _, rule := range rules {
				// get all files in token
				t := &file.Token{}
				if err := db.First(&t, rule.TokenID).Error; err != nil {
					log.Fatal(err)
				}

				u := &user.User{}
				if err := db.First(&u, rule.UserID).Error; err != nil {
					log.Fatal(err)
				}

				pf, err := notification.ExecuteRule(rule, files)
				if err != nil {
					log.Fatal(err)
				}
				//				for _, per := range pf {
				//					nfv := &NotificationForFileVersion{NotificationRuleID: rule.ID, FileVersionID: 1}
				//				}

				grouped = append(grouped, PerRuleGroup{
					Rule:    rule,
					PerFile: pf,
				})
			}
		}

		if len(grouped) != 0 {
			uuid := common.GetUUID()
			// check if the file was notified before

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

			if err := db.Save(n).Error; err != nil {
				log.Fatal(err)
			}
		}
	}
}
