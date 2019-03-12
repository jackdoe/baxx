package file

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"strings"

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

func GetStoreId(tokenID uint64) string {
	return fmt.Sprintf("%s.%s", tokenToBucketPrefix(tokenID), uuid.Must(uuid.NewV4()).String())
}

func tokenToBucketPrefix(id uint64) string {
	return fmt.Sprintf("baxxtoken%d", id)
}
func splitStoreID(id string) (string, string) {
	splitted := strings.Split(id, ".")
	return splitted[0], splitted[1]
}

func (s *Store) RemoveMany(remove []FileVersion) error {
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

func (s *Store) DownloadFile(key string, storeid string) (io.Reader, error) {
	bucket, id := splitStoreID(storeid)
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

func (s *Store) UploadFile(key string, storeid string, body io.Reader) (string, int64, error) {
	bucket, id := splitStoreID(storeid)
	counter := &StreamByteCounter{
		sha256: sha256.New(),
	}
	tee := io.TeeReader(body, counter)
	reader, err := sio.EncryptReader(tee, sio.Config{Key: []byte(key)})
	if err != nil {
		return "", 0, err
	}

	_, err = s.s3.PutObject(bucket, id, reader, -1, minio.PutObjectOptions{})
	// report on the actual size, not the encrypted size
	return counter.Sum(), counter.count, err
}

func (s *Store) DeleteFile(storeid string) error {
	bucket, id := splitStoreID(storeid)
	return s.s3.RemoveObject(bucket, id)
}

func (s *Store) MakeBucket(tokenID uint64) error {
	err := s.s3.MakeBucket(tokenToBucketPrefix(tokenID), "")
	return err
}

func (s *Store) ListObjects(tokenid uint64, err chan error, out chan string) {
	doneCh := make(chan struct{})
	defer close(doneCh)
	defer close(out)
	defer close(err)

	objectCh := s.s3.ListObjects(tokenToBucketPrefix(tokenid), "", false, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			err <- object.Err
			return
		}
		out <- object.Key
	}
}
