package common

import (
	"os"
	"time"
)

type CreateTokenInput struct {
	WriteOnly        bool   `json:"write_only"`
	NumberOfArchives uint64 `json:"keep_n_versions"`
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
	UUID string `binding:"required" json:"uuid"`
}

type Success struct {
	Success bool `json:"success"`
}

type TokenOutput struct {
	UUID             string    `json:"token"`
	WriteOnly        bool      `json:"write_only"`
	NumberOfArchives uint64    `json:"keep_n_versions"`
	SizeUsed         uint64    `json:"size_used"`
	CreatedAt        time.Time `json:"created_at"`
}

type UserStatusOutput struct {
	EmailVerified         *time.Time     `json:"email_verified"`
	Paid                  bool           `json:"paid"`
	StartedSubscription   *time.Time     `json:"started_subscription"`
	CancelledSubscription *time.Time     `json:"cancelled_subscription"`
	PaymentID             string         `json:"payment_id"`
	LastVerificationID    string         `json:"-"`
	Tokens                []*TokenOutput `json:"tokens"`
	Quota                 uint64         `json:"quota"`
	QuotaUsed             uint64         `json:"used"`
	Email                 string         `json:"email"`
}

type QueryError struct {
	Error string `json:"error"`
}

var EMPTY_STATUS = &UserStatusOutput{
	PaymentID: "WILL-BE-IN-YOUR-EMAIL",
	Email:     "your.email@example.com",
	Tokens:    []*TokenOutput{&TokenOutput{UUID: "TOKEN-UUID-A", WriteOnly: true, NumberOfArchives: 3}, &TokenOutput{UUID: "TOKEN-UUID-B", WriteOnly: false, NumberOfArchives: 3}},
}

type LocalFile struct {
	File           *os.File
	SHA            string
	Size           uint64
	TempName       string
	OriginFullPath string
}
