package file

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"strings"

	"time"

	judoc "github.com/jackdoe/judoc/client"
	"github.com/minio/sio"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Store struct {
	judoc *judoc.Client
}

// make sure judoc is listening on localhost
// otherwise we have to do ssl
func NewStore(j string) (*Store, error) {
	return &Store{
		judoc: judoc.NewClient(j, nil),
	}, nil
}

func GetStoreId(prefix uint64) string {
	return fmt.Sprintf("%d.%d.%s", prefix, time.Now().UnixNano(), strings.ToLower(uuid.Must(uuid.NewV4()).String()))
}

// this leak of FileVersion here is not needed, but it is nice for logging
func (s *Store) RemoveMany(namespace string, remove []FileVersion) error {
	for _, rm := range remove {
		log.Infof("removing %d %d %d %s", rm.ID, rm.TokenID, rm.FileMetadataID, rm.StoreID)
		err := s.DeleteFile(namespace, rm.StoreID)
		if err != nil {
			log.Warnf("error deleting %s: %s", rm.StoreID, err.Error())
			return err
		}
	}
	return nil
}

func (s *Store) DownloadFile(key string, namespace, id string) (io.Reader, error) {
	obj, err := s.judoc.Get(namespace, id)
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

func (s *Store) UploadFile(key string, namespace string, id string, body io.Reader) (string, int64, error) {
	counter := &StreamByteCounter{
		sha256: sha256.New(),
	}
	tee := io.TeeReader(body, counter)
	reader, err := sio.EncryptReader(tee, sio.Config{Key: []byte(key)})
	if err != nil {
		return "", 0, err
	}
	t0 := time.Now()
	err = s.judoc.Set(namespace, id, reader)
	log.Infof("took: %dms to create %d sized file", time.Now().Sub(t0).Nanoseconds()/int64(1000000), counter.count)
	// report on the actual size, not the encrypted size
	return counter.Sum(), counter.count, err
}

func (s *Store) DeleteFile(namespace, id string) error {
	return s.judoc.Delete(namespace, id)
}
