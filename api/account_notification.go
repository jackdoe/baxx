package main

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

type NotificationRule struct {
	ID     uint64 `gorm:"primary_key" json:"-"`
	UserID uint64 `gorm:"type:bigint not null REFERENCES users(id)" json:"-"`
	Name   string `gorm:"not null" json:"-"`
	Regex  string `gorm:"not null" json:"-"`
}

type EmailNotification struct {
	ID      uint64 `gorm:"primary_key" json:"-"`
	UserID  uint64 `gorm:"type:bigint not null REFERENCES users(id)" json:"-"`
	UUID    string `gorm:"type:varchar(255) not null unique" json:"-"`
	Example string `gorm:"not null;type:text" json:"-"`

	EmailText    string `gorm:"not null;type:text" json:"-"`
	EmailDest    string `gorm:"not null" json:"-"`
	EmailSubject string `gorm:"not null;type:text" json:"-"`

	BecauseOfRule uint64 `gorm:"type:bigint not null REFERENCES notification_rules(id)(" json:"-"`

	SentAt    time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

/*
#  same file with changes
example database.sql.gz
fnew version of the file:
+ is too small [table was deleted or truncated?]
+ too big [ wrong table is backed up?]
+ too old [ not backed up in a while ]
+ deleted [ bad script deletes everything ]

# content
such as photos
+ removed too many files [in case of wrong delete]
+ no backup in a while
+ directory was not updated in a while
+ daily/weekly update rate is weird


at glance there are 2 main scenarios, anomalies per file and per directory
so basic rule might look like
{
   "name": 'watch for old'
   "full path": 'regex',     // can be empty, by default is "all"
   "watch": [
      {
        "what": "age",
        "max": 10 * 24 * 3600 // 10 days
      },
      {
        "what": "count",
        "interval": 3600, //
        "delta": 10 * 24 * 3600 // 10 days
      },
   ]
}


*/
