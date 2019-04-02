package file

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func FindFile(db *gorm.DB, t *Token, p string) (*FileVersion, *FileMetadata, error) {
	dir, name := split(p)
	fm := &FileMetadata{}
	if err := db.Where("token_id = ? AND filename = ? AND path = ?", t.ID, name, dir).Take(fm).Error; err != nil {
		return nil, nil, err

	}
	if fm.LastVersionID == 0 {
		return nil, nil, fmt.Errorf("file without version, probably interrupted, please reupload")
	}
	fv := &FileVersion{}
	if err := db.Where("id = ?", fm.LastVersionID).Last(fv).Error; err != nil {
		return nil, nil, err
	}

	return fv, fm, nil
}

func FindFileBySHA(db *gorm.DB, t *Token, sha string) (*FileVersion, *FileMetadata, error) {
	// FIXME(jackdoe): make sure the found file is actually the latest version
	fv := &FileVersion{}
	if err := db.Where("token_id = ? AND sha256 = ?", t.ID, sha).Last(fv).Error; err != nil {
		return nil, nil, err
	}

	fm := &FileMetadata{}
	if err := db.Where("id = ?", fv.FileMetadataID).Take(fm).Error; err != nil {
		return nil, nil, err
	}

	if fv.ID != fm.LastVersionID {
		return nil, nil, fmt.Errorf("sha exists, but not as latest version")
	}
	return fv, fm, nil
}

func ListVersionsFile(db *gorm.DB, t *Token, fm *FileMetadata) ([]*FileVersion, error) {
	versions := []*FileVersion{}
	if err := db.Where("file_metadata_id = ?", fm.ID).Order("id").Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}
