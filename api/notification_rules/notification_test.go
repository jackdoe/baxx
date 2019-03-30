package notification_rules

import (
	"testing"
	"time"

	"github.com/jackdoe/baxx/api/file"
)

func secondsAgo(s int) time.Time {
	return time.Now().Add(time.Duration(-1*s) * time.Second)
}

func TestAge(t *testing.T) {
	rule := &NotificationRule{
		Name:                                      "age",
		Regexp:                                    "\\.sql",
		AcceptableAgeSeconds:                      10,
		AcceptableSizeDeltaPercentBetweenVersions: 10,
	}

	files := []file.FileMetadataAndVersion{
		file.FileMetadataAndVersion{
			FileMetadata: &file.FileMetadata{
				Path:     "/backup",
				Filename: "backup.sql",
			},
			Versions: []*file.FileVersion{
				&file.FileVersion{
					Size:      100,
					CreatedAt: secondsAgo(11),
				},
				&file.FileVersion{
					Size:      50,
					CreatedAt: secondsAgo(200),
				},
			},
		},
		file.FileMetadataAndVersion{
			FileMetadata: &file.FileMetadata{
				Path:     "/backup",
				Filename: "backup2.sql",
			},
			Versions: []*file.FileVersion{
				&file.FileVersion{
					Size:      100,
					CreatedAt: secondsAgo(9),
				},
				&file.FileVersion{
					Size:      120,
					CreatedAt: secondsAgo(11),
				},
			},
		},
		file.FileMetadataAndVersion{
			FileMetadata: &file.FileMetadata{
				Path:     "/backup",
				Filename: "backup3.sql",
			},
			Versions: []*file.FileVersion{
				&file.FileVersion{
					Size:      100,
					CreatedAt: secondsAgo(12),
				},
				&file.FileVersion{
					Size:      91,
					CreatedAt: secondsAgo(11),
				},
			},
		},
	}
	n, _ := ExecuteRule(rule, files)
	if len(n) != 3 {
		t.Fatalf("expected 3 notifications got %d", len(n))
	}
}
