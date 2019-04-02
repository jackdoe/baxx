package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"testing"
	"time"
	"unsafe"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/jackdoe/baxx/api/file"
	"github.com/jackdoe/baxx/api/helpers"
	"github.com/jackdoe/baxx/api/init_db"
	"github.com/jackdoe/baxx/api/user"
	"github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/help"
)

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func setup() *file.Store {
	// sudo docker run -p 9122:9122  jackdoe/judoc:0.6

	store, err := file.NewStore("http://localhost:9122")
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	return store
}

func testNotificationCreate(t *testing.T, db *gorm.DB, u *user.User, token *file.Token) {
	n, err := helpers.CreateNotificationRule(db, u, &common.CreateNotificationInput{
		TokenUUID:         token.UUID,
		Regexp:            ".*",
		Name:              "hello",
		AcceptableAgeDays: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("%#v", n)
}
func drop(db *gorm.DB) {
	db.Exec(`

DO $$ DECLARE
    r RECORD;
BEGIN
    -- if the schema you operate on is not "current", you will want to
    -- replace current_schema() in query with 'schematodeletetablesfrom'
    -- *and* update the generate 'DROP...' accordingly.
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;
 `)
}

func uploadFile(filePath string, t *testing.T, db *gorm.DB, store *file.Store, user *user.User, token *file.Token, size int) *file.FileMetadata {
	s := RandStringBytesMaskImprSrcUnsafe(size)
	_, _, err := SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(s)), filePath)
	if err != nil {
		t.Fatal(err)
	}

	fv, fm, err := file.FindFile(db, token, filePath)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("store id %s", fv.StoreID)

	reader, err := store.DownloadFile(token.Salt, token.UUID, fv.StoreID)
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
	return fm
}

func createUserAndToken(email string, t *testing.T, db *gorm.DB, store *file.Store) (*user.User, *file.Token) {
	status, user, err := registerUser(store, db, CreateUserInput{Email: email, Password: " abcabcabc"})
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("%#v", status)
	log.Print(help.Render(help.HelpObject{Template: help.EmailAfterRegistration, Email: status.Email, Status: status}))

	if err := db.Save(user).Error; err != nil {
		t.Fatal(err)
	}
	_, err = helpers.CreateToken(db, false, user, 7, "some-name", 10000, 10000, CONFIG.MaxTokens)
	if err != nil {
		t.Fatal(err)
	}

	status, err = helpers.GetUserStatus(db, user)
	if err != nil {
		t.Fatal(err)
	}

	token, _, err := helpers.FindTokenAndUser(db, status.Tokens[0].UUID)
	if err != nil {
		t.Fatal(err)
	}
	return user, token

}

func TestFileQuota(t *testing.T) {
	store := setup()
	db, err := gorm.Open("postgres", "host=localhost user=baxx dbname=baxx password=baxx")
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)
	drop(db)
	init_db.InitDatabase(db)
	defer db.Close()
	for k := 0; k < 10; k++ {
		user, token := createUserAndToken(RandStringBytesMaskImprSrcUnsafe(10)+"@example.com", t, db, store)
		testNotificationCreate(t, db, user, token)
		filePath := "/example/example.txt"
		var fmFirst *file.FileMetadata
		for i := 0; i < 20; i++ {
			fmFirst = uploadFile(filePath, t, db, store, user, token, i)
		}

		fmSecond := uploadFile(filePath+"second", t, db, store, user, token, 1000)
		versions, err := file.ListVersionsFile(db, token, fmFirst)
		if err != nil {
			t.Fatal(err)
		}

		if len(versions) != 7 {
			t.Fatalf("expected 7 versions got %d", len(versions))
		}

		files, err := file.ListFilesInPath(db, token, "/example/", false)
		if err != nil {
			t.Fatal(err)
		}

		log.Printf("%s", LSAL(files))
		if len(files) != 2 {
			t.Fatalf("expected 2 files got %d", len(files))
		}

		err = file.DeleteFile(store, db, token, fmFirst)
		if err != nil {
			t.Fatal(err)
		}

		used := getUsed(t, db, user)
		if used == 0 {
			t.Fatalf("expected something got %d", used)
		}

		err = file.DeleteFile(store, db, token, fmSecond)
		if err != nil {
			t.Fatal(err)
		}

		used = getUsed(t, db, user)
		if used != 0 {
			t.Fatalf("expected 0 got %d", used)
		}

		token.Quota = 10
		token.QuotaInode = 2
		if err := db.Save(token).Error; err != nil {
			t.Fatal(err)
		}

		_, _, err = SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d"))), filePath+"second")
		if err != nil {
			t.Fatal(err)
		}
		left, inodeLeft, _ := file.GetQuotaLeft(db, token)
		log.Printf("left: %d inodes: %d", left, inodeLeft)
		if inodeLeft != 1 {
			t.Fatalf("expected 1 got %d", inodeLeft)
		}

		_, inodeLeft, _ = file.GetQuotaLeft(db, token)
		if inodeLeft != 1 {
			t.Fatalf("(after error) expected 1 got %d", inodeLeft)
		}

		token.Quota = 500
		if err := db.Save(token).Error; err != nil {
			t.Fatal(err)
		}
		_, _, err = SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), filePath+"second")
		if err != nil {
			t.Fatal(err)
		}

		left, inodeLeft, _ = file.GetQuotaLeft(db, token)
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
			to, err := helpers.CreateToken(db, false, user, 1, "some-name", 10000, 10000, CONFIG.MaxTokens)
			if err != nil {
				t.Fatal(err)
			}
			created = append(created, to)
		}
		_, err = helpers.CreateToken(db, false, user, 1, "some-other-name", 10000, 10000, CONFIG.MaxTokens)
		if err.Error() != "max tokens created (max=10)" {
			t.Fatalf("expected max tokens created (max=10) got %s", err.Error())
		}
		CONFIG.MaxUserQuota = 1
		_, _, err = SaveFileProcess(store, db, user, created[0], bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), filePath+"secondz")
		if err.Error() != "quota limit reached" {
			t.Fatalf("expected quota limit reached got %s", err.Error())
		}

		CONFIG.MaxUserQuota = 10000
		_, _, err = SaveFileProcess(store, db, user, created[0], bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), filePath+"secondz")
		if err != nil {
			t.Fatal(err)
		}

		created = append(created, token)
		for _, to := range created {
			err := file.DeleteToken(store, db, to)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func getUsed(t *testing.T, db *gorm.DB, user *user.User) uint64 {
	tokens, err := helpers.ListTokens(db, user)
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
	_, err := ShaDiff(db, token, bytes.NewBuffer([]byte("abc")))
	if err == nil {
		t.Fatalf("expected error")
	}

	_, err = ShaDiff(db, token, bytes.NewBuffer([]byte("e8fb44fdcd108c238ea4a809bc758ffa5ebe636a  mail.go")))
	if err == nil {
		t.Fatalf("expected error")
	}
	diff, err := ShaDiff(db, token, bytes.NewBuffer([]byte(`21d551e4428872a077c6e76d8d9eda8d9b4714ae8ac1e98e084d1f1d48f1eb67  action_log.go
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
