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
	notifications "github.com/jackdoe/baxx/api/notification_rules"
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

func fp(filepath string) file.FileParams {
	seven := uint64(7)
	return file.FileParams{FullPath: filepath, KeepN: &seven}
}

func uploadFile(param file.FileParams, t *testing.T, db *gorm.DB, store *file.Store, user *user.User, token *file.Token, size int) *file.FileMetadata {
	s := RandStringBytesMaskImprSrcUnsafe(size)

	_, _, err := SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(s)), param)
	if err != nil {
		t.Fatal(err)
	}

	fv, fm, err := file.FindFile(db, token, param.FullPath)
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
	_, err = helpers.CreateToken(db, user, false, "some-name", CONFIG.MaxTokens)
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
func GetFileAndVersions(t *testing.T, db *gorm.DB, token *file.Token, fm *file.FileMetadata) file.FileMetadataAndVersion {
	fvv, err := file.ListVersionsFile(db, token, fm)
	if err != nil {
		t.Fatal(err)
	}
	return file.FileMetadataAndVersion{FileMetadata: fm, Versions: fvv}
}

func FileNotif(t *testing.T, db *gorm.DB, store *file.Store) {
	user, token := createUserAndToken(RandStringBytesMaskImprSrcUnsafe(10)+"@example.com", t, db, store)
	f := fp("/example/example.txt")
	v := uint64(10)
	f.AcceptableDelta = &v

	fm := uploadFile(f, t, db, store, user, token, 1000)
	fmv := GetFileAndVersions(t, db, token, fm)
	if len(fmv.Versions) != 1 {
		t.Fatalf("expected 2 got %d", len(fmv.Versions))
	}
	perFile, err := notifications.IgnoreAndMarkAlreadyNotified(db, notifications.ExecuteRule([]file.FileMetadataAndVersion{fmv}))
	if err != nil {
		t.Fatal(err)
	}
	if len(perFile) != 0 {
		t.Fatalf("expected 0 got %d", len(perFile))
	}

	uploadFile(f, t, db, store, user, token, 1005)
	fmv = GetFileAndVersions(t, db, token, fm)
	if len(fmv.Versions) != 2 {
		t.Fatalf("expected 2 got %d", len(fmv.Versions))
	}

	perFile, err = notifications.IgnoreAndMarkAlreadyNotified(db, notifications.ExecuteRule([]file.FileMetadataAndVersion{fmv}))
	if err != nil {
		t.Fatal(err)
	}
	if len(perFile) != 0 {
		t.Fatalf("expected 0 got %d", len(perFile))
	}

	uploadFile(f, t, db, store, user, token, 1120)
	fmv = GetFileAndVersions(t, db, token, fm)
	if len(fmv.Versions) != 3 {
		t.Fatalf("expected 3 got %d", len(fmv.Versions))
	}

	perFile, err = notifications.IgnoreAndMarkAlreadyNotified(db, notifications.ExecuteRule([]file.FileMetadataAndVersion{fmv}))
	if err != nil {
		t.Fatal(err)
	}
	if len(perFile) != 1 {
		t.Fatalf("expected 1 got %d", len(perFile))
	}

	perFile, err = notifications.IgnoreAndMarkAlreadyNotified(db, notifications.ExecuteRule([]file.FileMetadataAndVersion{fmv}))
	if err != nil {
		t.Fatal(err)
	}
	if len(perFile) != 0 {
		t.Fatalf("expected 0 got %d", len(perFile))
	}

	uploadFile(f, t, db, store, user, token, 0)
	fmv = GetFileAndVersions(t, db, token, fm)
	if len(fmv.Versions) != 4 {
		t.Fatalf("expected 4 got %d", len(fmv.Versions))
	}
	perFile, err = notifications.IgnoreAndMarkAlreadyNotified(db, notifications.ExecuteRule([]file.FileMetadataAndVersion{fmv}))
	if err != nil {
		t.Fatal(err)
	}
	if len(perFile) != 1 {
		t.Fatalf("expected 1 got %d", len(perFile))
	}

	status, err := helpers.GetUserStatus(db, user)
	if err != nil {
		t.Fatal(err)
	}

	perFile[0].Age = &common.AgeNotification{}
	perFile = append(perFile, perFile[0])
	log.Printf("%s", help.Render(help.HelpObject{
		Template:      help.EmailNotification,
		Email:         user.Email,
		Notifications: perFile,
		Status:        status,
	}))

	fm, err = helpers.StopNotifications(db, user, fmv.FileMetadata.ID)
	if err != nil {
		t.Fatal(err)
	}
	fmv = GetFileAndVersions(t, db, token, fm)

	perFile = notifications.ExecuteRule([]file.FileMetadataAndVersion{fmv})
	if len(perFile) != 0 {
		t.Fatalf("expected 0 got %d", len(perFile))
	}

}
func FileIO(t *testing.T, db *gorm.DB, store *file.Store) {
	for k := 0; k < 10; k++ {
		user, token := createUserAndToken(RandStringBytesMaskImprSrcUnsafe(10)+"@example.com", t, db, store)
		filePath := "/example/example.txt"
		var fmFirst *file.FileMetadata
		for i := 0; i < 20; i++ {
			fmFirst = uploadFile(fp(filePath), t, db, store, user, token, i)
		}

		fmSecond := uploadFile(fp(filePath+"second"), t, db, store, user, token, 1000)
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

		user.QuotaInode = 2
		if err := db.Save(user).Error; err != nil {
			t.Fatal(err)
		}

		_, _, err = SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d"))), fp(filePath+"second"))
		if err != nil {
			t.Fatal(err)
		}
		status, err := helpers.GetUserStatus(db, user)
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("%#v", status)
		if status.QuotaInodeUsed != 1 {
			t.Fatalf("expected 1 got %d", status.QuotaInodeUsed)
		}

		_, _, err = SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), fp(filePath+"second"))
		if err != nil {
			t.Fatal(err)
		}

		_, _, err = SaveFileProcess(store, db, user, token, bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), fp(filePath+"second"))
		if err.Error() != "inode quota limit reached" {
			t.Fatalf("expected inode quota limit reached got %s", err.Error())
		}

		CONFIG.MaxTokens = 10
		created := []*file.Token{}
		for i := 0; i < 9; i++ {
			to, err := helpers.CreateToken(db, user, false, "some-name", CONFIG.MaxTokens)
			if err != nil {
				t.Fatal(err)
			}
			created = append(created, to)
		}
		_, err = helpers.CreateToken(db, user, false, "some-other-name", CONFIG.MaxTokens)
		if err.Error() != "max tokens created (max=10)" {
			t.Fatalf("expected max tokens created (max=10) got %s", err.Error())
		}
		user.QuotaInode = 200
		user.Quota = 1
		if err := db.Save(user).Error; err != nil {
			t.Fatal(err)
		}

		_, _, err = SaveFileProcess(store, db, user, created[0], bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), file.FileParams{FullPath: filePath + "secondz"})
		if err.Error() != "quota limit reached" {
			t.Fatalf("expected quota limit reached got %s", err.Error())
		}

		user.QuotaInode = 200
		user.Quota = 1000
		if err := db.Save(user).Error; err != nil {
			t.Fatal(err)
		}

		_, _, err = SaveFileProcess(store, db, user, created[0], bytes.NewBuffer([]byte(fmt.Sprintf("a b c d e"))), file.FileParams{FullPath: filePath + "secondz"})
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

func TestEverything(t *testing.T) {
	store := setup()
	db, err := gorm.Open("postgres", "host=localhost user=baxx dbname=baxx password=baxx")
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)
	drop(db)
	init_db.InitDatabase(db)
	defer db.Close()
	FileNotif(t, db, store)
	//	FileIO(t, db, store)
}

func getUsed(t *testing.T, db *gorm.DB, user *user.User) uint64 {
	status, err := helpers.GetUserStatus(db, user)
	if err != nil {
		t.Fatal(err)
	}
	return status.QuotaUsed
}
