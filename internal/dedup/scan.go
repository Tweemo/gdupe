package dedup

import "errors"

// File describes one regular file found during a scan.
type File struct {
	Path string // absolute or dir-relative path to the file
	Size int64  // size in bytes
}

// Scan walks dir recursively and returns every regular file in it.
//
// TODO(you): implement.
//   - Use filepath.WalkDir (or io/fs.WalkDir) to traverse dir.
//   - Skip directories, symlinks, and other non-regular files.
//   - Consider skipping the "duplicates" subfolder so a previous
//     -move run doesn't get re-scanned.
func Scan(dir string) ([]File, error) {
	return nil, errors.New("dedup.Scan: not implemented")
}
