package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	. "github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func setup() *file.Store {
	// sudo docker run -e MINIO_SECRET_KEY=bbbbbbbb -e MINIO_ACCESS_KEY=aaa -p 9000:9000  minio/minio server /home/shared

	store, err := file.NewStore("localhost:9000", "aaa", "bbbbbbbb", true)

	if err != nil {
		log.Fatal(err)
	}

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
	status, user, err := registerUser(store, db, CreateUserInput{Email: "jack@prymr.nl", Password: " abcabcabc"})

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
		log.Printf("sha %s", fv.SHA256)
		reader, err := store.DownloadFile(token.Salt, token.Bucket, fv.StoreID)
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

	testShaDiff(t, db, token)
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
	if used != 77 {
		t.Fatalf("expected 77 got %d", used)
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
	for i := 0; i < 9; i++ {
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

	list := listSync(store, token.Bucket)
	if len(list) != 0 {
		t.Fatalf("items in the store: %v", list)
	}
}

func listSync(s *file.Store, tokenID string) []string {
	out := make(chan string)
	e := make(chan error)
	go s.ListObjects(tokenID, e, out)
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

func testShaDiff(t *testing.T, db *gorm.DB, token *file.Token) {
	_, err := file.ShaDiff(db, token, bytes.NewBuffer([]byte("abc")))
	if err == nil {
		t.Fatalf("expected error")
	}

	_, err = file.ShaDiff(db, token, bytes.NewBuffer([]byte("e8fb44fdcd108c238ea4a809bc758ffa5ebe636a  mail.go")))
	if err == nil {
		t.Fatalf("expected error")
	}
	diff, err := file.ShaDiff(db, token, bytes.NewBuffer([]byte(`21d551e4428872a077c6e76d8d9eda8d9b4714ae8ac1e98e084d1f1d48f1eb67  action_log.go
2997f66d71b5c0f2f396872536beed30835add1e1de8740b3136c9d550b1eb7c  api
2997f66d71b5c0f2f396872536beed30835add1e1de8740b3136c9d550b1eb7c  api2
8719d1dc6f98ebb5c04f8c1768342e865156b1582806b6c7d26e3fbdc99b8762  file_test.go
8d0a34b05558ad54c4a5949cc42636165b6449cf3324406d62e923bc060478dc  file_test.go.dl
c7c2c1d3c83afbc522ae08779cd661546e578b2dfc6a398467d293bd63e03290  mail.go
16c20b5cc3f937d49d6e003db609b4a8872eea9a4cb41028dad5cae6bd551e1b  mail.go2
16c20b5cc3f937d49d6e003db609b4a8872eea9a4cb41028dad5cae6bd551e1b  main.go
16c20b5cc3f937d49d6e003db609b4a8872eea9a4cb41028dad5cae6bd551e1b  main.go.dl
fc29fe749e8c62050094724e2bed50b65a508e18101eb7d6fdea11be77b2515b  util.go`)))
	if err != nil {
		t.Fatal(err)
	}

	if len(diff) != 7 {
		t.Fatalf("expected 10 got %d, diff: %q", len(diff), diff)
	}
	log.Printf("diff: %s", diff)
}
