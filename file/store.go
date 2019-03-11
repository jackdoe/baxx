package file

import (
	"io"

	. "github.com/jackdoe/baxx/config"
	"github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"
)

type Store struct {
	bucket string
	s3     *minio.Client
}

func NewStore(conf *StoreConfig) (*Store, error) {
	svc, err := minio.New(conf.Endpoint, conf.AccessKeyID, conf.SecretAccessKey, !conf.DisableSSL)
	if err != nil {
		return nil, err
	}
	return &Store{
		bucket: conf.Bucket,
		s3:     svc,
	}, nil
}

func (s *Store) removeMany(remove []FileVersion) {
	for _, rm := range remove {
		log.Infof("removing %d %d %d %s", rm.ID, rm.TokenID, rm.FileMetadataID, rm.StoreID)
		err := s.s3.RemoveObject(s.bucket, rm.StoreID)
		if err != nil {
			log.Warnf("error deleting %s: %s", rm.StoreID, err.Error())
		}
	}
}
func (s *Store) DownloadFile(fv *FileVersion) (io.Reader, error) {
	return s.s3.GetObject(s.bucket, fv.StoreID, minio.GetObjectOptions{})
}

func (s *Store) UploadFile(id string, reader io.Reader) (int64, error) {
	return s.s3.PutObject(s.bucket, id, reader, -1, minio.PutObjectOptions{})
}
