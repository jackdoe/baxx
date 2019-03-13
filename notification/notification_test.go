package notification

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/jackdoe/baxx/file"
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
	age, size, _ := ExecuteRule(rule, files)
	for _, a := range age {
		fmt.Printf("%s\n\n", a.String())
	}
	for _, a := range size {
		log.Printf("%s\n\n", a.String())
	}

	if len(age) != 2 {
		t.Fatalf("expected 2 notifications got %d", len(age))
	}
	if len(size) != 2 {
		t.Fatalf("expected 2 notifications got %d", len(size))
	}
}
