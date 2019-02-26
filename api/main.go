package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"mime/multipart"
	"time"
)

/*
   backup send to custom s3
   send data to to baxx.xyz
*/

func verify(client string) error {
	if client == "4c41a98a-915a-4a07-bf0c-901ee78bc594" {
		return nil
	}
	return errors.New("bad client")
}

type BackupsConfig struct {
	Client  string `json:"client"`
	Backups []struct {
		Pattern string `json:"pattern"`
		Alert   []struct {
			Sms   []string `json:"sms"`
			Email []string `json:"email"`
		} `json:"alert"`
		ExpectedSchedule string `json:"expectedSchedule"` // https://github.com/gorhill/cronexpr
		MinSize          int    `json:"minSize"`
		MaxDelta         int    `json:"maxDelta"`
	} `json:"backups"`
}

type FileMetadata struct {
	Path         string              `json:"path"`
	FileName     string              `json:"fileName"`
	UploadedAtNs uint64              `json:"uploadedAtNs"`
	Header       map[string][]string `json:"header"`
	Size         uint64              `json:"size"`
	ID           string              `json:"id"`
	SHA256       string              `json:"sha256"`
}

type ListOutput struct {
	Client string          `json:"client"`
	Files  []*FileMetadata `json:"files"`
}

func addFile(client string, path string, f multipart.FileHeader) (*FileMetadata, error) {
	file, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	u4 := uuid.Must(uuid.NewV4())
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return nil, err
	}
	sha := fmt.Sprintf("%x", h.Sum(nil))
	fo := &FileMetadata{
		Size:         uint64(f.Size),
		Header:       f.Header,
		Path:         path,
		Name:         f.Filename,
		SHA256:       sha,
		UploadedAtNs: time.Now().UnixNano(),
		ID:           fmt.Sprintf("%s-%s", sha, u4),
	}

	// store somewhere

	err := addToClientMetadata(client, fo)
	if err != nil {
		return nil, err
	}
	return fo, nil
}

func addToClientMetadata(client string, fo *FileMetadata) error {

	return nil
}

func main() {
	r := gin.Default()

	// setup alerting
	// e.g. notify sms if backup are missing (expected daily)
	// configure amount archives
	// curl -d '{"backups":[{"pattern":"*mysql*", "alert": [{"sms":["+123123"],"email":["jack@prymr.nl"]}], "expectedSchedule":"2 0 * * *","minSize": 512000, "maxDelta": 10000}]}' baxx.xyz/api/v1/config/4c41a98a-915a-4a07-bf0c-901ee78bc594

	// upload one or many files
	// cat mysql.gz | encrypt -p passfile | curl -F 'files[]=@-;filename=mysql.gz' baxx.xyz/api/v1/upload/4c41a98a-915a-4a07-bf0c-901ee78bc594/mysql/main_db
	// {"files": [{"path": "/mysql/main_db", "name":"mysql.gz", "uploadedAtMs": 1231231231, "size": 123}]}

	// list directory
	// curl baxx.xyz/api/v1/list/4c41a98a-915a-4a07-bf0c-901ee78bc594/mysql/main_db
	// {"files": [{"path": "/mysql/main_db", "name":"mysql.gz", "id" "123123", "uploadedAtMs": 1231231231, "size": 123}]}

	// download and decrypt a file
	// curl baxx.xyz/api/v1/download/4c41a98a-915a-4a07-bf0c-901ee78bc594/123123 | decrypt -p passfile > mysql.gz
	// ...
	//

	r.GET("/api/v1/config/:client", func(c *gin.Context) {
		client := c.Param("client")
		if err := verify(client); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		json := &Backups{}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, json)
	})

	// list all files in the backup
	r.GET("/api/v1/list/:client/*path", func(c *gin.Context) {
		client := c.Param("client")
		if err := verify(client); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOk, &ListOutput{})
	})

	// download specific file
	r.GET("/api/v1/download/:client/:id", func(c *gin.Context) {
		client := c.Param("client")
		token := c.Param("token")
		if err := verify(client); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//
	})

	// upload a bunch of files
	r.GET("/api/v1/upload/:client/*path", func(c *gin.Context) {
		client := c.Param("client")
		path := c.Param("path")
		if err := verify(client); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		out := &ListOutput{Files: []*FileMetadata{}}

		files := form.File["files[]"]
		for _, file := range files {
			fo, err := addToClientMetadata(client, path, file)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			out.Files = append(out.Files, fo)
		}

		// delete the temp files
		form.RemoveAll()

		c.JSON(http.StatusOk, out)
	})

	r.Run()
}
