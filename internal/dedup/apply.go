package dedup

import "errors"

// Delete removes every duplicate copy in the report from disk,
// leaving each group's keeper untouched.
//
// TODO(you): implement.
//   - os.Remove each file in each group's Duplicates.
//   - Decide your error strategy: stop at the first failure, or keep
//     going and return a combined error (errors.Join)?
func Delete(report *Report) error {
	return errors.New("dedup.Delete: not implemented")
}

// Move relocates every duplicate copy into a "duplicates" subfolder
// of dir, leaving each group's keeper untouched.
//
// TODO(you): implement.
//   - Create dir/duplicates with os.MkdirAll.
//   - os.Rename each duplicate into it.
//   - Watch out: two duplicates from different subfolders can share a
//     base name (a/pic.jpg and b/pic.jpg) — you need a collision
//     strategy (numbered suffix, preserved subpath, ...).
func Move(report *Report, dir string) error {
	return errors.New("dedup.Move: not implemented")
}
