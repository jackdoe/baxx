package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/jackdoe/baxx/api/file"
	"github.com/jackdoe/baxx/help"
)

func ShaDiff(db *gorm.DB, t *file.Token, body io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(body)
	out := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		splitted := strings.SplitN(line, "  ", 2)
		if len(splitted) != 2 || len(splitted[0]) != 64 || len(splitted[1]) == 0 {
			return nil, fmt.Errorf("expected 'shasum(64 chars)  path/to/file' (two spaces), basically output of shasum -a 256 file; got: %s", line)
		}
		_, _, err := file.FindFileBySHA(db, t, splitted[0])
		if err != nil {
			out = append(out, line)
		}
	}

	return out, nil
}

func setupSYNC(srv *server) {
	db := srv.db

	srv.r.GET("/sync/sha256/:token/:sha256", func(c *gin.Context) {
		t, _, err := srv.getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fv, fm, err := file.FindFileBySHA(db, t, c.Param("sha256"))
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if wantJson(c) {
			c.IndentedJSON(http.StatusOK, gin.H{"sha": fv.SHA256, "path": fm.Path, "name": fm.Filename})
			return
		}

		c.String(http.StatusOK, FileLine(fm, fv))
	})

	// lookup on many sha and return only the ones that are not found
	// expect input from:
	// find . -type f | grep -v .git | xargs -P4 -I '{}' shasum -a 256 {}
	// we want to have endpoint that is easy to hook to find . -type f | grep -v .git | xargs -P4 -I '{}' shasum -a 256 {} | curl -d@- https...
	srv.r.POST("/sync/sha256/:token", func(c *gin.Context) {
		t, _, err := srv.getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		body := c.Request.Body
		defer body.Close()
		out, err := ShaDiff(db, t, body)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		c.String(http.StatusOK, strings.Join(out, "\n")+"\n")
	})
	srv.registerHelp(false, help.HelpObject{Template: help.SyncMeta}, "/sync", "/sync/*path")
}
