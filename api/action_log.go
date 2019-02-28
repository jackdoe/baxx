package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"net/http"
	"net/http/httputil"
	//	"strings"
	"time"
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
	return
	// rlog, _ := extractLogFromRequest(req)

	// al := &ActionLog{
	// 	UserID:     user,
	// 	ActionType: actionType,
	// 	Action:     action,
	// 	Log:        fmt.Sprintf("%s\n%s", rlog, strings.Join(extra, "\n")),
	// }
	// db.Create(al)
}

func extractLogFromRequest(req *http.Request) (string, error) {
	l, err := httputil.DumpRequest(req, false)
	return fmt.Sprintf("%sRemoteAddr: %s\nURL: %s", string(l), req.RemoteAddr, req.URL), err
}
