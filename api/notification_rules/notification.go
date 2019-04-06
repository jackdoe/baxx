package notification_rules

import (
	"math"
	"time"

	"github.com/jackdoe/baxx/api/file"
	"github.com/jackdoe/baxx/common"
	"github.com/jinzhu/gorm"
)

type NotificationForFileVersion struct {
	ID             uint64    `gorm:"primary_key"`
	FileMetadataID uint64    `gorm:"type:bigint not null REFERENCES file_metadata(id) ON DELETE CASCADE"`
	FileVersionID  uint64    `gorm:"type:bigint not null REFERENCES file_versions(id) ON DELETE CASCADE"`
	Count          uint64    `gorm:"type:bigint not null"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type NotificationForUserQuota struct {
	ID        uint64    `gorm:"primary_key"`
	UserID    uint64    `gorm:"type:bigint not null REFERENCES users(id) ON DELETE CASCADE"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func IgnoreAndMarkAlreadyNotified(db *gorm.DB, n []common.FileNotification) ([]common.FileNotification, error) {
	out := []common.FileNotification{}
	for _, current := range n {
		nfv := &NotificationForFileVersion{
			FileMetadataID: current.FileMetadataID,
			FileVersionID:  current.FileVersionID,
		}
		res := db.Where(nfv).First(&nfv)
		if err := res.Error; err != nil {
			if res.RecordNotFound() {
				nfv.Count++
				if err := db.Save(nfv).Error; err != nil {
					return nil, err
				}
				out = append(out, current)
			} else {
				return nil, err
			}

		}
	}
	return out, nil
}

func ExecuteRule(files []file.FileMetadataAndVersion) []common.FileNotification {
	// super bad implementation
	// just checking the flow

	out := []common.FileNotification{}

	for _, f := range files {
		fullpath := f.FileMetadata.FullPath()
		now := time.Now()
		if f.FileMetadata.AcceptableAge == 0 && f.FileMetadata.AcceptableDelta == 0 {
			continue
		}

		if len(f.Versions) == 0 {
			// this is possible if the file is being uploaded now
			continue
		}

		// both of those are super simplified
		// FIXME(jackdoe): more work is needed!
		version := f.Versions[len(f.Versions)-1]
		current := common.FileNotification{
			CreatedAt:       version.CreatedAt,
			FullPath:        fullpath,
			LastVersionSize: version.Size,
			FileVersionID:   version.ID,
			FileMetadataID:  f.FileMetadata.ID,
		}

		if f.FileMetadata.AcceptableAge > 0 {
			acceptableAge := version.CreatedAt.Add(time.Duration(f.FileMetadata.AcceptableAge) * time.Second)
			if now.After(acceptableAge) {
				n := &common.AgeNotification{
					ActualAge: now.Sub(version.CreatedAt),
					Overdue:   now.Sub(acceptableAge),
				}
				current.Age = n
			}
		}

		if f.FileMetadata.AcceptableDelta > 0 && len(f.Versions) > 1 {
			lastVersion := version
			previousVersion := f.Versions[len(f.Versions)-2]
			deltaPercent := float64(0)
			if lastVersion.Size == 0 {
				deltaPercent = float64(previousVersion.Size) / float64(100)
			} else {
				deltaSize := float64(lastVersion.Size) - float64(previousVersion.Size)
				deltaPercent = float64(100) * (deltaSize / float64(lastVersion.Size))
			}
			if (math.Abs(deltaPercent)) > float64(f.FileMetadata.AcceptableDelta) {
				// delta trigger
				n := &common.SizeNotification{
					PreviousSize: previousVersion.Size,
					Delta:        deltaPercent,
				}
				current.Size = n
			}
		}
		if current.Age != nil || current.Size != nil {
			out = append(out, current)
		}
	}

	return out
}
