package main

import (
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"

	"github.com/badoux/checkmail"
	"github.com/gin-gonic/gin"
	"github.com/jackdoe/baxx/notification"
	"github.com/jackdoe/baxx/user"
	log "github.com/sirupsen/logrus"
)

func ValidateEmail(email string) error {
	err := checkmail.ValidateFormat(email)
	if err != nil {
		return fmt.Errorf("invalid email address (%s)", err.Error())
	}
	return nil
}

var ERR_BAD_PASSWORD = fmt.Errorf("password is too short, refer to https://www.xkcd.com/936/")

func ValidatePassword(p string) error {
	if len(p) < 8 {
		return ERR_BAD_PASSWORD
	}
	return nil
}
func BasicAuthDecode(c *gin.Context) (string, string) {
	auth := strings.SplitN(c.GetHeader("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		return "", ""
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		return "", ""
	}
	return pair[0], pair[1]
}

func wantJson(c *gin.Context) bool {
	format := c.DefaultQuery("format", "text")
	return format == "json"
}

func warnErr(c *gin.Context, err error) {
	x, isLoggedIn := c.Get("user")
	u := &user.User{}
	if isLoggedIn {
		u = x.(*user.User)
	}
	_, fn, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf("uid: %d, uri: %s, err: >> %s << [%s:%d]", u.ID, c.Request.RequestURI, err.Error(), fn, line)
	notification.SendSlackDefault("warning", msg)
	log.Warn(msg)
}
