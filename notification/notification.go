package notification

import (
	"math"
	"regexp"
	"time"

	"github.com/jackdoe/baxx/file"
	"github.com/jackdoe/baxx/help"
)

type AgeNotification struct {
	CreatedAt time.Time
	ActualAge time.Duration
	Overdue   time.Duration
}

func (n *AgeNotification) String() string {
	return help.Render(help.EMAIL_AGE_RULE, n)
}

type SizeNotification struct {
	CurrentSize  uint64
	PreviousSize uint64
	Delta        float64
	Overflow     float64
}

func (n *SizeNotification) String() string {
	return help.Render(help.EMAIL_SIZE_RULE, n)
}

type FileNotification struct {
	Age         *AgeNotification
	Size        *SizeNotification
	FileVersion *file.FileVersion
	FullPath    string

	CreatedAt time.Time
}

func ExecuteRule(rule *NotificationRule, files []file.FileMetadataAndVersion) ([]FileNotification, error) {
	// super bad implementation
	// just checking the flow
	re, err := regexp.Compile(rule.Regexp)
	if err != nil {
		return nil, err
	}

	out := []FileNotification{}

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
			current := FileNotification{
				FileVersion: f.Versions[0],
				CreatedAt:   f.Versions[0].CreatedAt,
				FullPath:    fullpath,
			}

			if rule.AcceptableAgeSeconds > 0 {
				version := f.Versions[0]
				acceptableAge := version.CreatedAt.Add(time.Duration(rule.AcceptableAgeSeconds) * time.Second)
				if now.After(acceptableAge) {
					n := &AgeNotification{
						ActualAge: now.Sub(version.CreatedAt),
						Overdue:   now.Sub(acceptableAge),
					}
					current.Age = n
				}
			}

			if rule.AcceptableSizeDeltaPercentBetweenVersions > 0 && len(f.Versions) > 1 {
				lastVersion := f.Versions[0]
				previousVersion := f.Versions[1]
				delta := (float64(lastVersion.Size) - float64(previousVersion.Size)) / float64(lastVersion.Size)
				if (math.Abs(delta) * 100) > float64(rule.AcceptableSizeDeltaPercentBetweenVersions) {
					// delta trigger
					n := &SizeNotification{
						CurrentSize:  lastVersion.Size,
						PreviousSize: previousVersion.Size,
						Delta:        delta * 100,
						Overflow:     float64(lastVersion.Size) * delta,
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
	Count              uint64    `gorm:"type:bigint not null`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

/*
#  same file with changes
example database.sql.gz
fnew version of the file:
+ is too small [table was deleted or truncated?]
+ too big [ wrong table is backed up?]
+ too old [ not backed up in a while ]
+ deleted [ bad script deletes everything ]
+ file is corrupt
+ table is missing

# content
such as photos
+ removed too many files [in case of wrong delete]
+ no backup in a while
+ directory was not updated in a while
+ daily/weekly update rate is weird


at glance there are 2 main scenarios, anomalies per file and per directory
so basic rule might look like
{
   "path": 'regex',     // can be empty, by default is "all"
   "watch": [
      {
        "what": "age",
        "max": 10 * 24 * 3600 // 10 days
      },
      {
        "what": "count",
        "interval": 3600, //
        "delta": 10 * 24 * 3600 // 10 days
      },
   ]
}

this approach seems very limited, but i guess its ok for now
*/
