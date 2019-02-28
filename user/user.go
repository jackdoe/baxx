package user

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"time"
)

type User struct {
	ID string `gorm:"primary_key"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	id := uuid.Must(uuid.NewV4())
	scope.SetColumn("ID", fmt.Sprintf("%s", id))
	return nil
}

type NotificationDest struct {
	Type  string `binding:"required"`
	Value string `binding:"required"`
}

type NotificationRule struct {
	Match        string `binding:"required"`
	Type         string `binding:"required"`
	Value        int64
	Destinations []NotificationDest
}

type NotificationConfiguration struct {
	Rules []*NotificationRule `binding:"required"`
}
type Token struct {
	ID                 string    `gorm:"primary_key"`
	Salt               string    `gorm:"not null";type:varchar(32)`
	UserID             string    `gorm:"not null"`
	WriteOnly          bool      `gorm:"not null"`
	NumberOfArchives   uint64    `gorm:"not null"`
	NotificationConfig string    `gorm:"not null";type:text`
	CreatedAt          time.Time `json:"-"`
	UpdatedAt          time.Time `json:"-"`
}

func (token *Token) getNotificationConfig() (*NotificationConfiguration, error) {
	nc := &NotificationConfiguration{}
	err := json.Unmarshal([]byte(token.NotificationConfig), nc)
	if err != nil {
		return nil, err
	}
	return nc, nil
}

func (token *Token) setNotificationConfig(nc *NotificationConfiguration) error {
	b, err := json.Marshal(nc)
	if err != nil {
		return err
	}
	token.NotificationConfig = string(b)
	return nil
}

func (token *Token) BeforeCreate(scope *gorm.Scope) error {
	id := uuid.Must(uuid.NewV4())

	scope.SetColumn("ID", fmt.Sprintf("%s", id))
	return nil
}

func FindToken(db *gorm.DB, user string, token string) (*Token, error) {
	t := &Token{}
	query := db.Where("user_id = ? AND id = ?", user, token).Take(t)
	if query.RecordNotFound() {
		return nil, query.Error
	}
	return t, nil
}
