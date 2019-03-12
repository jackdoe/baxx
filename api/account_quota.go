package main

import (
	"github.com/jackdoe/baxx/file"
	"github.com/jinzhu/gorm"
)

func (user *User) GetQuotaLeft(db *gorm.DB) (int64, int64, error) {
	usedSize, usedInodes, err := user.GetQuotaUsed(db)
	if err != nil {
		return 0, 0, err
	}
	return int64(user.Quota) - int64(usedSize), int64(user.QuotaInode) - int64(usedInodes), nil
}

func (user *User) GetQuotaUsed(db *gorm.DB) (int64, int64, error) {
	tokens, err := user.ListTokens(db)
	if err != nil {
		return 0, 0, err
	}

	usedSpace := uint64(0)
	usedInodes := uint64(0)
	for _, t := range tokens {
		usedSpace += t.SizeUsed
		c, err := file.CountFilesPerToken(db, t)
		if err != nil {
			return 0, 0, err
		}
		usedInodes += c
	}
	return int64(usedSpace), int64(usedInodes), nil
}
