package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// HashFile returns the SHA-256 digest of the file at path, as a hex string.
func HashFile(path string) (string, error) {

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	digest := hex.EncodeToString(h.Sum(nil))

	return digest, nil
}
