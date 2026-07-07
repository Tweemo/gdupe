package dedup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Delete removes every duplicate copy in the report from disk,
// leaving each group's keeper untouched.
func Delete(report *Report) error {
	var errs error
	for _, g := range report.Groups {
		for _, d := range g.Duplicates {
			err := os.Remove(d.Path)
			if err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	if errs != nil {
		return errs
	}

	return nil
}

// Move relocates every duplicate copy into a "duplicates" subfolder
// of dir, leaving each group's keeper untouched.
func Move(report *Report, dir string) error {
	err := os.MkdirAll(filepath.Join(dir, "duplicates"), 0750)
	if err != nil {
		return err
	}

	var errs error

	for _, g := range report.Groups {
		for _, d := range g.Duplicates {
			n := strings.ReplaceAll(d.Path, "/", "_")
			err := os.Rename(d.Path, filepath.Join(dir, "duplicates", n))
			if err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	if errs != nil {
		return errs
	}

	return nil
}
