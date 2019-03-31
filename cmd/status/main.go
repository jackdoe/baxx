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

func main() {
	message.MustHavePanic()
	message.MustHaveMonitoring()

	defer message.SlackPanic("node status")
	var pdebug = flag.Bool("debug", false, "debug")
	var pdiskThresh = flag.Float64("disk-used", 0.8, "alert if disk is more than 0.8 used")
	var pdisk = flag.String("disk", "md2", "disk for mdadm")
	var pinterval = flag.Int("interval", 60, "interval in seconds")
	flag.Parse()

	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()
	wait := (time.Duration(*pinterval+10) * time.Second).Seconds()
	monitoring.MustInitNode(db, KIND, fmt.Sprintf("node status not working working for %f seconds", wait), wait)
	diskName := *pdisk
	lastError := time.Unix(0, 0)
	for {
		du := monitoring.GetDiskUsage("/")
		md := monitoring.GetMDADM(diskName)
		dio := monitoring.GetDiskIOStats(diskName)
		mem := monitoring.GetMemoryStats()
		used := float64(du.DiskUsed) / float64(du.DiskAll)
		errorSent := false
		if md.ExitCode != 0 {
			if time.Since(lastError).Seconds() > 3600 {
				message.SendSlackMonitoring("disk "+diskName+" md error", fmt.Sprintf("```%s\n%s```", diskName, md.MDADM))
				errorSent = true
			}
		}

		if used > *pdiskThresh {
			if time.Since(lastError).Seconds() > 3600 {
				message.SendSlackMonitoring("disk "+diskName+" full", fmt.Sprintf("```used: %f%%, allB: %d, usedB: %d```", (used*100), du.DiskAll, du.DiskUsed))
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
		time.Sleep(time.Duration(*pinterval) * time.Second)
	}
}
