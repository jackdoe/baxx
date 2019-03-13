package main

import (
	"fmt"
	"strings"

	"github.com/jackdoe/baxx/common"
	"github.com/jackdoe/baxx/file"
	"github.com/jinzhu/gorm"
)

func (user *User) ListTokens(db *gorm.DB) ([]*file.Token, error) {
	tokens := []*file.Token{}
	if err := db.Where("user_id = ?", user.ID).Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

func (user *User) CreateToken(db *gorm.DB, writeOnly bool, numOfArchives uint64, name string) (*file.Token, error) {
	t := &file.Token{
		UUID:             getUUID(),
		Salt:             strings.Replace(getUUID(), "-", "", -1),
		Bucket:           strings.Replace(getUUID(), "-", "", -1),
		UserID:           user.ID,
		WriteOnly:        writeOnly,
		NumberOfArchives: numOfArchives,
		Name:             name,
		Quota:            CONFIG.DefaultQuota,
		QuotaInode:       CONFIG.DefaultInodeQuota,
	}

	tokens, err := user.ListTokens(db)
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

func FindToken(db *gorm.DB, token string) (*file.Token, *User, error) {
	t := &file.Token{}

	query := db.Where("uuid = ?", token).Take(t)
	if query.RecordNotFound() {
		return nil, nil, query.Error
	}

	u := &User{}
	query = db.Where("id = ?", t.UserID).Take(u)
	if query.RecordNotFound() {
		return nil, nil, query.Error
	}

	return t, u, nil
}

func FindTokenForUser(db *gorm.DB, token string, user *User) (*file.Token, error) {
	t := &file.Token{}

	query := db.Where("uuid = ? AND user_id = ?", token, user.ID).Take(t)
	if query.Error != nil {
		return nil, query.Error
	}

	return t, nil
}

func CreateTokenAndBucket(s *file.Store, db *gorm.DB, u *User, writeOnly bool, numOfArchives uint64, name string) (*file.Token, error) {
	t, err := u.CreateToken(db, writeOnly, numOfArchives, name)
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

func transformTokenForSending(t *file.Token, usedSize, usedInodes int64) *common.TokenOutput {
	return &common.TokenOutput{
		ID:               t.ID,
		UUID:             t.UUID,
		Name:             t.Name,
		WriteOnly:        t.WriteOnly,
		NumberOfArchives: t.NumberOfArchives,
		CreatedAt:        t.CreatedAt,
		QuotaUsed:        uint64(usedSize),
		QuotaInodeUsed:   uint64(usedInodes),
		Quota:            t.Quota,
		QuotaInode:       t.QuotaInode,
	}
}
