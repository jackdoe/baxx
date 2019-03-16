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
	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()

	runRules(db)
}

func runRules(db *gorm.DB) {
	tx := db.Begin()
	users := []*user.User{}
	if err := tx.Find(&users).Error; err != nil {
		log.Fatal(err)
	}

	totalFilesAlerted := 0

	for _, u := range users {
		tokens := []*file.Token{}
		if err := tx.Where("user_id = ?", u.ID).Find(&tokens).Error; err != nil {
			log.Fatal(err)
		}
		totalUsers++

		count := 0
		grouped := []common.PerRuleGroup{}
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
					totalFilesAlerted++
					nfv.Count++
					if err := tx.Save(nfv).Error; err != nil {
						log.Fatal(err)
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
			err := common.EnqueueMail(
				tx,
				u.ID,
				fmt.Sprintf("[ baxx.dev ] backup issues for %d files", count),
				help.Render(help.HelpObject{
					Template:      help.EmailNotification,
					Email:         u.Email,
					Notifications: grouped,
					Status:        common.EMPTY_STATUS,
				}))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		log.Fatal(err)
	}
}
