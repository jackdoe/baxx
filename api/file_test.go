package main

import (
	"bytes"
	. "github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"
	. "github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileQuota(t *testing.T) {
	IS_TESTING = true

	dir, err := ioutil.TempDir("", "test_file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	file.ROOT = dir

	tmpfn := filepath.Join(dir, "tmpfile")
	db, err := gorm.Open("sqlite3", tmpfn)
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)

	defer db.Close()
	initDatabase(db)
	status, user, err := registerUser(db, CreateUserInput{Email: "jack@prymr.nl", Password: " abcabcabc"})
	if err != nil {
		t.Fatal(err)
	}

	/* test uploading a file */
	log.Printf("%#v", user)
	now := time.Now()
	user.StartedSubscription = &now
	user.EmailVerified = &now

	if err := db.Save(user).Error; err != nil {
		t.Fatal(err)
	}

	token, _, err := FindToken(db, user.SemiSecretID, status.TokenRW)
	if err != nil {
		t.Fatal(err)
	}
	filePath := "/example/example.txt"
	body := []byte("a b c d")

	fv, err := file.SaveFile(db, token, bytes.NewBuffer(body), filePath)
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := user.ListTokens(db)
	if err != nil {
		t.Fatal(err)
	}

	used := uint64(0)
	for _, t := range tokens {
		used += t.SizeUsed
	}
	if used != 7 {
		t.Fatalf("expected 7 got %d", used)
	}
	log.Printf("%#v %#v used %d", tokens, fv, used)
}
