package common

import (
	"os"
	"time"
)

type CreateTokenInput struct {
	WriteOnly        bool   `json:"write_only"`
	Name             string `json:"name"`
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
	UUID string `binding:"required" json:"token"`
}

type Force struct {
	Force     *bool `json:"force"`
	Recursive *bool `json:"recursive"`
}

type ModifyTokenInput struct {
	WriteOnly        *bool  `json:"write_only"`
	Name             string `json:"name"`
	NumberOfArchives uint64 `json:"keep_n_versions"`
	UUID             string `binding:"required" json:"token"`
}

type CreateNotificationInput struct {
	AcceptableAgeDays                         int64  `json:"age_days"`
	AcceptableSizeDeltaPercentBetweenVersions int64  `json:"size_delta_percent"`
	TokenUUID                                 string `binding:"required" json:"token"`
	Regexp                                    string `binding:"required" json:"regexp"`
	Name                                      string `binding:"required" json:"name"`
}

type ModifyNotificationInput struct {
	AcceptableAgeDays                         *int64  `json:"age_days"`
	AcceptableSizeDeltaPercentBetweenVersions *int64  `json:"size_delta_percent"`
	UUID                                      string  `binding:"required" json:"notification_uuid"`
	Regexp                                    *string `json:"regexp"`
	Name                                      *string `json:"name"`
}

type DeleteNotificationInput struct {
	UUID string `binding:"required" json:"notification_uuid"`
}

type Success struct {
	Success bool `json:"success"`
}

type DeleteSuccess struct {
	Success bool `json:"success"`
	Count   int  `json:"deleted"`
}

type NotificationRuleOutput struct {
	AcceptableAgeDays                         uint64 `json:"age_days"`
	AcceptableSizeDeltaPercentBetweenVersions uint64 `json:"size_delta_percent"`
	UUID                                      string `json:"notification_uuid"`
	Regexp                                    string `json:"regexp"`
	Name                                      string `json:"name"`
}

type TokenOutput struct {
	UUID              string                   `json:"token"`
	ID                uint64                   `json:"id"`
	WriteOnly         bool                     `json:"write_only"`
	Name              string                   `json:"name"`
	NumberOfArchives  uint64                   `json:"keep_n_versions"`
	CreatedAt         time.Time                `json:"created_at"`
	Quota             uint64                   `json:"quota"`
	QuotaUsed         uint64                   `json:"quota_used"`
	QuotaInode        uint64                   `json:"quota_inodes"`
	QuotaInodeUsed    uint64                   `json:"quota_inodes_used"`
	NotificationRules []NotificationRuleOutput `json:"notification_rules"`
}

type UserStatusOutput struct {
	EmailVerified         *time.Time     `json:"email_verified"`
	Paid                  bool           `json:"paid"`
	StartedSubscription   *time.Time     `json:"started_subscription"`
	CancelledSubscription *time.Time     `json:"cancelled_subscription"`
	PaymentID             string         `json:"payment_id"`
	LastVerificationID    string         `json:"-"`
	Tokens                []*TokenOutput `json:"tokens"`
	Email                 string         `json:"email"`
}

type QueryError struct {
	Error string `json:"error"`
}

var EMPTY_STATUS = &UserStatusOutput{
	PaymentID: "WILL-BE-IN-YOUR-EMAIL",
	Email:     "your.email@example.com",
	Tokens: []*TokenOutput{
		&TokenOutput{
			ID:               0,
			UUID:             "TOKEN-UUID-A",
			WriteOnly:        false,
			NumberOfArchives: 3,
			Name:             "db-example-a",
			NotificationRules: []NotificationRuleOutput{
				NotificationRuleOutput{
					Name:              "more than 1 day old database backup",
					Regexp:            "\\.sql",
					AcceptableAgeDays: 1,
					UUID:              "NOTIFICATION-UUID",
				},
				NotificationRuleOutput{
					Name:   "database file is 50% different",
					Regexp: "\\.sql",
					UUID:   "NOTIFICATION-UUID",
					AcceptableSizeDeltaPercentBetweenVersions: 50,
				},
			},
		},
		&TokenOutput{
			ID:               0,
			UUID:             "TOKEN-UUID-B",
			WriteOnly:        false,
			NumberOfArchives: 3,
			Name:             "content-example-b",
			NotificationRules: []NotificationRuleOutput{
				NotificationRuleOutput{
					Name:              "more than 1 day old config backup",
					Regexp:            "etc\\.\\.tar\\.gz",
					UUID:              "NOTIFICATION-UUID",
					AcceptableAgeDays: 1,
				},
				NotificationRuleOutput{
					Name:   "file is 90% different",
					Regexp: ".*",
					UUID:   "NOTIFICATION-UUID",
					AcceptableSizeDeltaPercentBetweenVersions: 90,
				},
			},
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
	Overflow     uint64
}

type FileNotification struct {
	Age             *AgeNotification
	Size            *SizeNotification
	FullPath        string
	LastVersionSize uint64
	FileVersionID   uint64
	CreatedAt       time.Time
}

type PerRuleGroup struct {
	PerFile        []FileNotification
	Rule           NotificationRuleOutput
	AlertCreatedAt time.Time
}
