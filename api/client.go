package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"time"
)

type Client struct {
	ID string `gorm:"primary_key"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (client *Client) BeforeCreate(scope *gorm.Scope) error {
	id := uuid.Must(uuid.NewV4())
	scope.SetColumn("ID", fmt.Sprintf("%s", id))
	return nil
}

type Token struct {
	ID               string    `gorm:"primary_key"`
	Salt             string    `gorm:"not null";type:varchar(32)`
	ClientID         string    `gorm:"not null"`
	WriteOnly        bool      `gorm:"not null"`
	NumberOfArchives uint64    `gorm:"not null"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

func (token *Token) BeforeCreate(scope *gorm.Scope) error {
	id := uuid.Must(uuid.NewV4())

	scope.SetColumn("ID", fmt.Sprintf("%s", id))
	return nil
}
