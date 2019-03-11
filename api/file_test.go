package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	. "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/config"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
	. "github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func setup() *file.Store {
	// sudo docker run -e MINIO_SECRET_KEY=bbbbbbbb -e MINIO_ACCESS_KEY=aaa -p 9000:9000  minio/minio server /home/shared

	store, err := file.NewStore(&StoreConfig{
		Endpoint:        "localhost:9000",
		Region:          "",
		Bucket:          "baxx",
		AccessKeyID:     "aaa",
		SecretAccessKey: "bbbbbbbb",
		SessionToken:    "",
		DisableSSL:      true,
	})

	if err != nil {
		log.Fatal(err)
	}

	store.MakeBucket()
	return store
}

func TestFileQuota(t *testing.T) {
	store := setup()
	db, err := gorm.Open("postgres", "host=localhost user=baxx dbname=baxx password=baxx")
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

	log.Print(help.Render(help.EMAIL_AFTER_REGISTRATION, status))
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
	var fmFirst *file.FileMetadata
	for i := 0; i < 20; i++ {
		s := fmt.Sprintf("a b c d %d", i)
		_, fmFirst, err = SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(s)), filePath)
		if err != nil {
			t.Fatal(err)
		}

		fv, _, err := file.FindFile(db, token, filePath)
		if err != nil {
			t.Fatal(err)
		}
		reader, err := store.DownloadFile(fv)
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
	}

	_, fmSecond, err := SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d"))), filePath+"second")
	if err != nil {
		t.Fatal(err)
	}

	versions, err := file.ListVersionsFile(db, token, fmFirst)
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 7 {
		t.Fatalf("expected 7 versions got %d", len(versions))
	}

	used := getUsed(t, db, user)
	if used != 71 {
		t.Fatalf("expected 71 got %d", used)
	}
	files, err := file.ListFilesInPath(db, token, "/example/", false)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("%s", file.LSAL(files))
	if len(files) != 2 {
		t.Fatalf("expected 2 files got %d", len(files))
	}

	versions, err = file.ListVersionsFile(db, token, fmSecond)
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 1 {
		t.Fatalf("expected1 versions got %d", len(versions))
	}

	err = file.DeleteFile(store, db, token, fmFirst)
	if err != nil {
		t.Fatal(err)
	}

	used = getUsed(t, db, user)
	if used != 7 {
		t.Fatalf("expected 7 got %d", used)
	}

	err = file.DeleteFile(store, db, token, fmSecond)
	if err != nil {
		t.Fatal(err)
	}

	used = getUsed(t, db, user)
	if used != 0 {
		t.Fatalf("expected 0 got %d", used)
	}

	user.Quota = 10
	user.QuotaInode = 2
	if err := db.Save(user).Error; err != nil {
		t.Fatal(err)
	}

	_, _, err = SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d"))), filePath+"second")
	if err != nil {
		t.Fatal(err)
	}
	left, inodeLeft, _ := user.GetQuotaLeft(db)
	log.Printf("left: %d inodes: %d", left, inodeLeft)
	if inodeLeft != 1 {
		t.Fatalf("expected 1 got %d", inodeLeft)
	}

	_, inodeLeft, _ = user.GetQuotaLeft(db)
	if inodeLeft != 1 {
		t.Fatalf("(after error) expected 1 got %d", inodeLeft)
	}

	user.Quota = 500
	if err := db.Save(user).Error; err != nil {
		t.Fatal(err)
	}
	_, _, err = SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), filePath+"second")
	if err != nil {
		t.Fatal(err)
	}

	left, inodeLeft, _ = user.GetQuotaLeft(db)
	if inodeLeft != 0 {
		t.Fatalf("expected 1 got %d", inodeLeft)
	}
	log.Printf("left: %d inodes: %d", left, inodeLeft)
	_, _, err = SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), filePath+"second")
	if err.Error() != "inode quota limit reached" {
		t.Fatalf("expected inode quota limit reached got %s", err.Error())
	}

	CONFIG.MaxTokens = 10
	created := []*file.Token{}
	for i := 0; i < 8; i++ {
		to, err := user.CreateToken(db, false, 1, "some-name")
		if err != nil {
			t.Fatal(err)
		}
		created = append(created, to)
	}
	_, err = user.CreateToken(db, false, 1, "some-other-name")
	if err.Error() != "max tokens created (max=10)" {
		t.Fatalf("expected max tokens created (max=10) got %s", err.Error())
	}

	created = append(created, token)
	for _, to := range created {
		err := file.DeleteToken(store, db, to)
		if err != nil {
			t.Fatal(err)
		}
	}

	list := listSync(store)
	if len(list) != 0 {
		t.Fatalf("items in the store: %v", list)
	}
}

func listSync(s *file.Store) []string {
	out := make(chan string)
	e := make(chan error)
	go s.ListObjects(e, out)
	res := []string{}
	for v := range out {
		res = append(res, v)
	}
	return res
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
