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

const KIND = "node_status"

func send(key, title, text string) {
	err := message.SendSlack(key, title, text)
	if err != nil {
		log.Warnf("error sending to slack: title: %s, text: %s, error: %s", title, text, err)
	}
}

func main() {
	message.MustHavePanic()

	defer message.SlackPanic("node status")
	var pdebug = flag.Bool("debug", false, "debug")
	flag.Parse()

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
	monitoring.MustInitNode(db, KIND, "node status not working working for 75 seconds", (65 * time.Second).Seconds())
	diskName := os.Getenv("BAXX_MAIN_DISK")
	if diskName == "" {
		diskName = "md2"
	}
	lastError := time.Unix(0, 0)
	for {

		du := monitoring.GetDiskUsage("/")
		md := monitoring.GetMDADM(diskName)
		dio := monitoring.GetDiskIOStats(diskName)
		mem := monitoring.GetMemoryStats()
		free := float64(du.DiskFree) / float64(du.DiskAll)
		errorSent := false
		if md.ExitCode != 0 {
			if time.Since(lastError).Seconds() > 3600 {
				send(slackMonitoring, "disk "+diskName+" md error", fmt.Sprintf("```%s\n%s```", diskName, md.MDADM))
				errorSent = true
			}
		}

		if free > 0 {
			if time.Since(lastError).Seconds() > 3600 {
				send(slackMonitoring, "disk "+diskName+" full", fmt.Sprintf("```%s\nused: %f%%, allB: %d, usedB: %d```", diskName, 100-(free*100), du.DiskAll, du.DiskUsed))
				errorSent = true
			}
		}
		if errorSent {
			lastError = time.Now()
		}

		if err := db.Save(&du).Error; err != nil {
			log.Panic(err)
		}

		if err := db.Save(&md).Error; err != nil {
			log.Panic(err)
		}

		if err := db.Save(&dio).Error; err != nil {
			log.Panic(err)
		}

		if err := db.Save(&mem).Error; err != nil {
			log.Panic(err)
		}

		monitoring.MustTick(db, KIND)
		time.Sleep(1 * time.Minute)
	}
}
