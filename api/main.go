package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type CreateTokenInput struct {
	WriteOnly        bool
	NumberOfArchives uint64
}

type CreateDestinationInput struct {
	Type  string `binding:"required"`
	Value string `binding:"required"`
}

type CreateNotificationInput struct {
	Match        string `binding:"required"`
	Type         string `binding:"required"`
	Value        int64  `binding:"required"`
	Destinations []struct {
		Type  string
		Value string
	} `binding:"required"`
}

func initDatabase(db *gorm.DB) {
	if err := db.AutoMigrate(&Client{}, &Token{}, &FileOrigin{}, &FileMetadata{}, &FileVersion{}, &ActionLog{}, &NotificationDestination{}, &NotificationConfiguration{}, &NotificationQueue{}).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&Token{}).AddIndex("idx_token_client_id", "client_id").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&NotificationDestination{}).AddIndex("idx_nd_client_id", "client_id").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&NotificationDestination{}).AddUniqueIndex("idx_nd_client_id_type_value", "client_id", "type", "value").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&NotificationConfiguration{}).AddIndex("idx_nd_client_id_token_id", "client_id", "token_id").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&FileOrigin{}).AddUniqueIndex("idx_sha", "sha256").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&FileMetadata{}).AddUniqueIndex("idx_fm_client_id_token_id_path", "client_id", "token_id", "path", "filename").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&FileVersion{}).AddUniqueIndex("idx_fv_version_origin", "file_metadata_id", "file_origin_id").Error; err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := gin.Default()

	db, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)
	defer db.Close()

	initDatabase(db)
	store := &Store{db: db}
	r.POST("/v1/create/client", func(c *gin.Context) {
		client := &Client{}

		if err := db.Create(client).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		actionLog(db, client.ID, "client", "create", c.Request)
		c.JSON(http.StatusOK, client)
	})

	r.POST("/v1/create/destination/:client", func(c *gin.Context) {
		client := &Client{}
		query := db.Where("id = ?", c.Param("client")).Take(client)
		if query.RecordNotFound() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client not found"})
			return
		}

		var json CreateDestinationInput
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.Type != "email" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only email is supported"})
			return
		}
		if json.Type == "email" {
			err = validateEmail(json.Value)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		nd := &NotificationDestination{ClientID: client.ID, Type: json.Type, Value: json.Value}
		if err := db.Create(nd).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, client.ID, "destination", "create", c.Request)
		c.JSON(http.StatusOK, nd)
	})

	r.POST("/v1/create/notification/:client/:token", func(c *gin.Context) {
		client := c.Param("client")
		token := c.Param("token")
		_, err := store.FindToken(client, token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var json CreateNotificationInput
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(json.Destinations) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "destinations are requred"})
			return
		}

		if json.Type != "delta%" && json.Type != "age" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only 'delta%' and 'age' is supported"})
			return
		}

		if json.Type != "delta%" && json.Type != "age" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only 'delta%' and 'age' is supported"})
			return
		}

		_, err = regexp.Compile(json.Match)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx := db.Begin()

		nc := &NotificationConfiguration{ClientID: client, TokenID: token, Type: json.Type, Value: json.Value, Match: json.Match}
		if err := tx.Create(nc).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, d := range json.Destinations {
			nd := &NotificationDestination{}
			if err := tx.Where("client_id = ? AND type = ? AND value = ?", client, d.Type, d.Value).Take(nd).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := tx.Model(nc).Association("Destinations").Append(nd).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, client, "notification", "create", c.Request)
		c.JSON(http.StatusOK, nc)
	})

	r.POST("/v1/create/token/:client", func(c *gin.Context) {
		client := &Client{}
		query := db.Where("id = ?", c.Param("client")).Take(client)
		if query.RecordNotFound() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client not found"})
			return
		}

		var json CreateTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token := &Token{
			ClientID:         client.ID,
			Salt:             strings.Replace(fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), "-", "", -1),
			NumberOfArchives: json.NumberOfArchives,
			WriteOnly:        json.WriteOnly,
		}
		if err := db.Create(token).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		actionLog(db, client.ID, "token", "create", c.Request)
		c.JSON(http.StatusOK, token)
	})

	r.GET("/v1/download/:client/:token/*path", func(c *gin.Context) {
		client := c.Param("client")
		token := c.Param("token")
		dir, name := split(c.Param("path"))

		t, err := store.FindToken(client, token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fo, err := store.FindFile(t, dir, name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		file, err := os.Open(locate(fo.SHA256))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer file.Close()
		reader, err := decompressAndDecrypt(t.Salt, file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.DataFromReader(http.StatusOK, int64(fo.Size), "octet/stream", reader, map[string]string{})
	})

	r.POST("/v1/upload/:client/:token/*path", func(c *gin.Context) {
		client := c.Param("client")
		token := c.Param("token")
		t, err := store.FindToken(client, token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		dir, name := split(c.Param("path"))
		body := c.Request.Body
		defer body.Close()
		fv, err := store.SaveFile(t, body, dir, name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, client, "file", "upload", c.Request, fmt.Sprintf("FileVersion: %d/%d/%d", fv.ID, fv.FileMetadataID, fv.FileOriginID))
		c.JSON(http.StatusOK, fv)
	})

	r.Run()
}
