package common

import (
	"errors"
	"fmt"
	"github.com/badoux/checkmail"
	"github.com/jackdoe/baxx/user"
	"time"
)

type CreateTokenInput struct {
	WriteOnly        bool   `json:"writeonly"`
	NumberOfArchives uint64 `json:"number_of_archives"`
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

type Success struct {
	Success bool `json:"success"`
}

func ValidateEmail(email string) error {
	err := checkmail.ValidateFormat(email)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid email address (%s)", err.Error()))
	}
	return nil
}
func ValidatePassword(p string) error {
	if len(p) < 8 {
		return errors.New("password is too short, refer to https://www.xkcd.com/936/")
	}
	return nil
}

type ChangeSecretOutput struct {
	Secret string `json:"secret"`
}

type UserStatusOutput struct {
	EmailVerified         *time.Time    `json:"email_verified"`
	Paid                  bool          `json:"paid"`
	StartedSubscription   *time.Time    `json:"started_subscription"`
	CancelledSubscription *time.Time    `json:"cancelled_subscription"`
	Secret                string        `json:"secret"`
	PaymentID             string        `json:"payment_id"`
	Tokens                []*user.Token `json:"tokens"`
	Quota                 uint64        `json:"quota"`
	QuotaUsed             uint64        `json:"used"`
}

type CreateUserOutput struct {
	Secret    string `json:"secret"`
	PaymentID string `json:"payment_id"`
	TokenWO   string `json:"token_wo"`
	TokenRW   string `json:"token_rw"`
	Help      string `json:"help"`
}

type DeleteToken struct {
	MoveFilesToToken string `json:"move_files_to_token"`
}

type QueryError struct {
	Error string `json:"error"`
}
