package main

import (
	"bytes"
	"fmt"
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

	for i := 0; i < 20; i++ {
		_, err := file.SaveFile(db, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d %d", i))), filePath)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = file.SaveFile(db, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d"))), filePath+"second")
	if err != nil {
		t.Fatal(err)
	}

	versions, err := file.ListVersionsFile(db, token, filePath)
	if len(versions) != 7 {
		t.Fatalf("expected 7 versions got %d", len(versions))
	}

	used := getUsed(t, db, user)
	if used != 77 {
		t.Fatalf("expected 77 got %d", used)
	}
	files, err := file.ListFilesInPath(db, token, "/example/")
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("%s", file.LSAL(files))
	if len(files) != 2 {
		t.Fatalf("expected 2 files got %d", len(files))
	}

	versions, err = file.ListVersionsFile(db, token, filePath+"second")
	if len(versions) != 1 {
		t.Fatalf("expected1 versions got %d", len(versions))
	}

	err = file.DeleteFile(db, token, filePath)
	if err != nil {
		t.Fatal(err)
	}

	used = getUsed(t, db, user)
	if used != 7 {
		t.Fatalf("expected 7 got %d", used)
	}

	err = file.DeleteFile(db, token, filePath+"second")
	if err != nil {
		t.Fatal(err)
	}

	used = getUsed(t, db, user)
	if used != 0 {
		t.Fatalf("expected 0 got %d", used)
	}
}

func getUsed(t *testing.T, db *gorm.DB, user *User) uint64 {
	tokens, err := user.ListTokens(db)
	if err != nil {
		t.Fatal(err)
	}

	used := uint64(0)
	for _, t := range tokens {
		used += t.SizeUsed
	}
	return used
}
