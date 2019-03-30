package main

import (
	"flag"
	"fmt"
	"os"

	"time"

	"github.com/jackdoe/baxx/message"
	"github.com/jackdoe/baxx/monitoring"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func send(key, title, text string) {
	err := message.SendSlack(key, title, text)
	if err != nil {
		log.Warnf("error sending to slack: title: %s, text: %s, error: %s", title, text, err)
	}

}
func main() {
	defer message.SlackPanic("monitoring")
	var pdebug = flag.Bool("debug", false, "debug")
	var pprintIdAndExit = flag.Bool("id", false, "print id and exit")
	flag.Parse()
	if *pprintIdAndExit {
		fmt.Println(monitoring.Hostname())
		os.Exit(0)
	}
	slackMonitoring := os.Getenv("BAXX_SLACK_MONITORING")
	if slackMonitoring == "" {
		log.Fatalf("empty BAXX_SLACK_MONITORING")
	}
	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()
	lastError := time.Now()
	for {
		items, err := monitoring.Watch(db)
		if err != nil {
			if time.Since(lastError).Seconds() > 60 {
				send(slackMonitoring, "error watching", err.Error())
				lastError = time.Now()
			}
		}

		if len(items) > 0 {
			m := ""
			for _, item := range items {
				m = fmt.Sprintf("%s%s\n", m, item.String())
			}
			send(slackMonitoring, "monitoring", fmt.Sprintf("```%s```", m))
			for _, item := range items {
				t := time.Now()
				item.NotifiedAt = &t
				if err := db.Save(item).Error; err != nil {
					log.Panic(err)
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}
