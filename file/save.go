package file

import (
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func SaveFile(s *Store, db *gorm.DB, t *Token, fullpath string, body io.Reader) (*FileVersion, *FileMetadata, error) {
	// get the metadata

	// upload the file to s3
	storeID := fmt.Sprintf("%d.%s", t.ID, uuid.Must(uuid.NewV4()).String())
	shah := sha256.New()
	tee := io.TeeReader(body, shah)
	sha := fmt.Sprintf("%x", shah.Sum(nil))
	size, err := s.UploadFile(storeID, tee)
	if err != nil {
		return nil, nil, err
	}

	removeBeforeExit := map[string]bool{storeID: true}
	defer func() {
		log.Printf("before exit: %+v", removeBeforeExit)
		for id, _ := range removeBeforeExit {
			log.Infof("on save removing %d %s %s", t.ID, fullpath, id)
			err := s.DeleteFile(id)
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
	}
	if err := tx.Create(fv).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	fm.LastVersionID = fv.ID
	t.SizeUsed += fv.Size

	if err := tx.Save(fm).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	versions, err := ListVersionsFile(tx, t, fm)
	if err != nil {
		return nil, nil, err
	}

	limit := int(t.NumberOfArchives)
	if len(versions) > limit {
		needToDelete := len(versions) - limit
	DELETE:
		for _, rm := range versions {
			if rm.ID == fv.ID {
				continue
			}

			t.SizeUsed -= rm.Size
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
