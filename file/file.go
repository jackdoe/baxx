package file

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/jackdoe/baxx/config"
	. "github.com/jackdoe/baxx/user"
	"github.com/jinzhu/gorm"
	"github.com/pierrec/lz4"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Store struct {
	temporaryRoot string
	bucket        string
	sess          *session.Session
	uploader      *s3manager.Uploader
	downloader    *s3manager.Downloader
	batcher       *s3manager.BatchDelete
}

func (s *Store) getTemporaryName(tokenid uint64) string {
	dir := path.Join(s.temporaryRoot, "baxx")
	os.MkdirAll(dir, 0700)
	return path.Join(dir, fmt.Sprintf("%d.%d.%s", time.Now().UnixNano(), tokenid, uuid.Must(uuid.NewV4())))
}

func NewStore(conf *StoreConfig) *Store {
	creds := credentials.NewStaticCredentials(conf.AccessKeyID, conf.SecretAccessKey, conf.SessionToken)
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		DisableSSL:  &conf.DisableSSL,
		Endpoint:    &conf.Endpoint,
		Region:      &conf.Region,
	}))

	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)
	batcher := s3manager.NewBatchDelete(sess)
	return &Store{
		bucket:        conf.Bucket,
		temporaryRoot: conf.TemporaryRoot,
		sess:          sess,
		batcher:       batcher,
		uploader:      uploader,
		downloader:    downloader,
	}
}

type FileMetadata struct {
	ID        uint64    `gorm:"primary_key" json:"-"`
	TokenID   uint64    `gorm:"not null" json:"-"`
	Path      string    `gorm:"not null" json:"path"`
	Filename  string    `gorm:"not null" json:"filename"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (fm *FileMetadata) FullPath() string {
	if fm.Path == "/" {
		return fmt.Sprintf("/%s", fm.Filename)
	}
	return fmt.Sprintf("%s/%s", fm.Path, fm.Filename)
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

func (v *FileVersion) key() string {
	return fmt.Sprintf("%d.%d.%s", v.ID, v.TokenID, v.SHA256)
}

func split(s string) (string, string) {
	s = filepath.Clean(s)
	name := filepath.Base(s)
	dir := filepath.Dir(s)
	return dir, name
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

func DeleteFile(s *Store, db *gorm.DB, t *Token, p string) error {
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
	removeFiles := []s3manager.BatchDeleteObject{}
	// go and delete all versions

	for _, rm := range versions {
		if t.SizeUsed >= rm.Size {
			t.SizeUsed -= rm.Size
		}
		removeFiles = append(removeFiles, s3manager.BatchDeleteObject{
			Object: &s3.DeleteObjectInput{
				Key:    aws.String(rm.key()),
				Bucket: aws.String(s.bucket),
			},
		})
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

	if err := s.batcher.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{
		Objects: removeFiles,
	}); err != nil {
		tx.Rollback()
		return err
	}

	// goooo
	if err := tx.Commit().Error; err != nil {
		return err
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

func FindFileBySHA(db *gorm.DB, t *Token, sha string) (*FileVersion, *FileMetadata, error) {
	// FIXME(jackdoe): make sure the found file is actually the latest version
	fv := &FileVersion{}
	if err := db.Where("token_id = ? AND sha256 = ?", t.ID, sha).Take(fv).Error; err != nil {
		return nil, nil, err
	}

	fm := &FileMetadata{}
	if err := db.Where("id = ?", fv.FileMetadataID).Take(fm).Error; err != nil {
		return nil, nil, err
	}

	return fv, fm, nil
}

func FindAndOpenFile(s *Store, db *gorm.DB, t *Token, p string) (*FileVersion, func(), io.Reader, error) {
	fv, _, err := FindFile(db, t, p)
	if err != nil {
		return nil, nil, nil, err
	}
	tempName := s.getTemporaryName(t.ID)
	file, err := os.OpenFile(tempName, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, nil, nil, err
	}

	done := func() {
		file.Close()
		os.Remove(tempName)
	}

	// XXX: saving the file locally so we can download it concurrently
	// Download() takes WriterAt because of the chunks and even
	// though we can use inmemory buffer, it might get too big
	// so just save the file on disk and then delete it
	_, err = s.downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(fv.key()),
		})
	if err != nil {
		return nil, nil, nil, err
	}

	reader, err := decompressAndDecrypt(t.Salt, file)
	if err != nil {
		return nil, nil, nil, err
	}
	return fv, done, reader, nil

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
	for p := range grouped {
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

func DeleteToken(s *Store, db *gorm.DB, token *Token) error {
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

	removeFiles := []s3manager.BatchDeleteObject{}

	for _, v := range versions {
		removeFiles = append(removeFiles, s3manager.BatchDeleteObject{
			Object: &s3.DeleteObjectInput{
				Key:    aws.String(v.key()),
				Bucket: aws.String(s.bucket),
			},
		})
	}

	// delete token
	if err := tx.Delete(Token{}, "id = ?", token.ID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := s.batcher.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{
		Objects: removeFiles,
	}); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func SaveFile(s *Store, db *gorm.DB, t *Token, user *User, body io.Reader, p string) (*FileVersion, *FileMetadata, error) {
	dir, name := split(p)
	tempName := s.getTemporaryName(t.ID)
	defer os.Remove(tempName)
	sha, size, err := saveUploadedFile(t.Salt, tempName, body)
	if err != nil {
		return nil, nil, err
	}
	tx := db.Begin()
	fm := &FileMetadata{}

	if err := tx.FirstOrCreate(&fm, FileMetadata{TokenID: t.ID, Path: dir, Filename: name}).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
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
			return nil, nil, err
		}

		// only count once if file is created
		t.SizeUsed += fv.Size

		if err := tx.Save(fv).Error; err != nil {
			tx.Rollback()
			return nil, nil, err
		}
	} else {
		fv.DuplicatedSave++
		fv.UpdatedAtNs = uint64(time.Now().UnixNano())
		if err := tx.Save(fv).Error; err != nil {
			tx.Rollback()
			return nil, nil, err
		}

		if err := tx.Commit().Error; err != nil {
			return nil, nil, err
		}
		return fv, nil, nil
	}

	// create file metadata if we did not create it

	// check how many versions we have of this file
	versions, err := ListVersionsFile(tx, t, p)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	limit := int(t.NumberOfArchives)
	removeFiles := []s3manager.BatchDeleteObject{}
	if len(versions) > limit {
		toDelete := versions[0:(len(versions) - limit)]
		for _, rm := range toDelete {
			if t.SizeUsed >= rm.Size {
				t.SizeUsed -= rm.Size
			}

			removeFiles = append(removeFiles, s3manager.BatchDeleteObject{
				Object: &s3.DeleteObjectInput{
					Key:    aws.String(rm.key()),
					Bucket: aws.String(s.bucket),
				},
			})

			if err := tx.Delete(rm).Error; err != nil {
				tx.Rollback()
				return nil, nil, err
			}
		}
	}

	if err := tx.Save(t).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	left, err := user.GetQuotaLeft(tx)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if left < 0 {
		tx.Rollback()
		return nil, nil, errors.New("quota limit reached")
	}
	f, err := os.Open(tempName)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	defer f.Close()

	_, err = s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fv.key()),
		Body:   f,
	})
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	log.Infof("removing limit: %d, versions: %d %+v", limit, len(versions), removeFiles)
	if err := s.batcher.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{
		Objects: removeFiles,
	}); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// goooo
	if err := tx.Commit().Error; err != nil {
		// well at ths point we might have the file already saved..
		return nil, nil, err
	}

	return fv, fm, nil
}
