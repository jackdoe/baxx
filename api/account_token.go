package main

import (
	"fmt"
	"strings"

	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/notification"
	"github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
)

func CreateToken(db *gorm.DB, writeOnly bool, u *user.User, numOfArchives uint64, name string) (*file.Token, error) {
	t := &file.Token{
		UUID:             common.GetUUID(),
		Salt:             strings.Replace(common.GetUUID(), "-", "", -1),
		Bucket:           strings.Replace(common.GetUUID(), "-", "", -1),
		UserID:           u.ID,
		WriteOnly:        writeOnly,
		NumberOfArchives: numOfArchives,
		Name:             name,
		Quota:            CONFIG.DefaultQuota,
		QuotaInode:       CONFIG.DefaultInodeQuota,
	}

	tokens, err := ListTokens(db, u)
	if err != nil {
		return nil, err
	}

	if len(tokens) >= int(CONFIG.MaxTokens) {
		return nil, fmt.Errorf("max tokens created (max=%d)", CONFIG.MaxTokens)
	}

	if err := db.Create(t).Error; err != nil {
		return nil, err
	}

	return t, nil
}

func FindTokenAndUser(db *gorm.DB, token string) (*file.Token, *user.User, error) {
	t := &file.Token{}

	query := db.Where("uuid = ?", token).Take(t)
	if query.RecordNotFound() {
		return nil, nil, query.Error
	}

	u := &user.User{}
	query = db.Where("id = ?", t.UserID).Take(u)
	if query.RecordNotFound() {
		return nil, nil, query.Error
	}

	return t, u, nil
}

func CreateTokenAndBucket(s *file.Store, db *gorm.DB, u *user.User, writeOnly bool, numOfArchives uint64, name string) (*file.Token, error) {
	t, err := CreateToken(db, writeOnly, u, numOfArchives, name)
	if err != nil {
		return nil, err
	}

	_, err = createNotificationRule(db, u, &common.CreateNotificationInput{
		Name:      "versions are too different",
		Regexp:    ".*",
		TokenUUID: t.UUID,
		AcceptableSizeDeltaPercentBetweenVersions: 90,
	})
	if err != nil {
		return nil, err
	}

	err = s.MakeBucket(t.Bucket)
	if err != nil {
		db.Delete(t)
		return nil, err
	}
	return t, nil
}

func transformTokenForSending(t *file.Token, usedSize, usedInodes int64, rules []*notification.NotificationRule) *common.TokenOutput {
	ru := []common.NotificationRuleOutput{}
	for _, r := range rules {
		ru = append(ru, notification.TransformRuleToOutput(r))
	}
	return &common.TokenOutput{
		ID:                t.ID,
		UUID:              t.UUID,
		Name:              t.Name,
		WriteOnly:         t.WriteOnly,
		NumberOfArchives:  t.NumberOfArchives,
		CreatedAt:         t.CreatedAt,
		QuotaUsed:         uint64(usedSize),
		QuotaInodeUsed:    uint64(usedInodes),
		Quota:             t.Quota,
		QuotaInode:        t.QuotaInode,
		NotificationRules: ru,
	}
}

func ListTokens(db *gorm.DB, u *user.User) ([]*file.Token, error) {
	tokens := []*file.Token{}
	if err := db.Where("user_id = ?", u.ID).Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

func FindTokenForUser(db *gorm.DB, token string, u *user.User) (*file.Token, error) {
	t := &file.Token{}

	query := db.Where("uuid = ? AND user_id = ?", token, u.ID).Take(t)
	if query.Error != nil {
		return nil, query.Error
	}

	return t, nil
}
