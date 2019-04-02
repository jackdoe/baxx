package main

import (
	"flag"
	"fmt"
	"os"

	"testing"
	"time"

	"github.com/jackdoe/baxx/monitoring"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func TestMain(t *testing.T) {
	//	var pdisk = flag.String("disk", "nvme0n1p1", "disk for mdadm")
	var pinterval = flag.Int("interval", 60, "interval in seconds")
	flag.Parse()

	db, err := gorm.Open("postgres", "host=localhost user=baxx dbname=baxx password=baxx")
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)
	defer db.Close()
	wait := (time.Duration(*pinterval+10) * time.Second).Seconds()
	monitoring.MustInitNode(db, KIND, fmt.Sprintf("node status not working working for %f seconds", wait), wait)
	//	diskName := *pdisk
	for i := 0; i < 100; i++ {

		mem := monitoring.GetMemoryStats()

		if err := db.Save(&mem).Error; err != nil {
			log.Panic(err)
		}

		monitoring.MustTick(db, KIND)
	}
	s, err := monitoring.GetStats(db, monitoring.Hostname(), 10)
	if err != nil {
		log.Panic(err)
	}
	s[0].ASCII(os.Stdout, 30)
}
