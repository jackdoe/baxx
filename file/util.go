package file

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jinzhu/gorm"
)

func split(s string) (string, string) {
	s = filepath.Clean(s)
	name := filepath.Base(s)
	dir := filepath.Dir(s)
	return dir, name
}

func prettySize(b uint64) string {
	gb := float64(b) / float64(1024*1024*1024)
	if gb < 0.00009 {
		return fmt.Sprintf("%.4fMB", gb*1024)
	}
	return fmt.Sprintf("%.4fGB", gb)
}

func FileLine(fm *FileMetadata, fv *FileVersion) string {
	isCurrent := ""
	if fm.LastVersionID == fv.ID {
		isCurrent = "*"
	}
	return fmt.Sprintf("%d\t%s\t%s@v%d%s\t%s\n", fv.Size, fv.CreatedAt.Format(time.ANSIC), fm.FullPath(), fv.ID, isCurrent, fv.SHA256)
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
	usedInodes := uint64(0)
	usedInodes, err := CountFilesPerToken(db, t)
	if err != nil {
		return 0, 0, err
	}
	return int64(usedSpace), int64(usedInodes), nil
}
