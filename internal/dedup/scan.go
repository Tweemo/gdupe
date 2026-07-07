package dedup

import (
	"io/fs"
	"path/filepath"
)

// File describes one regular file found during a scan.
type File struct {
	Path string // absolute or dir-relative path to the file
	Size int64  // size in bytes
}

// Scan walks dir recursively and returns every regular file in it.
func Scan(dir string) ([]File, error) {
	files := []File{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == "duplicates" && d.IsDir() {
			return fs.SkipDir
		}

		if !d.Type().IsRegular() {
			return nil
		}

		entryInfo, err := d.Info()
		if err != nil {
			return err
		}

		fileEntry := File{Path: path, Size: entryInfo.Size()}
		files = append(files, fileEntry)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
