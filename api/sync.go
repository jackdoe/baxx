package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackdoe/baxx/file"
)

func setupSYNC(srv *server) {
	r := srv.r
	db := srv.db

	r.GET("/sync/sha256/:token/:sha256", func(c *gin.Context) {
		t, _, err := srv.getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fv, fm, err := file.FindFileBySHA(db, t, c.Param("sha256"))
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if wantJson(c) {
			c.JSON(http.StatusOK, gin.H{"sha": fv.SHA256, "path": fm.Path, "name": fm.Filename})
			return
		}

		c.String(http.StatusOK, file.FileLine(fm, fv))
	})

	// lookup on many sha and return only the ones that are not found
	// expect input from:
	// find . -type f | grep -v .git | xargs -P4 -I '{}' shasum -a 256 {}
	// we want to have endpoint that is easy to hook to find . -type f | grep -v .git | xargs -P4 -I '{}' shasum -a 256 {} | curl -d@- https...
	r.POST("/sync/sha256/:token", func(c *gin.Context) {
		t, _, err := srv.getViewTokenLoggedOrNot(c)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		body := c.Request.Body
		defer body.Close()
		out, err := file.ShaDiff(db, t, body)
		if err != nil {
			warnErr(c, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		c.String(http.StatusOK, strings.Join(out, "\n")+"\n")
	})

}
