// Package dedup finds exact duplicate files in a directory tree by
// comparing SHA-256 hashes, and can delete or relocate the extra copies.
package dedup

import (
	"errors"
	"fmt"
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
//
// TODO(you): implement — sum the Size of every file in every
// group's Duplicates slice.
func (r *Report) WastedBytes() int64 {
	return 0
}

// FindDuplicates groups byte-identical files together.
//
// TODO(you): implement.
//   - Hash each file with HashFile and bucket files by hash
//     (a map[string][]File works well).
//   - Buckets with 2+ files become a Group; pick a keeper
//     deterministically (e.g. the lexicographically smallest path)
//     so runs are reproducible.
//   - Buckets with a single file are not duplicates — ignore them.
func FindDuplicates(files []File) (*Report, error) {
	return nil, errors.New("dedup.FindDuplicates: not implemented")
}

// FormatSize renders a byte count for humans, e.g. "1.4 MB".
//
// TODO(you): implement — divide by 1024 (or 1000, your call) until the
// value is small, then print with one decimal and the right unit.
func FormatSize(bytes int64) string {
	return fmt.Sprintf("%d B", bytes)
}
