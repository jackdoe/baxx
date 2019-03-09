package file

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/jackdoe/baxx/baxxio"
	. "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/config"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Token struct {
	ID     uint64 `gorm:"primary_key"`
	UUID   string `gorm:"not null"`
	Salt   string `gorm:"not null;type:varchar(32)"`
	Name   string `gorm:"null;type:varchar(255)"`
	UserID uint64 `gorm:"not null"`

	WriteOnly        bool   `gorm:"not null"`
	NumberOfArchives uint64 `gorm:"not null"`
	SizeUsed         uint64 `gorm:"not null;default:0"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Store struct {
	bucket     string
	sess       *session.Session
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	batcher    *s3manager.BatchDelete
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
		bucket:     conf.Bucket,
		sess:       sess,
		batcher:    batcher,
		uploader:   uploader,
		downloader: downloader,
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

func s3key(tokenid, fileMetadataId uint64, sha string) string {
	return fmt.Sprintf("%d.%d.%s", tokenid, fileMetadataId, sha)
}

func split(s string) (string, string) {
	s = filepath.Clean(s)
	name := filepath.Base(s)
	dir := filepath.Dir(s)
	return dir, name
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
	removeFiles := []s3manager.BatchDeleteObject{}
	// go and delete all versions

	for _, rm := range versions {
		if t.SizeUsed >= rm.Size {
			t.SizeUsed -= rm.Size
		}
		removeFiles = append(removeFiles, s3manager.BatchDeleteObject{
			Object: &s3.DeleteObjectInput{
				Key:    aws.String(s3key(fm.TokenID, fm.ID, rm.SHA256)),
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

	// goooo
	if err := tx.Commit().Error; err != nil {
		return err
	}

	if err := s.batcher.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{
		Objects: removeFiles,
	}); err != nil {
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

func CountFilesPerToken(db *gorm.DB, t *Token) (uint64, error) {
	c := uint64(0)
	if err := db.Model(&FileVersion{}).Where("token_id = ?", t.ID).Count(&c).Error; err != nil {
		return 0, err
	}
	return c, nil
}

func FindAndDecodeFile(s *Store, db *gorm.DB, t *Token, localFile *LocalFile) (*FileVersion, io.Reader, error) {
	fv, fm, err := FindFile(db, t, localFile.OriginFullPath)
	if err != nil {
		return nil, nil, err
	}

	// XXX: saving the file locally so we can download it concurrently
	// Download() takes WriterAt because of the chunks and even
	// though we can use inmemory buffer, it might get too big
	// so just save the file on disk and then delete it
	_, err = s.downloader.Download(localFile.File,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(s3key(fm.TokenID, fm.ID, fv.SHA256)),
		})
	if err != nil {
		return nil, nil, err
	}

	_, err = localFile.File.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, err
	}

	reader, err := DecompressAndDecrypt(t.Salt, localFile.File)
	if err != nil {
		return nil, nil, err
	}
	return fv, reader, nil
}

type FileMetadataAndVersion struct {
	FileMetadata *FileMetadata
	Versions     []*FileVersion
}

func FileLine(fm *FileMetadata, fv *FileVersion) string {
	return fmt.Sprintf("%d\t%s\t%s@v%d\t%s\n", fv.Size, fv.CreatedAt.Format(time.ANSIC), fm.FullPath(), fv.ID, fv.SHA256)
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
			for _, v := range f.Versions {
				buf.WriteString(FileLine(f.FileMetadata, v))
			}
		}
		fmt.Fprintf(buf, "\n")
	}
	return buf.String()
}

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
		if err := db.Where("file_metadata_id = ?", fm.ID).Order("updated_at_ns ASC").Find(&versions).Error; err != nil {
			return nil, err
		}
		out = append(out, FileMetadataAndVersion{fm, versions})
	}

	return out, nil
}

func ListVersionsFile(db *gorm.DB, t *Token, fm *FileMetadata) ([]*FileVersion, error) {
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
				Key:    aws.String(s3key(v.TokenID, v.FileMetadataID, v.SHA256)),
				Bucket: aws.String(s.bucket),
			},
		})
	}

	// delete token
	if err := tx.Delete(Token{}, "id = ?", token.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	if err := s.batcher.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{
		Objects: removeFiles,
	}); err != nil {
		return err
	}

	return nil
}

func SaveFile(s *Store, db *gorm.DB, t *Token, localFile *LocalFile) (*FileVersion, *FileMetadata, error) {
	dir, name := split(localFile.OriginFullPath)
	fm := &FileMetadata{}
	if err := db.FirstOrCreate(&fm, FileMetadata{TokenID: t.ID, Path: dir, Filename: name}).Error; err != nil {
		return nil, nil, err
	}
	sha := localFile.SHA
	size := localFile.Size

	// create file origin
	fv := &FileVersion{}
	res := db.Where("token_id = ? AND sha256 = ? AND file_metadata_id =?", t.ID, sha, fm.ID).Take(fv)
	if res.RecordNotFound() {
		// create new one
		fv.SHA256 = sha
		fv.Size = uint64(size)
		fv.FileMetadataID = fm.ID
		fv.TokenID = t.ID
		fv.DuplicatedSave = 0
		fv.UpdatedAtNs = uint64(time.Now().UnixNano())
		if err := db.Save(fv).Error; err != nil {
			return nil, nil, err
		}

		// only count once if file is created
		t.SizeUsed += fv.Size
		if err := db.Save(t).Error; err != nil {
			return nil, nil, err
		}

		if err := db.Save(fv).Error; err != nil {
			return nil, nil, err
		}
	} else {
		fv.DuplicatedSave++
		fv.UpdatedAtNs = uint64(time.Now().UnixNano())
		if err := db.Save(fv).Error; err != nil {
			return nil, nil, err
		}

		return fv, fm, nil
	}

	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3key(fm.TokenID, fm.ID, sha)),
		Body:   localFile.File,
	})
	if err != nil {
		return nil, nil, err
	}

	// create file metadata if we did not create it

	// check how many versions we have of this file
	tx := db.Begin()
	versions, err := ListVersionsFile(tx, t, fm)
	if err != nil {
		return nil, nil, err
	}

	limit := int(t.NumberOfArchives)
	removeFiles := []s3manager.BatchDeleteObject{}
	if len(versions) > limit {
		toDelete := versions[0:(len(versions) - limit)]
		for _, rm := range toDelete {
			// can get negative
			t.SizeUsed -= rm.Size

			removeFiles = append(removeFiles, s3manager.BatchDeleteObject{
				Object: &s3.DeleteObjectInput{
					Key:    aws.String(s3key(fm.TokenID, fm.ID, rm.SHA256)),
					Bucket: aws.String(s.bucket),
				},
			})

			if err := tx.Delete(rm).Error; err != nil {
				tx.Rollback()
				return nil, nil, err
			}
		}
	}

	// goooo
	if err := tx.Commit().Error; err != nil {
		// well at ths point we might have the file already saved..
		return nil, nil, err
	}

	log.Infof("removing limit: %d, versions: %d %+v", limit, len(versions), removeFiles)
	if err := s.batcher.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{
		Objects: removeFiles,
	}); err != nil {
		return nil, nil, err
	}

	return fv, fm, nil
}
