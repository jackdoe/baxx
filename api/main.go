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

      the baxx daemon needs sql (pg, mysql, sqlite) to store the metadata


    on other servers you simply do
      mysqldump | [encrypt -p passfile] | curl -' -k https://baxIP.IP.IP.IP/v1/upload/$CLIENT/$TOKEN/mysql.gz
      (encrypt is optional, and you might want to ignore it, you might also want to have SSL properly)

    another example upload everything, only different files will be added
     find . -type f -exec curl --binary-data @{} https://baxx.dev/v1/upload/$CLIENT/$TOKEN/{} \;

     for i in $(find . -type f); do
        curl -f https://baxx.dev/v1/diff/$CLIENT/$TOKEN/$(shasum $i | cut -f 1 -d ' ') \
        && curl --binary-data @$i https://baxx.dev/v1/upload/$CLIENT/$TOKEN
     done

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

type FileOrigin struct {
	ID        uint64 `gorm:"primary_key"`
	Size      uint64 `gorm:"not null"`
	SHA256    string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileMetadata struct {
	ID uint64 `gorm:"primary_key"`

	ClientID string `gorm:"not null"`
	TokenID  string `gorm:"not null"`
	Path     string `gorm:"not null"`
	Filename string `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileVersion struct {
	ID             uint64    `gorm:"primary_key"`
	FileMetadataID uint64    `gorm:"not null" json:"-"`
	Log            string    `gorm:"type:text" json:"-"`
	FileOriginID   uint64    `gorm:"not null" json:"-"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

func split(s string) (string, string) {
	s = filepath.Clean(s)
	name := filepath.Base(s)
	dir := filepath.Dir(s)
	return dir, name
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
	dir := path.Join("/", "tmp", "baxx")
	return path.Join(dir, f)
}

func extractLogFromRequest(req *http.Request) (string, error) {
	l, err := httputil.DumpRequest(req, false)
	return string(l), err
}

func saveUploadedFile(body io.Reader) (string, int64, error) {
	sha := sha256.New()
	tee := io.TeeReader(body, sha)

	temporary := locate(fmt.Sprintf("%d.%s", time.Now().UnixNano(), uuid.Must(uuid.NewV4())))
	dest, err := os.Create(temporary)
	if err != nil {
		return "", 0, err
	}

	size, err := io.Copy(dest, tee)
	if err != nil {
		dest.Close()
		os.Remove(temporary)
		return "", 0, err
	}
	dest.Close()

	shasum := fmt.Sprintf("%x", sha.Sum(nil))
	err = os.Rename(temporary, locate(shasum))
	if err != nil {
		os.Remove(temporary)
		return "", 0, err
	}
	return shasum, size, nil
}

func main() {
	r := gin.Default()

	db, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)
	defer db.Close()

	db.AutoMigrate(&Client{}, &Token{}, &FileOrigin{}, &FileMetadata{}, &FileVersion{})
	db.Model(&Token{}).AddIndex("idx_token_client_id", "client_id")

	db.Model(&FileOrigin{}).AddUniqueIndex("idx_sha", "sha256")

	db.Model(&FileMetadata{}).AddUniqueIndex("idx_fm_client_id_token_id_path", "client_id", "token_id", "path", "filename")
	db.Model(&FileVersion{}).AddUniqueIndex("idx_fv_version_origin", "file_metadata_id", "file_origin_id")

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
		dir, name := split(c.Param("path"))

		fm := &FileMetadata{}
		if err := db.Where("client_id = ? AND token_id = ? AND filename = ? AND path = ?", client, token, name, dir).Take(fm).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fv := &FileVersion{}
		if err := db.Where("file_metadata_id = ?", fm.ID).Last(fv).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fo := &FileOrigin{}
		if err := db.Where("id = ?", fv.FileOriginID).Take(fo).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+fm.Filename)
		c.Header("Content-Type", "application/octet-stream")
		c.File(locate(fo.SHA256))
	})

	r.POST("/v1/upload/:client/:token/*path", func(c *gin.Context) {
		client := c.Param("client")
		token := c.Param("token")
		query := db.Where("client_id = ? AND id = ?", client, token)
		if query.RecordNotFound() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client/token not found"})
			return
		}

		rlog, err := extractLogFromRequest(c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		dir, name := split(c.Param("path"))
		body := c.Request.Body
		defer body.Close()

		sha, size, err := saveUploadedFile(body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// create file origin
		fo := &FileOrigin{}
		tx := db.Begin()
		res := tx.Where("sha256 = ?", sha).Take(fo)
		if res.RecordNotFound() {
			// create new one
			fo.SHA256 = sha
			fo.Size = uint64(size)
			if err := tx.Save(fo).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

		}

		// create file metadata
		fm := &FileMetadata{}
		if err := tx.FirstOrCreate(&fm, FileMetadata{ClientID: client, TokenID: token, Path: dir, Filename: name}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// create the version
		fv := &FileVersion{}
		if err := tx.Where(FileVersion{FileMetadataID: fm.ID, FileOriginID: fo.ID}).Attrs(FileVersion{Log: rlog}).FirstOrCreate(&fv).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// goooo
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// FIXME: can create file and then nobody points to it
		// in case of rollback

		c.JSON(http.StatusOK, fv)
	})

	r.Run()
}
