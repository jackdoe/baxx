package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	. "github.com/jackdoe/baxx/file"
	. "github.com/jackdoe/baxx/user"
	auth "github.com/jackdoe/gin-basic-auth-dynamic"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"strings"
)

type CreateTokenInput struct {
	WriteOnly        bool   `json:"writeonly"`
	NumberOfArchives uint64 `json:"number_of_archives"`
}

type CreateUserInput struct {
	Email    string `binding:"required" json:"email"`
	Password string `binding:"required" json:"password"`
}

func initDatabase(db *gorm.DB) {
	if err := db.AutoMigrate(&User{}, &Token{}, &FileOrigin{}, &FileMetadata{}, &FileVersion{}, &ActionLog{}).Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&User{}).AddUniqueIndex("idx_email", "email").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&Token{}).AddIndex("idx_token_user_id", "user_id").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&FileOrigin{}).AddUniqueIndex("idx_sha", "sha256").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&FileMetadata{}).AddUniqueIndex("idx_fm_user_id_token_id_path", "user_id", "token_id", "path", "filename").Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Model(&FileVersion{}).AddUniqueIndex("idx_fv_metadata_origin", "file_metadata_id", "file_origin_id").Error; err != nil {
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

	authorized := r.Group("/v1/protected")
	authorized.Use(auth.BasicAuth(func(context *gin.Context, realm, user, pass string) auth.AuthResult {
		u, err := FindUser(db, user, pass)
		if err != nil {
			return auth.AuthResult{Success: false, Text: "not authorized"}
		}
		context.Set("user", u)
		return auth.AuthResult{Success: true}
	}))

	r.POST("/v1/register", func(c *gin.Context) {
		var json CreateUserInput
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// if the user is created again with current password just return it
		u, err := FindUser(db, json.Email, json.Password)
		if err == nil {
			c.JSON(http.StatusOK, u)
			return
		}

		user := &User{Email: json.Email}
		user.SetPassword(json.Password)
		if err := db.Create(user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		actionLog(db, user.ID, "user", "create", c.Request)
		c.JSON(http.StatusOK, user)
	})

	authorized.POST("/create/token", func(c *gin.Context) {
		user := c.MustGet("user").(*User)

		var json CreateTokenInput
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token := &Token{
			UserID:           user.ID,
			Salt:             strings.Replace(fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), "-", "", -1),
			NumberOfArchives: json.NumberOfArchives,
			WriteOnly:        json.WriteOnly,
		}

		//		if err := token.setNotificationConfig(json.NotificationConfig); err != nil {
		//			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		//			return
		//		}

		if err := db.Create(token).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, user.ID, "token", "create", c.Request)
		c.JSON(http.StatusOK, token)
	})

	r.GET("/v1/download/:user_semi_secret_id/:token/*path", func(c *gin.Context) {
		user := c.Param("user_semi_secret_id")
		token := c.Param("token")

		t, err := FindToken(db, user, token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fo, file, reader, err := FindAndOpenFile(db, t, c.Param("path"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer file.Close()
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		c.DataFromReader(http.StatusOK, int64(fo.Size), "octet/stream", reader, map[string]string{})
	})

	r.POST("/v1/upload/:user/:token/*path", func(c *gin.Context) {
		user := c.Param("user")
		token := c.Param("token")
		t, err := FindToken(db, user, token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		body := c.Request.Body
		defer body.Close()
		fv, err := SaveFile(db, t, body, c.Param("path"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		actionLog(db, t.UserID, "file", "upload", c.Request, fmt.Sprintf("FileVersion: %d/%d/%d", fv.ID, fv.FileMetadataID, fv.FileOriginID))
		c.JSON(http.StatusOK, fv)
	})

	r.Run()
}
