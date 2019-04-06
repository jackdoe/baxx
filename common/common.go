package common

import (
	"os"
	"time"
)

type CreateTokenInput struct {
	WriteOnly bool   `json:"write_only"`
	Name      string `json:"name"`
}

type CreateUserInput struct {
	Email    string `binding:"required" json:"email"`
	Password string `binding:"required" json:"password"`
}

type ChangeEmailInput struct {
	NewEmail string `binding:"required" json:"new_email"`
}

type ChangePasswordInput struct {
	NewPassword string `binding:"required" json:"new_password"`
}

type DeleteTokenInput struct {
	UUID string `binding:"required" json:"token"`
}

type Force struct {
	Force     *bool `json:"force"`
	Recursive *bool `json:"recursive"`
}

type ModifyTokenInput struct {
	WriteOnly *bool  `json:"write_only"`
	Name      string `json:"name"`
	UUID      string `binding:"required" json:"token"`
}

type Success struct {
	Success bool `json:"success"`
}

type DeleteSuccess struct {
	Success bool `json:"success"`
	Count   int  `json:"deleted"`
}

type TokenOutput struct {
	UUID       string    `json:"token"`
	ID         uint64    `json:"id"`
	WriteOnly  bool      `json:"write_only"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	SizeUsed   uint64    `json:"size_used"`
	InodesUsed uint64    `json:"inodes_used"`
}

type UserStatusOutput struct {
	EmailVerified         *time.Time     `json:"email_verified"`
	Paid                  bool           `json:"paid"`
	StartedSubscription   *time.Time     `json:"started_subscription"`
	CancelledSubscription *time.Time     `json:"cancelled_subscription"`
	PaymentID             string         `json:"payment_id"`
	SubscribeURL          string         `json:"subscribe_url"`
	CancelSubscriptionURL string         `json:"cancel_subscription_url"`
	LastVerificationID    string         `json:"-"`
	UserID                uint64         `json:"-"`
	Tokens                []*TokenOutput `json:"tokens"`
	Email                 string         `json:"email"`
	Quota                 uint64         `json:"quota"`
	QuotaUsed             uint64         `json:"quota_used"`
	QuotaInode            uint64         `json:"quota_inodes"`
	QuotaInodeUsed        uint64         `json:"quota_inodes_used"`
}

type QueryError struct {
	Error string `json:"error"`
}

var EMPTY_STATUS = &UserStatusOutput{
	PaymentID: "WILL-BE-IN-YOUR-EMAIL",
	Email:     "your.email@example.com",
	Tokens: []*TokenOutput{
		&TokenOutput{
			ID:        0,
			UUID:      "TOKEN-UUID-A",
			WriteOnly: false,
			Name:      "db-example-a",
		},
		&TokenOutput{
			ID:        0,
			UUID:      "TOKEN-UUID-B",
			WriteOnly: false,
			Name:      "content-example-b",
		},
	},
}

type LocalFile struct {
	File           *os.File
	SHA            string
	Size           uint64
	TempName       string
	OriginFullPath string
}

type AgeNotification struct {
	ActualAge time.Duration
	Overdue   time.Duration
}

type SizeNotification struct {
	PreviousSize uint64
	Delta        float64
}

type FileNotification struct {
	Age             *AgeNotification
	Size            *SizeNotification
	FullPath        string
	LastVersionSize uint64
	FileVersionID   uint64
	FileMetadataID  uint64
	CreatedAt       time.Time
}
