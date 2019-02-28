package main

import (
	"github.com/badoux/checkmail"
	"time"
)

type NotificationQueue struct {
	ID                        uint64 `gorm:"primary_key"`
	NotificationDestinationID uint64 `gorm:"not null"`
	Value                     string `gorm:"type:text";"not null"`
	Sent                      bool   `gorm:"not null"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type NotificationDestination struct {
	ID        uint64    `gorm:"primary_key"`
	ClientID  string    `gorm:"not null"`
	Type      string    `gorm:"not null"` // sms, email etc
	Value     string    `gorm:"not null"`
	Verified  bool      `gorm:"not null"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

/*
notify [x,y,z] when backups(matching pattern) are old
notify [x,y,z] when backups(matching pattern) are with weird delta% (can be 100%)
*/

type NotificationConfiguration struct {
	ID           uint64                    `gorm:"primary_key"`
	ClientID     string                    `gorm:"not null"`
	TokenID      string                    `gorm:"not null"`
	Destinations []NotificationDestination `gorm:"many2many:notification_destination_notification_configuration;"`

	// e.g.
	// match *.sql
	// type delta%
	// value -10

	// e.g.
	// match *.sql
	// type age
	// value (86400+3600) 1 day + 1 hour

	Match string
	Type  string
	Value int64

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func validateEmail(email string) error {
	err := checkmail.ValidateFormat(email)
	if err != nil {
		return err
	}

	//	err = checkmail.ValidateHost(email)
	//	if err != nil {
	//		return err
	//	}

	return nil
}
