package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"

	"github.com/jinzhu/gorm"
)

func SaveFileProcess(s *file.Store, db *gorm.DB, t *file.Token, body io.Reader, p string) (*file.FileVersion, *file.FileMetadata, error) {
	leftSize, leftInodes, err := file.GetQuotaLeft(db, t)
	if err != nil {
		return nil, nil, err
	}

	if leftSize < 0 {
		return nil, nil, errors.New("quota limit reached")
	}

	if leftInodes < 1 {
		return nil, nil, errors.New("inode quota limit reached")
	}

	return file.SaveFile(s, db, t, p, body)
}

func setupIO(srv *server) {
	r := srv.r
	store := srv.store
	db := srv.db
	getViewTokenLoggedOrNot := srv.getViewTokenLoggedOrNot
	download := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fv, _, err := file.FindFile(db, t, c.Param("path"))
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		}

		reader, err := store.DownloadFile(t.Salt, t.Bucket, fv.StoreID)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p := c.Param("path")
		fv, fm, err := SaveFileProcess(store, db, t, body, p)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// check if over quota

		actionLog(db, t.UserID, "file", "upload", c.Request, fmt.Sprintf("FileVersion: %d/%d", fv.ID, fv.FileMetadataID))
		if wantJson(c) {
			c.JSON(http.StatusOK, fv)
			return
		}
		c.String(http.StatusOK, file.FileLine(fm, fv))
	}

	deleteFile := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			n = 1
		}

		actionLog(db, t.UserID, "file", "delete", c.Request, "")
		c.JSON(http.StatusOK, &common.DeleteSuccess{Success: true, Count: n})
	}

	listFiles := func(c *gin.Context) {
		t, _, err := getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		p := c.Param("path")
		if !strings.HasSuffix(p, "/") {
			p = p + "/"
		}

		files, err := file.ListFilesInPath(db, t, p, false)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if wantJson(c) {
			c.JSON(http.StatusOK, files)
			return
		}
		c.String(http.StatusOK, file.LSAL(files))
	}

	mutateSinglePATH := "/io/:token/*path"
	r.GET(mutateSinglePATH, download)
	r.POST(mutateSinglePATH, upload)
	r.PUT(mutateSinglePATH, upload)
	r.DELETE(mutateSinglePATH, deleteFile)

	for _, a := range []string{"dir", "ls"} {
		lsPath := "/" + a + "/:token/*path"
		r.GET(lsPath, listFiles)
	}

}
