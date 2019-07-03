package file

import (
	"io"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type FileParams struct {
	KeepN           *uint64 `form:"keep_n" json:"keep_n"`
	AcceptableAge   *uint64 `form:"age" json:"age"`
	AcceptableDelta *uint64 `form:"delta" json:"delta"`
	ContentType     string
	FullPath        string
}

func SaveFile(s *Store, db *gorm.DB, t *Token, body io.Reader, fp FileParams) (*FileVersion, *FileMetadata, error) {
	fullpath := fp.FullPath
	// upload the file to s3
	storeID := GetStoreId(t.ID)
	sha, size, err := s.UploadFile(t.Salt, t.UUID, storeID, body)
	if err != nil {
		return nil, nil, err
	}

	removeBeforeExit := map[string]bool{storeID: true}
	defer func() {
		for id := range removeBeforeExit {
			log.Infof("on save removing %d %s %s", t.ID, fullpath, id)
			err := s.DeleteFile(t.UUID, id)
			if err != nil {
				log.Warnf("error removing %s: %s", id, err.Error())
			}
		}
	}()

	// get the metadata
	tx := db.Begin()
	dir, name := split(fullpath)
	fm := &FileMetadata{}
	if err := tx.FirstOrCreate(&fm, FileMetadata{TokenID: t.ID, Path: dir, Filename: name}).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	fv := &FileVersion{
		Size:           uint64(size),
		FileMetadataID: fm.ID,
		SHA256:         sha,
		TokenID:        t.ID,
		StoreID:        storeID,
		ContentType:    fp.ContentType,
	}
	if err := tx.Create(fv).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	fm.LastVersionID = fv.ID
	fm.CountWrite++

	if fp.KeepN != nil && *fp.KeepN >= 1 {
		fm.KeepN = *fp.KeepN
	}

	if fp.AcceptableAge != nil && *fp.AcceptableAge >= 0 {
		fm.AcceptableAge = *fp.AcceptableAge
	}

	if fp.AcceptableDelta != nil && *fp.AcceptableDelta >= 0 {
		fm.AcceptableDelta = *fp.AcceptableDelta
	}

	t.SizeUsed += fv.Size
	t.CountFiles++

	if err := tx.Save(fm).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	versions, err := ListVersionsFile(tx, t, fm)
	if err != nil {
		return nil, nil, err
	}

	limit := int(fm.KeepN)
	if len(versions) > limit {
		needToDelete := len(versions) - limit
	DELETE:
		for _, rm := range versions {
			if rm.ID == fv.ID {
				continue
			}

			t.SizeUsed -= rm.Size
			t.CountFiles--
			removeBeforeExit[rm.StoreID] = true
			if err := tx.Delete(rm).Error; err != nil {
				tx.Rollback()
				return nil, nil, err
			}
			needToDelete--
			if needToDelete == 0 {
				break DELETE
			}
		}
	}

	if err := tx.Save(t).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	delete(removeBeforeExit, storeID)

	return fv, fm, nil
}
