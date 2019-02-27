package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/satori/go.uuid"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"path/filepath"
	"time"
)

/*

goal of the service is to help developers trust that they have backups
use some machine learning to predict backup size and report when abnormal

For who?

people like me(jackdoe) that have few machines and few databases and want
to know that their backups actually work

or some of my friends who always call me with their broken backups

Pricing

  * free [ this is how i would use it ]
    start the daemon
    configure it and you are good to go (setup sms, email and mysql)
    you can also make it upload to your s3 bucket
    create tokens for backup clients

    example flow:
      on backup server:
          install baxx
          baxx -conf /etc/baxx.conf

      the baxx daemon needs sql (pg, mysql, sqlite to store the metadata


    on other servers you simply do
      mysqldump | [encrypt -p passfile] | curl -F 'file=@-;filename=mysql.gz' -k https://baxIP.IP.IP.IP/api/v1/upload/clientID/tokenID
      (encrypt is optional, and you might want to ignore it, you might also want to have SSL properly)

    another example upload everything, only different files will be added
     find . -type f -exec url -f -F"file=@{}" https://baxx.dev/api/v1/upload/clientID/tokenID \;
    FIXME: find more efficient 1 liner

    - notifications
        + per directory
          . schecule [ when no new files are added in N hours ]
          . size [ when size does not change in N hours ]
          . when new files are smaller than old files


  * services

    * notifications only 0.99C per month
      still have local client but use baxx.dev for notifications
      this requires the baxx daemon to send metadata to baxx.dev
      tokens have to be created there as well (though they should be
      unique enough, so probably can just be added)
      config has to be uploaded as well (the notifications config)


    * storage + notification 0.99 + some buckets
      same as notification but you directly send the files to us and we upload it to s3
      costs same as notification plus s3 cost


  api:
    create client
    create token for client
    upload file in client/token
    list files in directory in client/token
    set config for token


Encryption:

* the files are supposed to be encrypted on input, so the service does
  not handle that


*/

type Client struct {
	ID        string `gorm:"primary_key"`
	Log       string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Token struct {
	ID        string `gorm:"primary_key"`
	ClientID  string `gorm:"not null"`
	Log       string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileMetadata struct {
	ID       string `gorm:"primary_key"`
	ClientID string `gorm:"not null"`
	TokenID  string `gorm:"not null"`
	Path     string `gorm:"not null"`
	Filename string `gorm:"not null"`
	Log      string `gorm:"type:text"`
	Size     uint64 `gorm:"not null"`
	SHA256   string `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (fm *FileMetadata) BeforeCreate(scope *gorm.Scope) error {
	id := uuid.Must(uuid.NewV4())
	scope.SetColumn("ID", fmt.Sprintf("%s", id))
	return nil
}

func (token *Token) BeforeCreate(scope *gorm.Scope) error {
	id := uuid.Must(uuid.NewV4())
	scope.SetColumn("ID", fmt.Sprintf("%s", id))
	return nil
}

func (client *Client) BeforeCreate(scope *gorm.Scope) error {
	id := uuid.Must(uuid.NewV4())
	scope.SetColumn("ID", fmt.Sprintf("%s", id))
	return nil
}

func locate(f string) string {
	return path.Join("/", "tmp", "baxx", f)
}
func extractLogFromRequest(req *http.Request) (string, error) {
	l, err := httputil.DumpRequest(req, false)
	return string(l), err
}
func main() {
	r := gin.Default()

	db, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)
	defer db.Close()

	db.AutoMigrate(&Client{}, &Token{}, &FileMetadata{})
	db.Model(&Token{}).AddIndex("idx_token_client_id", "client_id")
	db.Model(&FileMetadata{}).AddIndex("idx_fm_client_id_token_id_sha256", "client_id", "token_id", "sha256")
	db.Model(&FileMetadata{}).AddUniqueIndex("idx_fm_client_id_token_id_sha_name", "client_id", "token_id", "sha256", "path", "filename")

	r.POST("/v1/create/client", func(c *gin.Context) {
		rlog, err := extractLogFromRequest(c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := &Client{
			Log: rlog,
		}

		if err := db.Create(client).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, client)
	})

	r.POST("/v1/create/token/:client", func(c *gin.Context) {
		rlog, err := extractLogFromRequest(c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		client := &Client{}
		query := db.Where("id = ?", c.Param("client")).Take(client)
		if query.RecordNotFound() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client not found"})
			return
		}

		token := &Token{ClientID: client.ID, Log: rlog}
		if err := db.Create(token).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, token)
	})

	r.GET("/v1/download/:client/:token/*path", func(c *gin.Context) {
		client := c.Param("client")
		token := c.Param("token")

		name := filepath.Base(c.Param("path"))
		dir := filepath.Dir(c.Param("path"))
		fo := &FileMetadata{}
		if err := db.Where("client_id = ? AND token_id = ? AND filename = ? AND path = ?", client, token, name, dir).Take(fo).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+fo.Filename)
		c.Header("Content-Type", "application/octet-stream")
		c.File(locate(fo.ID))
	})

	// upload a bunch of files
	r.POST("/v1/upload/:client/:token/*path", func(c *gin.Context) {
		rlog, err := extractLogFromRequest(c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		client := c.Param("client")
		token := c.Param("token")

		query := db.Where("client_id = ? AND id = ?", client, token)
		if query.RecordNotFound() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client/token not found"})
			return
		}
		name := filepath.Base(c.Param("path"))
		dir := filepath.Dir(c.Param("path"))

		body := c.Request.Body
		defer body.Close()

		fo := &FileMetadata{}
		id := fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))
		insert := true
		db.Where("client_id = ? AND token_id = ? AND filename = ? AND path = ?", client, token, name, dir).Take(fo)
		if fo.ID != "" {
			id = fo.ID
			insert = false
		}

		sha := sha256.New()

		tee := io.TeeReader(body, sha)
		fn := locate(id)
		temporary := fmt.Sprintf("%s.%d.%s", locate(id), time.Now().UnixNano(), uuid.Must(uuid.NewV4()))
		dest, err := os.Create(temporary)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer dest.Close()

		size, err := io.Copy(dest, tee)
		if err != nil {
			os.Remove(fn)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		shasum := fmt.Sprintf("%x", sha.Sum(nil))
		tx := db.Begin()
		fo.ClientID = client
		fo.TokenID = token
		fo.SHA256 = shasum
		fo.Size = uint64(size)
		fo.Path = dir
		fo.Filename = name
		fo.ID = id
		fo.Log = rlog

		if insert {
			if err := tx.Save(fo).Error; err != nil {
				tx.Rollback()
				os.Remove(temporary)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else {
			if err := tx.Model(fo).Update(fo).Error; err != nil {
				tx.Rollback()
				os.Remove(temporary)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		err = os.Rename(temporary, fn)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, fo)
	})

	r.Run()
}
