package main

import (
	"flag"
	"fmt"
	"os"

	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/jackdoe/baxx/message"
	"github.com/jackdoe/baxx/monitoring"
)

func main() {
	var pkind = flag.String("kind", "", "what kind is it")
	var pinterval = flag.Int("interval", 60, "how often does it run in seconds + some delta")
	var pmessage = flag.String("message", "not working", "what message do you want")
	var pdebug = flag.Bool("debug", false, "debug")
	flag.Parse()
	if *pkind == "" {
		log.Fatal("need -kind")
	}
	kind := fmt.Sprintf("crony-%s", *pkind)
	defer message.SlackPanic(kind)

	flag.Parse()
	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()

	monitoring.MustInitNode(db, kind, *pmessage, (time.Duration(*pinterval) * time.Second).Seconds())
	monitoring.MustTick(db, kind)
}
