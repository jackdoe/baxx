package baxxio

import (
	"crypto/sha256"
	"fmt"
	. "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/config"
	"github.com/satori/go.uuid"
	"io"
	"os"
	"path"

	"time"
)

func SaveUploadedFile(key string, dest *os.File, body io.Reader) (string, int64, error) {
	sha := sha256.New()
	tee := io.TeeReader(body, sha)

	size, err := io.Copy(dest, tee)
	if err != nil {
		return "", 0, err
	}
	shasum := fmt.Sprintf("%x", sha.Sum(nil))

	return shasum, size, nil
}

func getTemporaryName(tokenid uint64) (string, error) {
	dir := path.Join(CONFIG.TemporaryRoot, "baxx")
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return "", nil
	}
	return path.Join(dir, fmt.Sprintf("%d.%d.%s", time.Now().UnixNano(), tokenid, uuid.Must(uuid.NewV4()))), nil
}

func CreateTemporaryFile(tokenid uint64, origin string) (*LocalFile, error) {
	tempName, err := getTemporaryName(tokenid)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(tempName, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	return &LocalFile{
		File:           file,
		OriginFullPath: origin,
		TempName:       tempName,
	}, nil
}
