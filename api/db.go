package main

import (
	"github.com/jinzhu/gorm"
	"io"
)

type Store struct {
	db *gorm.DB
}

func (s *Store) FindToken(client string, token string) (*Token, error) {
	t := &Token{}
	query := s.db.Where("client_id = ? AND id = ?", client, token).Take(t)
	if query.RecordNotFound() {
		return nil, query.Error
	}
	return t, nil
}

func (s *Store) FindFile(t *Token, dir string, name string) (*FileOrigin, error) {
	fm := &FileMetadata{}
	if err := s.db.Where("client_id = ? AND token_id = ? AND filename = ? AND path = ?", t.ClientID, t.ID, name, dir).Take(fm).Error; err != nil {
		return nil, err

	}
	fv := &FileVersion{}
	if err := s.db.Where("file_metadata_id = ?", fm.ID).Last(fv).Error; err != nil {
		return nil, err
	}

	fo := &FileOrigin{}
	if err := s.db.Where("id = ?", fv.FileOriginID).Take(fo).Error; err != nil {
		return nil, err
	}

	return fo, nil
}

func (s *Store) SaveFile(t *Token, body io.Reader, dir string, name string) (*FileVersion, error) {
	sha, size, err := saveUploadedFile(t.Salt, body)
	if err != nil {
		return nil, err
	}

	// create file origin
	fo := &FileOrigin{}
	tx := s.db.Begin()
	res := tx.Where("sha256 = ?", sha).Take(fo)
	if res.RecordNotFound() {
		// create new one
		fo.SHA256 = sha
		fo.Size = uint64(size)
		if err := tx.Save(fo).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// create file metadata
	fm := &FileMetadata{}
	if err := tx.FirstOrCreate(&fm, FileMetadata{ClientID: t.ClientID, TokenID: t.ID, Path: dir, Filename: name}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// create the version
	fv := &FileVersion{}
	if err := tx.Where(FileVersion{FileMetadataID: fm.ID, FileOriginID: fo.ID}).FirstOrCreate(&fv).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// goooo
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return fv, nil

}
