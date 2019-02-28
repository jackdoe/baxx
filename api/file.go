package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"github.com/pierrec/lz4"
	"github.com/satori/go.uuid"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

type FileOrigin struct {
	ID     uint64 `gorm:"primary_key"`
	Size   uint64 `gorm:"not null"`
	SHA256 string `gorm:"not null"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type FileMetadata struct {
	ID uint64 `gorm:"primary_key"`

	ClientID string `gorm:"not null"`
	TokenID  string `gorm:"not null"`
	Path     string `gorm:"not null"`
	Filename string `gorm:"not null"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type FileVersion struct {
	ID             uint64 `gorm:"primary_key"`
	FileMetadataID uint64 `gorm:"not null" json:"-"`
	FileOriginID   uint64 `gorm:"not null" json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func split(s string) (string, string) {
	s = filepath.Clean(s)
	name := filepath.Base(s)
	dir := filepath.Dir(s)
	return dir, name
}

func locate(f string) string {
	dir := path.Join("/", "tmp", "baxx")
	return path.Join(dir, f)
}

func saveUploadedFile(key string, body io.Reader) (string, int64, error) {
	sha := sha256.New()
	tee := io.TeeReader(body, sha)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", 0, err
	}

	temporary := locate(fmt.Sprintf("%d.%s", time.Now().UnixNano(), uuid.Must(uuid.NewV4())))
	dest, err := os.Create(temporary)
	if err != nil {
		return "", 0, err
	}
	var iv [aes.BlockSize]byte

	stream := cipher.NewOFB(block, iv[:])
	encryptedWriter := &cipher.StreamWriter{S: stream, W: dest}
	// compress -> encrypt
	lz4Writer := lz4.NewWriter(encryptedWriter)
	size, err := io.Copy(lz4Writer, tee)
	if err != nil {
		dest.Close()
		os.Remove(temporary)
		return "", 0, err
	}
	// XXX: not to be trusted, attacker can flip bits
	// the only reason we encrypt is so we dont accidentally receive unencrypted data
	// or if someone steals the data

	lz4Writer.Close()
	encryptedWriter.Close()
	dest.Close()

	shasum := fmt.Sprintf("%x", sha.Sum(nil))
	err = os.Rename(temporary, locate(shasum))
	if err != nil {
		os.Remove(temporary)
		return "", 0, err
	}
	return shasum, size, nil
}

func decompressAndDecrypt(key string, r io.Reader) (io.Reader, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])
	// compress -> encrypt -> decrypt -> decompress
	decryptReader := &cipher.StreamReader{S: stream, R: r}
	lz4Reader := lz4.NewReader(decryptReader)
	return lz4Reader, nil
	// XXX: not to be trusted, attacker can flip bits
	// the only reason we encrypt is so we dont accidentally receive unencrypted data
	// or if someone steals the data

}
