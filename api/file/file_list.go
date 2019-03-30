package file

import (
	"strings"

	"github.com/jinzhu/gorm"
)

func ListFilesInPath(db *gorm.DB, t *Token, p string, strict bool) ([]FileMetadataAndVersion, error) {
	metadata := []*FileMetadata{}
	p = strings.TrimSuffix(p, "/")
	if p == "" {
		p = "/"
	}
	if strict {
		if err := db.Where("token_id = ? AND path = ?", t.ID, p).Order("id").Find(&metadata).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.Where("token_id = ? AND path like ?", t.ID, p+"%").Order("id").Find(&metadata).Error; err != nil {
			return nil, err
		}
	}

	out := []FileMetadataAndVersion{}

	for _, fm := range metadata {
		versions := []*FileVersion{}
		if err := db.Where("file_metadata_id = ?", fm.ID).Find(&versions).Error; err != nil {
			return nil, err
		}
		out = append(out, FileMetadataAndVersion{fm, versions})
	}

	return out, nil
}
