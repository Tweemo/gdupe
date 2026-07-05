package dedup

import "errors"

// HashFile returns the SHA-256 digest of the file at path, as a hex string.
//
// TODO(you): implement.
//   - Open the file, stream it through crypto/sha256 with io.Copy —
//     do NOT read the whole file into memory (media files are big).
//   - hex.EncodeToString the sum.
//   - Tip: two files can only be duplicates if their sizes match, so
//     callers can skip hashing files with a unique size. (Optimization,
//     not required for correctness.)
func HashFile(path string) (string, error) {
	return "", errors.New("dedup.HashFile: not implemented")
}
