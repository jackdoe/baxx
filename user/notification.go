package user

import (
	"time"
)

type Feature struct {
	Value float32
	Name  string
}

type Namespace struct {
	Name     string
	Features []*Feature
}

type Example struct {
	Prediction float32
	Truth      float32
	Namespaces []*Namespace
}

type EmailNotification struct {
	ID uint64 `gorm:"primary_key" json:"-"`

	Example string `gorm:"not null;type:text" json:"-"`

	EmailText    string `gorm:"not null;type:text" json:"-"`
	EmailDest    string `gorm:"not null" json:"-"`
	EmailSubject string `gorm:"not null;type:text" json:"-"`

	BecauseOfRule uint64 `gorm:"not null" json:"-"`

	SentAt    time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type NotificationRule struct {
	ID     uint64 `gorm:"primary_key" json:"-"`
	UserID uint64 `gorm:"not null" json:"-"`
	Name   string `gorm:"not null" json:"-"`
}
