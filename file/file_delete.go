package file

import "github.com/jinzhu/gorm"

func DeleteToken(s *Store, db *gorm.DB, token *Token) error {
	// delete metadata
	tx := db.Begin()
	// delete versions
	removeFiles := []FileVersion{}
	versions := []*FileVersion{}
	if err := tx.Where("token_id = ?", token.ID).Find(&versions).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, v := range versions {
		removeFiles = append(removeFiles, *v)
	}

	if err := tx.Delete(FileVersion{}, "token_id = ?", token.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(FileMetadata{}, "token_id = ?", token.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// delete token
	if err := tx.Delete(Token{}, "id = ?", token.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return s.RemoveMany(token.Bucket, removeFiles)
}

func DeleteFileWithPath(s *Store, db *gorm.DB, t *Token, p string) error {
	_, fm, err := FindFile(db, t, p)
	if err != nil {
		return err
	}
	return DeleteFile(s, db, t, fm)
}

func DeleteFile(s *Store, db *gorm.DB, t *Token, fm *FileMetadata) error {
	tx := db.Begin()

	versions, err := ListVersionsFile(tx, t, fm)
	if err != nil {
		tx.Rollback()
		return err
	}
	// go and delete all versions
	remove := []FileVersion{}
	for _, rm := range versions {
		if t.SizeUsed >= rm.Size {
			t.SizeUsed -= rm.Size
		}
		remove = append(remove, *rm)
		if err := tx.Delete(rm).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// delete the metadata
	if err := tx.Delete(fm).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Save(t).Error; err != nil {
		tx.Rollback()
		return err
	}

	// goooo
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return s.RemoveMany(t.Bucket, remove)
}
