package baxxio

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	. "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/config"
	"github.com/pierrec/lz4"
	"github.com/satori/go.uuid"
	"io"
	"os"
	"path"

	"time"
)

func SaveUploadedFile(key string, dest *os.File, body io.Reader) (string, int64, error) {
	sha := sha256.New()
	tee := io.TeeReader(body, sha)
	block, err := aes.NewCipher([]byte(key))
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
		return "", 0, err
	}

	// XXX: not to be trusted, attacker can flip bits
	// the only reason we encrypt is so we dont accidentally receive unencrypted data
	// or if someone steals the data

	// the stream does not buffer, so no explicit call is needed
	// flush however is needed here so users of the function can seek to the beginning
	lz4Writer.Flush()
	shasum := fmt.Sprintf("%x", sha.Sum(nil))

	return shasum, size, nil
}

func DecompressAndDecrypt(key string, r io.Reader) (io.Reader, error) {
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
