package main

import (
	"fmt"
	"regexp"

	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/notification"
	"github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
)

func createNotificationRule(db *gorm.DB, u *user.User, json *common.CreateNotificationInput) (*notification.NotificationRule, error) {
	token, err := FindTokenForUser(db, json.TokenUUID, u)
	if err != nil {
		return nil, err
	}

	_, err = regexp.Compile(json.Regexp)
	if err != nil {
		return nil, err
	}

	if json.AcceptableAgeDays < 0 {
		return nil, fmt.Errorf("age_days must be >= 0 (0 is disabled), got: %d", json.AcceptableAgeDays)
	}

	if json.AcceptableSizeDeltaPercentBetweenVersions < 0 {
		return nil, fmt.Errorf("delta_percent must be >= 0 (0 is disabled), got: %d", json.AcceptableSizeDeltaPercentBetweenVersions)
	}

	if json.AcceptableSizeDeltaPercentBetweenVersions == 0 && json.AcceptableAgeDays == 0 {
		return nil, fmt.Errorf("both ")
	}
	n := &notification.NotificationRule{
		TokenID:                                   token.ID,
		UserID:                                    u.ID,
		Regexp:                                    json.Regexp,
		Name:                                      json.Name,
		UUID:                                      common.GetUUID(),
		AcceptableAgeSeconds:                      uint64(json.AcceptableAgeDays) * 86400,
		AcceptableSizeDeltaPercentBetweenVersions: uint64(json.AcceptableSizeDeltaPercentBetweenVersions),
	}

	if err := db.Create(n).Error; err != nil {
		return nil, err
	}
	return n, nil
}

func changeNotificationRule(db *gorm.DB, u *user.User, json *common.ModifyNotificationInput) (*notification.NotificationRule, error) {
	n := &notification.NotificationRule{}
	if err := db.Where("uuid = ? AND user_id = ?", json.UUID, u.ID).First(&n).Error; err != nil {
		return nil, err
	}

	if json.Name != nil {
		n.Name = *json.Name
	}
	if json.AcceptableSizeDeltaPercentBetweenVersions != nil {
		if *json.AcceptableSizeDeltaPercentBetweenVersions < 0 {
			return nil, fmt.Errorf("delta_percent must be >= 0 (0 is disabled), got: %d", *json.AcceptableSizeDeltaPercentBetweenVersions)
		}

		n.AcceptableSizeDeltaPercentBetweenVersions = uint64(*json.AcceptableSizeDeltaPercentBetweenVersions)
	}
	if json.AcceptableAgeDays != nil {
		if *json.AcceptableAgeDays < 0 {
			return nil, fmt.Errorf("age_days must be >= 0 (0 is disabled), got: %d", *json.AcceptableAgeDays)
		}

		n.AcceptableAgeSeconds = uint64(*json.AcceptableAgeDays) * 86400
	}
	if json.Regexp != nil {
		n.Regexp = *json.Regexp
	}

	_, err := regexp.Compile(n.Regexp)
	if err != nil {
		return nil, err
	}

	if err := db.Save(n).Error; err != nil {
		return nil, err
	}
	return n, nil
}

func ListNotifications(db *gorm.DB, t *file.Token) ([]*notification.NotificationRule, error) {
	rules := []*notification.NotificationRule{}
	if err := db.Where("token_id = ?", t.ID).Find(&rules).Error; err != nil {
		return nil, err
	}

	return rules, nil
}
