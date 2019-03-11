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
	conf   *StoreConfig
}

func NewStore(conf *StoreConfig) (*Store, error) {
	svc, err := minio.New(conf.Endpoint, conf.AccessKeyID, conf.SecretAccessKey, !conf.DisableSSL)
	if err != nil {
		return nil, err
	}
	return &Store{
		conf:   conf,
		bucket: conf.Bucket,
		s3:     svc,
	}, nil
}

func (s *Store) removeMany(remove []FileVersion) error {
	for _, rm := range remove {
		log.Infof("removing %d %d %d %s", rm.ID, rm.TokenID, rm.FileMetadataID, rm.StoreID)

		err := s.DeleteFile(rm.StoreID)
		if err != nil {
			log.Warnf("error deleting %s: %s", rm.StoreID, err.Error())
			return err
		}
	}
	return nil
}
func (s *Store) DownloadFile(fv *FileVersion) (io.Reader, error) {
	return s.s3.GetObject(s.bucket, fv.StoreID, minio.GetObjectOptions{})
}

func (s *Store) UploadFile(id string, reader io.Reader) (int64, error) {
	return s.s3.PutObject(s.bucket, id, reader, -1, minio.PutObjectOptions{})
}

func (s *Store) DeleteFile(id string) error {
	return s.s3.RemoveObject(s.bucket, id)
}

func (s *Store) MakeBucket() error {
	err := s.s3.MakeBucket(s.bucket, "")
	return err
}

func (s *Store) ListObjects(err chan error, out chan string) {
	doneCh := make(chan struct{})
	defer close(doneCh)
	defer close(out)
	defer close(err)

	objectCh := s.s3.ListObjects(s.bucket, "", false, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			err <- object.Err
			return
		}
		out <- object.Key
	}
}
