package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
)

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
		if err := db.Where("file_metadata_id = ?", fm.ID).Find(&versions).Error; err != nil {
			return nil, err
		}
		out = append(out, FileMetadataAndVersion{fm, versions})
	}

	return out, nil
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
	total := uint64(0)
	for _, k := range keys {
		files := grouped[k]

		size := uint64(0)
		for _, f := range files {
			for _, v := range f.Versions {
				size += v.Size
				total += v.Size
			}
		}
		fmt.Fprintf(buf, "%s: total size: %d (%s)\n", k, size, prettySize(size))
		for _, f := range files {
			for _, v := range f.Versions {
				buf.WriteString(FileLine(f.FileMetadata, v))
			}
		}
		fmt.Fprintf(buf, "\n")
	}
	fmt.Fprintf(buf, "sum total size: %d (%s)\n", total, prettySize(total))
	return buf.String()
}

func ShaDiff(db *gorm.DB, t *Token, body io.Reader) (string, error) {
	scanner := bufio.NewScanner(body)
	buf := bytes.NewBufferString("")
	for scanner.Scan() {
		line := scanner.Text()
		splitted := strings.SplitN(line, "  ", 2)
		if len(splitted) != 2 || len(splitted[0]) != 64 || len(splitted[1]) == 0 {
			return "", fmt.Errorf("expected 'shasum(64 chars)  path/to/file' (two spaces), basically output of shasum -a 256 file; got: %s", line)
		}
		_, _, err := FindFileBySHA(db, t, splitted[0])
		if err != nil {
			fmt.Fprintf(buf, "%s\n", line)
		}
	}

	return buf.String(), nil
}
