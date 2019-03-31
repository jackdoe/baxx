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
		// XXX: send sms at this point
		log.Warnf("error sending to slack: title: %s, text: %s, error: %s", title, text, err)
	}

}

const KIND = "who_watches_the_watchers"

func main() {
	message.MustHavePanic()
	message.MustHaveMonitoring()
	defer message.SlackPanic("who_watches_the_watchers")
	var pdebug = flag.Bool("debug", false, "debug")
	var pprintIdAndExit = flag.Bool("id", false, "print id and exit")
	flag.Parse()
	if *pprintIdAndExit {
		fmt.Println(monitoring.Hostname())
		os.Exit(0)
	}

	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()
	monitoring.MustInitNode(db, KIND, "monitoring not working for more than 5 seconds", (5 * time.Second).Seconds())
	lastError := time.Now()
	for {
		items, err := monitoring.Watch(db)
		if err != nil {
			if time.Since(lastError).Seconds() > 60 {
				message.SendSlackMonitoring("error watching", fmt.Sprintf("monitoring\n```%s```", err.Error()))
				lastError = time.Now()
			}
		}

		if len(items) > 0 {
			m := ""
			for _, item := range items {
				m = fmt.Sprintf("%s%s\n", m, item.String())
			}
			message.SendSlackMonitoring("monitoring", fmt.Sprintf("monitoring\n```%s```", m))
			for _, item := range items {
				t := time.Now()
				item.NotifiedAt = &t
				if err := db.Save(item).Error; err != nil {
					log.Panic(err)
				}
			}
		}

		monitoring.MustTick(db, KIND)
		time.Sleep(1 * time.Second)
	}
}
