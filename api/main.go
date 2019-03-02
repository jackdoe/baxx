package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	. "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
	. "github.com/jackdoe/baxx/user"
	auth "github.com/jackdoe/gin-basic-auth-dynamic"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"os"
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
	var pbind = flag.String("bind", "127.0.0.1:9123", "bind")
	var proot = flag.String("root", "/tmp", "root")
	flag.Parse()

	dbType := os.Getenv("BAXX_DB")
	dbURL := os.Getenv("BAXX_DB_URL")
	debug := true
	ROOT = *proot
	if dbType == "" {
		dbType = "sqlite3"
		dbURL = "/tmp/gorm.db"
	} else {
		debug = false
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := gorm.Open(dbType, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(debug)
	defer db.Close()

	initDatabase(db)

	r := gin.Default()

	authorized := r.Group("/protected")
	authorized.Use(auth.BasicAuth(func(context *gin.Context, realm, user, pass string) auth.AuthResult {
		u, _, err := FindUser(db, user, pass)
		if err != nil {
			return auth.AuthResult{Success: false, Text: `{"error":"not authorized"}`}
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
		if err := ValidatePassword(json.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := ValidateEmail(json.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// if the user is created again with current password just return it
		u, exists, err := FindUser(db, json.Email, json.Password)

		if err == nil {
			c.JSON(http.StatusOK, &CreateUserOutput{Secret: u.SemiSecretID, TokenWO: "", TokenRW: ""})
			return
		}

		if exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
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

		c.JSON(http.StatusOK, &CreateUserOutput{
			Secret:  user.SemiSecretID,
			TokenWO: tokenWO.ID,
			TokenRW: tokenRW.ID,
			Help:    help.AfterRegistration(user.SemiSecretID, tokenRW.ID, tokenWO.ID),
		})
	})

	authorized.POST("/v1/replace/secret", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		user.SetSemiSecretID()
		if err := db.Save(user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &ChangeSecretOutput{
			Secret: user.SemiSecretID,
		})
	})

	authorized.POST("/v1/replace/password", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		var json ChangePasswordInput
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err != ValidatePassword(json.NewPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user.SetPassword(json.NewPassword)
		if err := db.Save(user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &Success{true})
	})

	authorized.POST("/v1/replace/email", func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		var json ChangeEmailInput
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err != ValidateEmail(json.NewEmail) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.NewEmail != user.Email {
			user.Email = json.NewEmail
			user.EmailVerified = nil
		}

		if err := db.Save(user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, &Success{true})
	})

	authorized.POST("/v1/create/token", func(c *gin.Context) {
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
			if user != x.(*User).SemiSecretID {
				return nil, errors.New("wrong token/user combination")
			}
		}

		token := c.Param("token")

		t, err := FindToken(db, user, token)
		if err != nil {
			return nil, err
		}

		if !isLoggedIn {
			if t.WriteOnly && c.Request.Method != "POST" {
				return nil, errors.New("write only token, use /v1/protected/io/:secret/:token/*path")
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

	upload := func(c *gin.Context) {
		t, err := getViewTokenLoggedOrNot(c)
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
	}

	mutatePATH := "/v1/io/:user_semi_secret_id/:token/*path"

	authorized.GET(mutatePATH, download)
	r.GET(mutatePATH, download)

	authorized.POST(mutatePATH, upload)
	r.POST(mutatePATH, upload)

	r.Run(*pbind)
}
