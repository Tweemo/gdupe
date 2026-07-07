// Package dedup finds exact duplicate files in a directory tree by
// comparing SHA-256 hashes, and can delete or relocate the extra copies.
package dedup

import (
	"fmt"
	"slices"
	"strings"
)

// Group is one set of byte-identical files: a keeper plus its copies.
type Group struct {
	Keeper     File   // the copy that is kept
	Duplicates []File // the redundant copies (never includes Keeper)
}

// Report is the result of a duplicate scan.
type Report struct {
	Groups []Group
}

// DuplicateCount returns the total number of redundant copies across
// all groups (the keeper in each group does not count).
func (r *Report) DuplicateCount() int {
	n := 0
	for _, g := range r.Groups {
		n += len(g.Duplicates)
	}
	return n
}

// WastedBytes returns the total size of all redundant copies.
func (r *Report) WastedBytes() int64 {
	var wb int64
	for _, g := range r.Groups {
		for _, f := range g.Duplicates {
			wb += f.Size
		}
	}
	return wb
}

// FindDuplicates groups byte-identical files together.
func FindDuplicates(files []File) (*Report, error) {
	sizeMap := map[int64][]File{}
	for _, f := range files {
		sizeMap[f.Size] = append(sizeMap[f.Size], f)
	}

	fileMap := map[string][]File{}
	for _, sameSize := range sizeMap {
		if len(sameSize) < 2 {
			continue
		}

		for _, f := range sameSize {
			fileHash, err := HashFile(f.Path)
			if err != nil {
				return nil, err
			}
			fileMap[fileHash] = append(fileMap[fileHash], f)
		}
	}

	report := Report{}
	for _, bucket := range fileMap {
		if len(bucket) > 1 {
			slices.SortFunc(bucket, func(a, b File) int {
				return strings.Compare(a.Path, b.Path)
			})
			g := Group{Keeper: bucket[0], Duplicates: bucket[1:]}
			report.Groups = append(report.Groups, g)
		}
	}

	return &report, nil
}

// FormatSize renders a byte count for humans, e.g. "1.4 MB".
func FormatSize(bytes int64) string {
	b := float64(bytes)
	c := 0 // counter

	unit := []string{"B", "KB", "MB", "GB", "TB", "PB"}

	for b >= 1024 {
		b = b / 1024
		c++
	}

	if c == 0 {
		return fmt.Sprintf("%d %s", bytes, unit[c])
	}

	return fmt.Sprintf("%.1f %s", b, unit[c])
}
