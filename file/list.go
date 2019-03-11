package file

import (
	"bytes"
	"fmt"
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
