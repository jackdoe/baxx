package main

import (
	"bytes"
	"fmt"
	. "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/config"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
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
	dir, err := ioutil.TempDir("", "test_file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	CONFIG.FileRoot = dir

	tmpfn := filepath.Join(dir, "tmpfile")
	db, err := gorm.Open("sqlite3", tmpfn)
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)

	defer db.Close()
	initDatabase(db)
	status, user, err := registerUser(db, CreateUserInput{Email: "jack@prymr.nl", Password: " abcabcabc"})
	log.Print(help.Render(help.EMAIL_AFTER_REGISTRATION, status))

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

	token, _, err := FindToken(db, status.Tokens[0].UUID)
	if err != nil {
		t.Fatal(err)
	}

	filePath := "/example/example.txt"

	for i := 0; i < 20; i++ {
		s := fmt.Sprintf("a b c d %d", i)
		_, err := file.SaveFile(db, token, user, bytes.NewBuffer([]byte(s)), filePath)
		if err != nil {
			t.Fatal(err)
		}
		_, file, reader, err := file.FindAndOpenFile(db, token, filePath)
		if err != nil {
			t.Fatal(err)
		}

		b, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}

		if string(b) != s {
			t.Fatalf("expected %s got %s", s, string(b))
		}
		file.Close()

	}

	_, err = file.SaveFile(db, token, user, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d"))), filePath+"second")
	if err != nil {
		t.Fatal(err)
	}

	versions, err := file.ListVersionsFile(db, token, filePath)
	if err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Fatal(err)
	}

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

	user.Quota = 10
	if err := db.Save(user).Error; err != nil {
		t.Fatal(err)
	}

	_, err = file.SaveFile(db, token, user, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d"))), filePath+"second")
	if err != nil {
		t.Fatal(err)
	}

	left, _ := user.GetQuotaLeft(db)
	log.Printf("left: %d", left)

	_, err = file.SaveFile(db, token, user, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), filePath+"second")
	if err.Error() != "quota limit reached" {
		t.Fatalf("expected quota limit reached got %s", err.Error())
	}
	left, _ = user.GetQuotaLeft(db)
	log.Printf("left: %d", left)

	CONFIG.MaxTokens = 10
	created := []*Token{}
	for i := 0; i < 8; i++ {
		to, err := user.CreateToken(db, false, 1)
		if err != nil {
			t.Fatal(err)
		}
		created = append(created, to)
	}
	_, err = user.CreateToken(db, false, 1)
	if err.Error() != "max tokens created (max=10)" {
		t.Fatalf("expected max tokens created (max=10) got %s", err.Error())
	}

	created = append(created, token)
	for _, to := range created {
		err := file.DeleteToken(db, to)
		if err != nil {
			t.Fatal(err)
		}
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
