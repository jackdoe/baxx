package main

import (
	"testing"
	"time"

	"github.com/jackdoe/baxx/api/file"
	"github.com/jackdoe/baxx/api/notification_rules"
)

func secondsAgo(s int) time.Time {
	return time.Now().Add(time.Duration(-1*s) * time.Second)
}

func TestAge(t *testing.T) {
	files := []file.FileMetadataAndVersion{
		file.FileMetadataAndVersion{
			FileMetadata: &file.FileMetadata{
				Path:            "/backup",
				Filename:        "backup.sql",
				AcceptableAge:   10,
				AcceptableDelta: 10,
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
				Path:            "/backup",
				Filename:        "backup2.sql",
				AcceptableAge:   10,
				AcceptableDelta: 10,
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
				Path:            "/backup",
				Filename:        "backup3.sql",
				AcceptableAge:   10,
				AcceptableDelta: 10,
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
	n := notification_rules.ExecuteRule(files)
	if len(n) != 3 {
		t.Fatalf("expected 3 notifications got %d", len(n))
	}
}
