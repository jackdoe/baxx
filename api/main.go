package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	. "github.com/jackdoe/baxx/common"
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
			c.JSON(http.StatusOK, &CreateUserOutput{Secret: u.SemiSecretID, TokenWO: "", TokenRW: ""})
			return
		}

		user := &User{Email: json.Email}
		user.SetPassword(json.Password)
		if err := db.Create(user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		actionLog(db, user.ID, "user", "create", c.Request)

		tokenWO := &Token{
			UserID:           user.ID,
			Salt:             strings.Replace(fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), "-", "", -1),
			NumberOfArchives: 7,
			WriteOnly:        true,
		}
		if err := db.Create(tokenWO).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tokenRW := &Token{
			UserID:           user.ID,
			Salt:             strings.Replace(fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), "-", "", -1),
			NumberOfArchives: 7,
			WriteOnly:        false,
		}
		if err := db.Create(tokenRW).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &CreateUserOutput{Secret: user.SemiSecretID, TokenWO: tokenWO.ID, TokenRW: tokenRW.ID})
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

	getViewTokenLoggedOrNot := func(c *gin.Context) (*Token, error) {
		user := c.Param("user_semi_secret_id")
		x, isLoggedIn := c.Get("user")
		if isLoggedIn {
			user = x.(*User).SemiSecretID
		}

		token := c.Param("token")

		t, err := FindToken(db, user, token)
		if err != nil {
			return nil, err
		}

		if !isLoggedIn {
			if t.WriteOnly {
				return nil, errors.New("write only token, use /v1/protected/{list,download}/:secret/:token/*path")
			}
		}
		return t, nil

	}
	download := func(c *gin.Context) {
		t, err := getViewTokenLoggedOrNot(c)
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
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		c.DataFromReader(http.StatusOK, int64(fo.Size), "octet/stream", reader, map[string]string{})
	}

	authorized.GET("/download/:user_semi_secret_id/:token/*path", download)
	r.GET("/v1/download/:user_semi_secret_id/:token/*path", download)

	r.POST("/v1/upload/:user_semi_secret_id/:token/*path", func(c *gin.Context) {
		user := c.Param("user_semi_secret_id")
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
