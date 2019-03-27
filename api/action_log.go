package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type ActionLog struct {
	ID         uint64 `gorm:"primary_key"`
	UserID     uint64 `gorm:"not null"`
	ActionType string `gorm:"not null"`
	Action     string `gorm:"not null"`
	Log        string `gorm:"type:text"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func actionLog(db *gorm.DB, user uint64, actionType, action string, req *http.Request, extra ...string) {
	rlog, _ := extractLogFromRequest(req)
	al := &ActionLog{
		UserID:     user,
		ActionType: actionType,
		Action:     action,
		Log:        fmt.Sprintf("%s\n%s", rlog, strings.Join(extra, "\n")),
	}
	db.Create(al)
}

func extractLogFromRequest(req *http.Request) (string, error) {
	return fmt.Sprintf("X-Forwarded-For: %s\nRemoteAddr: %s\nURL: %s\nUser-Agent: %s", req.Header.Get("X-Forwarded-For"), req.RemoteAddr, req.URL, req.Header.Get("User-Agent")), nil
}
