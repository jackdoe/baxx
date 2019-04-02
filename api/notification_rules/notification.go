package notification_rules

import (
	"math"
	"regexp"
	"time"

	"github.com/jackdoe/baxx/api/file"
	"github.com/jackdoe/baxx/common"
)

type NotificationRule struct {
	ID                                        uint64 `gorm:"primary_key"`
	UserID                                    uint64 `gorm:"type:bigint not null REFERENCES users(id) ON DELETE CASCADE"`
	TokenID                                   uint64 `gorm:"type:bigint not null REFERENCES tokens(id) ON DELETE CASCADE"`
	UUID                                      string `gorm:"type:varchar(255) not null unique"`
	Name                                      string `gorm:"type:varchar(255) not null"`
	Regexp                                    string `gorm:"type:varchar(255) not null"`
	AcceptableAgeSeconds                      uint64
	AcceptableSizeDeltaPercentBetweenVersions uint64
	CreatedAt                                 time.Time `json:"created_at"`
	UpdatedAt                                 time.Time `json:"updated_at"`
}

type NotificationForFileVersion struct {
	ID                 uint64    `gorm:"primary_key"`
	NotificationRuleID uint64    `gorm:"type:bigint not null REFERENCES notification_rules(id) ON DELETE CASCADE"`
	FileVersionID      uint64    `gorm:"type:bigint not null REFERENCES file_versions(id) ON DELETE CASCADE"`
	Count              uint64    `gorm:"type:bigint not null"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type NotificationForQuota struct {
	ID        uint64    `gorm:"primary_key"`
	TokenID   uint64    `gorm:"type:bigint not null REFERENCES tokens(id) ON DELETE CASCADE"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ExecuteRule(rule *NotificationRule, files []file.FileMetadataAndVersion) ([]common.FileNotification, error) {
	// super bad implementation
	// just checking the flow
	re, err := regexp.Compile(rule.Regexp)
	if err != nil {
		return nil, err
	}

	out := []common.FileNotification{}

	for _, f := range files {
		fullpath := f.FileMetadata.FullPath()
		now := time.Now()

		if re.MatchString(fullpath) {
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
			}

			if rule.AcceptableAgeSeconds > 0 {

				acceptableAge := version.CreatedAt.Add(time.Duration(rule.AcceptableAgeSeconds) * time.Second)
				if now.After(acceptableAge) {
					n := &common.AgeNotification{
						ActualAge: now.Sub(version.CreatedAt),
						Overdue:   now.Sub(acceptableAge),
					}
					current.Age = n
				}
			}

			if rule.AcceptableSizeDeltaPercentBetweenVersions > 0 && len(f.Versions) > 1 {
				lastVersion := version
				previousVersion := f.Versions[len(f.Versions)-2]
				delta := (1 + (float64(lastVersion.Size) - float64(previousVersion.Size))) / float64(1+lastVersion.Size)
				if (math.Abs(delta) * 100) > float64(rule.AcceptableSizeDeltaPercentBetweenVersions) {
					// delta trigger
					n := &common.SizeNotification{
						PreviousSize: previousVersion.Size,
						Delta:        delta * 100,
					}
					current.Size = n
				}
			}

			if current.Age != nil || current.Size != nil {
				out = append(out, current)
			}
		}
	}

	return out, nil
}

func TransformRuleToOutput(n *NotificationRule) common.NotificationRuleOutput {
	return common.NotificationRuleOutput{
		AcceptableAgeDays:                         n.AcceptableAgeSeconds / 86400,
		AcceptableSizeDeltaPercentBetweenVersions: n.AcceptableSizeDeltaPercentBetweenVersions,
		UUID:   n.UUID,
		Regexp: n.Regexp,
		Name:   n.Name,
	}
}
