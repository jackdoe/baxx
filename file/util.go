package file

import (
	"path/filepath"

	"github.com/jinzhu/gorm"
)

func split(s string) (string, string) {
	s = filepath.Clean(s)
	name := filepath.Base(s)
	dir := filepath.Dir(s)
	return dir, name
}

func CountFilesPerToken(db *gorm.DB, t *Token) (uint64, error) {
	c := uint64(0)
	if err := db.Model(&FileVersion{}).Where("token_id = ?", t.ID).Count(&c).Error; err != nil {
		return 0, err
	}
	return c, nil
}

func GetQuotaLeft(db *gorm.DB, t *Token) (int64, int64, error) {
	usedSize, usedInodes, err := GetQuotaUsed(db, t)
	if err != nil {
		return 0, 0, err
	}
	return int64(t.Quota) - int64(usedSize), int64(t.QuotaInode) - int64(usedInodes), nil
}

func GetQuotaUsed(db *gorm.DB, t *Token) (int64, int64, error) {
	usedSpace := t.SizeUsed
	usedInodes, err := CountFilesPerToken(db, t)
	if err != nil {
		return 0, 0, err
	}
	return int64(usedSpace), int64(usedInodes), nil
}
