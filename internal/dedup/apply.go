package dedup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Delete moves every duplicate copy in the report to the system's Trash directory,
// leaving each group's keeper untouched.
func Delete(report *Report, dir string) error {
	return relocateDuplicates(report, dir)
}

// Move relocates every duplicate copy into a "duplicates" subfolder
// of dir, leaving each group's keeper untouched.
func Move(report *Report, dir string) error {
	dd := filepath.Join(dir, "duplicates")
	err := os.MkdirAll(dd, 0750)
	if err != nil {
		return err
	}

	return relocateDuplicates(report, dd)
}

// relocateDuplicates moves every duplicate in the report into dir, using collusion-safe names.
func relocateDuplicates(report *Report, dir string) error {
	var errs error

	for _, g := range report.Groups {
		for _, d := range g.Duplicates {
			n := strings.ReplaceAll(d.Path, "/", "_")
			err := os.Rename(d.Path, filepath.Join(dir, n))
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
