package file

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"fmt"
	. "github.com/jackdoe/baxx/config"
	. "github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
	"github.com/pierrec/lz4"
	"github.com/satori/go.uuid"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FileMetadata struct {
	ID        uint64    `gorm:"primary_key" json:"-"`
	TokenID   uint64    `gorm:"not null" json:"-"`
	Path      string    `gorm:"not null" json:"path"`
	Filename  string    `gorm:"not null" json:"filename"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FileVersion struct {
	ID             uint64 `gorm:"primary_key" json:"id"`
	DuplicatedSave uint64 `gorm:"not null" json:"duplicate_save"`

	// denormalized for simplicity
	TokenID        uint64 `gorm:"not null" json:"-"`
	FileMetadataID uint64 `gorm:"not null" json:"-"`

	Size   uint64 `gorm:"not null" json:"size"`
	SHA256 string `gorm:"not null" json:"sha"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedAtNs uint64    `json:"-"`
}

func (fv *FileVersion) FSPath() string {
	return locate(fv.TokenID, fv.SHA256)
}

func split(s string) (string, string) {
	s = filepath.Clean(s)
	name := filepath.Base(s)
	dir := filepath.Dir(s)
	return dir, name
}

func locate(tokenid uint64, f string) string {
	dir := path.Join(CONFIG.FileRoot, "baxx", fmt.Sprintf("%d", tokenid), f[0:2], f[2:4])
	os.MkdirAll(dir, 0700)
	return path.Join(dir, f)
}

func saveUploadedFile(key string, temporary string, body io.Reader) (string, int64, error) {
	sha := sha256.New()
	tee := io.TeeReader(body, sha)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", 0, err
	}

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

func DeleteFile(db *gorm.DB, t *Token, p string) error {
	tx := db.Begin()

	_, fm, err := FindFile(tx, t, p)
	if err != nil {
		tx.Rollback()
		return err
	}

	versions, err := ListVersionsFile(tx, t, p)
	if err != nil {
		tx.Rollback()
		return err
	}
	removeFiles := []string{}
	// go and delete all versions

	for _, rm := range versions {
		if t.SizeUsed >= rm.Size {
			t.SizeUsed -= rm.Size
		}
		removeFiles = append(removeFiles, rm.FSPath())
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

	for _, f := range removeFiles {
		log.Printf("removing %s, because of file delete operation", f)
		os.Remove(f)
	}
	return nil

}

func FindFile(db *gorm.DB, t *Token, p string) (*FileVersion, *FileMetadata, error) {
	dir, name := split(p)
	fm := &FileMetadata{}
	if err := db.Where("token_id = ? AND filename = ? AND path = ?", t.ID, name, dir).Take(fm).Error; err != nil {
		return nil, nil, err

	}
	fv := &FileVersion{}
	if err := db.Where("file_metadata_id = ?", fm.ID).Order("updated_at_ns DESC").First(fv).Error; err != nil {
		return nil, nil, err
	}

	return fv, fm, nil
}

func FindAndOpenFile(db *gorm.DB, t *Token, p string) (*FileVersion, *os.File, io.Reader, error) {
	fv, _, err := FindFile(db, t, p)
	if err != nil {
		return nil, nil, nil, err
	}
	file, err := os.Open(fv.FSPath())
	if err != nil {
		return nil, nil, nil, err
	}

	reader, err := decompressAndDecrypt(t.Salt, file)
	if err != nil {
		file.Close()
		return nil, nil, nil, err
	}
	return fv, file, reader, nil

}

type FileMetadataAndVersion struct {
	FileMetadata *FileMetadata
	Versions     []*FileVersion
}

func LSAL(files []FileMetadataAndVersion) string {
	buf := bytes.NewBufferString("")
	grouped := map[string][]FileMetadataAndVersion{}

	for _, f := range files {
		grouped[f.FileMetadata.Path] = append(grouped[f.FileMetadata.Path], f)
	}

	keys := []string{}
	for p, _ := range grouped {
		keys = append(keys, p)
	}
	sort.Strings(keys)
	for _, k := range keys {
		files := grouped[k]
		fmt.Fprintf(buf, "%s:\n", k)
		size := uint64(0)
		for _, f := range files {
			for _, v := range f.Versions {
				size += v.Size
			}
		}
		fmt.Fprintf(buf, "total %d\n", size)
		for _, f := range files {
			for i, v := range f.Versions {
				fmt.Fprintf(buf, "%d\t%s\t%s@v%d\t%s\n", v.Size, v.CreatedAt.Format(time.ANSIC), f.FileMetadata.Filename, i, v.SHA256)
			}
		}
		fmt.Fprintf(buf, "\n")
	}
	return buf.String()
}

func ListFilesInPath(db *gorm.DB, t *Token, p string) ([]FileMetadataAndVersion, error) {
	metadata := []*FileMetadata{}
	p = strings.TrimSuffix(p, "/")

	if err := db.Where("token_id = ? AND path like ?", t.ID, p+"%").Order("id").Find(&metadata).Error; err != nil {
		return nil, err
	}

	out := []FileMetadataAndVersion{}

	for _, fm := range metadata {
		versions := []*FileVersion{}
		if err := db.Where("file_metadata_id = ?", fm.ID).Order("updated_at_ns ASC").Find(&versions).Error; err != nil {
			return nil, err
		}
		out = append(out, FileMetadataAndVersion{fm, versions})
	}

	return out, nil
}

func ListVersionsFile(db *gorm.DB, t *Token, p string) ([]*FileVersion, error) {
	dir, name := split(p)
	fm := &FileMetadata{}
	if err := db.Where(FileMetadata{TokenID: t.ID, Path: dir, Filename: name}).Take(&fm).Error; err != nil {
		return nil, err
	}

	versions := []*FileVersion{}
	if err := db.Where("file_metadata_id = ?", fm.ID).Order("updated_at_ns ASC").Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}

func DeleteToken(db *gorm.DB, token *Token) error {
	// delete metadata
	tx := db.Begin()
	if err := tx.Delete(FileMetadata{}, "token_id = ?", token.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// delete versions
	versions := []*FileVersion{}
	if err := tx.Where("token_id = ?", token.ID).Find(&versions).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, v := range versions {
		log.Printf("token delete removing %s", v.FSPath())
		os.Remove(v.FSPath())
	}

	// delete token
	if err := tx.Delete(Token{}, "id = ?", token.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func SaveFile(db *gorm.DB, t *Token, user *User, body io.Reader, p string) (*FileVersion, error) {
	dir, name := split(p)
	tempName := locate(t.ID, fmt.Sprintf("%d.%s", time.Now().UnixNano(), uuid.Must(uuid.NewV4())))
	defer os.Remove(tempName)
	sha, size, err := saveUploadedFile(t.Salt, tempName, body)
	if err != nil {
		return nil, err
	}
	tx := db.Begin()
	fm := &FileMetadata{}

	if err := tx.FirstOrCreate(&fm, FileMetadata{TokenID: t.ID, Path: dir, Filename: name}).Error; err != nil {

		tx.Rollback()
		return nil, err
	}

	// create file origin
	fv := &FileVersion{}

	res := tx.Where("token_id = ? AND sha256 = ?", t.ID, sha).Take(fv)
	if res.RecordNotFound() {
		// create new one
		fv.SHA256 = sha
		fv.Size = uint64(size)
		fv.FileMetadataID = fm.ID
		fv.TokenID = t.ID
		fv.DuplicatedSave = 0
		fv.UpdatedAtNs = uint64(time.Now().UnixNano())
		if err := tx.Save(fv).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// only count once if file is created
		t.SizeUsed += fv.Size

		if err := tx.Save(fv).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		fv.DuplicatedSave++
		fv.UpdatedAtNs = uint64(time.Now().UnixNano())
		if err := tx.Save(fv).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return fv, nil
	}

	// create file metadata if we did not create it

	// check how many versions we have of this file
	versions, err := ListVersionsFile(tx, t, p)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	limit := int(t.NumberOfArchives)
	removeFiles := []string{}
	if len(versions) > limit {
		toDelete := versions[0:(len(versions) - limit)]
		for _, rm := range toDelete {
			if t.SizeUsed >= rm.Size {
				t.SizeUsed -= rm.Size
			}

			removeFiles = append(removeFiles, rm.FSPath())
			if err := tx.Delete(rm).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	if err := tx.Save(t).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	left, err := user.GetQuotaLeft(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if left < 0 {
		tx.Rollback()
		return nil, errors.New("quota limit reached")
	}

	err = os.Rename(tempName, locate(t.ID, sha))
	if err != nil {
		os.Remove(tempName)
		tx.Rollback()
		return nil, err
	}

	// goooo
	if err := tx.Commit().Error; err != nil {
		// well at ths point we might have the file already saved..
		return nil, err
	}

	for _, f := range removeFiles {
		log.Printf("removing %s, limit: %d, versions: %d", f, limit, len(versions))
		os.Remove(f)
	}

	return fv, nil
}
