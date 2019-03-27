package file

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"strings"

	"time"

	"github.com/minio/minio-go"
	"github.com/minio/sio"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Store struct {
	s3 *minio.Client
}

func NewStore(endpoint, key, secret string, disableSSL bool) (*Store, error) {
	svc, err := minio.New(endpoint, key, secret, !disableSSL)
	if err != nil {
		return nil, err
	}
	return &Store{
		s3: svc,
	}, nil
}

func GetStoreId(prefix uint64) string {
	return fmt.Sprintf("%d.%d.%s", prefix, time.Now().UnixNano(), strings.ToLower(uuid.Must(uuid.NewV4()).String()))
}

// this leak of FileVersion here is not needed, but it is nice for logging
func (s *Store) RemoveMany(bucket string, remove []FileVersion) error {
	for _, rm := range remove {
		log.Infof("removing %d %d %d %s", rm.ID, rm.TokenID, rm.FileMetadataID, rm.StoreID)
		err := s.DeleteFile(bucket, rm.StoreID)
		if err != nil {
			log.Warnf("error deleting %s: %s", rm.StoreID, err.Error())
			return err
		}
	}
	return nil
}

func (s *Store) DownloadFile(key string, bucket, id string) (io.Reader, error) {
	obj, err := s.s3.GetObject(bucket, id, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	reader, err := sio.DecryptReader(obj, sio.Config{
		Key: []byte(key),
	})
	if err != nil {
		return nil, err
	}
	return reader, nil
}

type StreamByteCounter struct {
	count  int64
	sha256 hash.Hash
}

func (s *StreamByteCounter) Sum() string {
	return fmt.Sprintf("%x", s.sha256.Sum(nil))
}

func (s *StreamByteCounter) Write(p []byte) (nn int, err error) {
	s.count += int64(len(p))
	_, err = s.sha256.Write(p)
	return len(p), err
}

func (s *Store) UploadFile(key string, bucket string, id string, body io.Reader) (string, int64, error) {
	counter := &StreamByteCounter{
		sha256: sha256.New(),
	}
	tee := io.TeeReader(body, counter)
	reader, err := sio.EncryptReader(tee, sio.Config{Key: []byte(key)})
	if err != nil {
		return "", 0, err
	}
	t0 := time.Now()
	_, err = s.s3.PutObject(bucket, id, reader, -1, minio.PutObjectOptions{})
	log.Infof("took: %dms to create %d sized file", time.Now().Sub(t0).Nanoseconds()/int64(1000000), counter.count)
	// report on the actual size, not the encrypted size
	return counter.Sum(), counter.count, err
}

func (s *Store) DeleteFile(bucket, id string) error {
	return s.s3.RemoveObject(bucket, id)
}

func (s *Store) MakeBucket(tokenBucket string) error {
	err := s.s3.MakeBucket(tokenBucket, "")
	return err
}

func (s *Store) ListObjects(tokenBucket string, err chan error, out chan string) {
	doneCh := make(chan struct{})
	defer close(doneCh)
	defer close(out)
	defer close(err)

	objectCh := s.s3.ListObjects(tokenBucket, "", false, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			err <- object.Err
			return
		}
		out <- object.Key
	}
}
