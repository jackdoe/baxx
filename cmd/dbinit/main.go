package main

import (
	"log"
	"os"

	"github.com/jackdoe/baxx/api/init_db"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/jinzhu/gorm"
)

func main() {
	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(true)
	defer db.Close()

	init_db.InitDatabase(db)
}
