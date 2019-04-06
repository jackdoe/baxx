package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"strings"

	"github.com/gin-gonic/gin"

	al "github.com/jackdoe/baxx/api/action_log"
	"github.com/jackdoe/baxx/api/file"
	"github.com/jackdoe/baxx/api/helpers"
	"github.com/jackdoe/baxx/api/user"
	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/help"

	"github.com/jinzhu/gorm"
)

func FileLine(fm *file.FileMetadata, fv *file.FileVersion) string {
	return fmt.Sprintf("%s %s %d %s\n", fm.FullPath(), fv.SHA256, fv.Size, fv.CreatedAt.Format(time.ANSIC))
}

func SaveFileProcess(s *file.Store, db *gorm.DB, u *user.User, t *file.Token, body io.Reader, fp file.FileParams) (*file.FileVersion, *file.FileMetadata, error) {
	status, err := helpers.GetUserStatus(db, u)
	if err != nil {
		return nil, nil, err
	}

	if status.QuotaUsed >= u.Quota {
		return nil, nil, errors.New("quota limit reached")
	}

	if status.QuotaInodeUsed >= u.QuotaInode {
		return nil, nil, errors.New("inode quota limit reached")
	}

	return file.SaveFile(s, db, t, body, fp)
}

func LSAL(files []file.FileMetadataAndVersion) string {
	buf := bytes.NewBufferString("")
	grouped := map[string][]file.FileMetadataAndVersion{}

	for _, f := range files {
		grouped[f.FileMetadata.Path] = append(grouped[f.FileMetadata.Path], f)
	}

	keys := []string{}
	for p := range grouped {
		keys = append(keys, p)
	}
	sort.Strings(keys)
	total := uint64(0)
	for _, k := range keys {
		files := grouped[k]

		size := uint64(0)
		for _, f := range files {
			for _, v := range f.Versions {
				size += v.Size
				total += v.Size
			}
		}

		for _, f := range files {
			if len(f.Versions) > 0 {
				buf.WriteString(FileLine(f.FileMetadata, f.Versions[len(f.Versions)-1]))
			}
		}
	}
	return buf.String()
}

func setupIO(srv *server) {
	store := srv.store
	db := srv.db
	getViewTokenLoggedOrNot := srv.getViewTokenLoggedOrNot
	download := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fv, fm, err := file.FindFile(db, t, c.Param("path"))
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Exec("UPDATE file_metadata SET count_read = count_read + 1 WHERE id = ?", fm.ID).Error; err != nil {
			// just warn, for whatever reason this might error
			// its better if we continue because the store might not be affected
			warnErr(c, err)
		}

		reader, err := store.DownloadFile(t.Salt, t.UUID, fv.StoreID)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// FIXME: close?
		c.Header("Content-Length", fmt.Sprintf("%d", fv.Size))

		c.Header("Content-Disposition", "attachment; filename="+fv.SHA256+".sha") // make sure people dont use it for loading js
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		c.DataFromReader(http.StatusOK, int64(fv.Size), "octet/stream", reader, map[string]string{})
	}

	upload := func(c *gin.Context) {
		body := c.Request.Body
		defer body.Close()

		t, u, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p := c.Param("path")
		var fileParams file.FileParams
		if err := c.Bind(&fileParams); err != nil {
			warnErr(c, err)
		}

		fileParams.FullPath = p
		fv, fm, err := SaveFileProcess(store, db, u, t, body, fileParams)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// check if over quota

		al.Log(db, t.UserID, "file", "upload", c.Request, fmt.Sprintf("FileVersion: %d/%d", fv.ID, fv.FileMetadataID))
		if wantJson(c) {
			c.IndentedJSON(http.StatusOK, fv)
			return
		}

		c.String(http.StatusOK, FileLine(fm, fv))
	}

	deleteFile := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		force := false
		recursive := false
		var json common.Force
		if err := c.ShouldBindJSON(&json); err == nil {
			if json.Force != nil {
				force = *json.Force
			}

			if json.Recursive != nil {
				recursive = *json.Recursive
			}
		}
		p := c.Param("path")
		n := 0

		if force {
			if err := file.DeleteFileWithPath(store, db, t, p); err == nil {
				n++
			}
			files, err := file.ListFilesInPath(db, t, p, !recursive)
			if err == nil {
				for _, f := range files {
					if err := file.DeleteFile(store, db, t, f.FileMetadata); err == nil {
						n++
					}
				}
			}
		} else {
			if err := file.DeleteFileWithPath(store, db, t, p); err != nil {
				warnErr(c, err)
				c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			n = 1
		}

		al.Log(db, t.UserID, "file", "delete", c.Request, "")
		c.IndentedJSON(http.StatusOK, &common.DeleteSuccess{Success: true, Count: n})
	}

	listFiles := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		p := c.Param("path")
		if !strings.HasSuffix(p, "/") {
			p = p + "/"
		}

		files, err := file.ListFilesInPath(db, t, p, false)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if wantJson(c) {
			c.IndentedJSON(http.StatusOK, files)
			return
		}
		c.String(http.StatusOK, LSAL(files))
	}

	mutateSinglePATH := "/io/:token/*path"
	srv.r.GET(mutateSinglePATH, download)
	srv.r.POST(mutateSinglePATH, upload)
	srv.r.PUT(mutateSinglePATH, upload)
	srv.r.DELETE(mutateSinglePATH, deleteFile)
	srv.r.GET("/ls/:token/*path", listFiles)

	srv.registerHelp(false, help.HelpObject{Template: help.FileMeta}, "/io", "/io/*path", "/ls", "/ls/*path")
}
