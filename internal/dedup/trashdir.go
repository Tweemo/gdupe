package dedup

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// TrashDir finds the location of the trash/recycle bin directory depending on the OS used.
func TrashDir() (string, error) {
	sys := runtime.GOOS

	if sys == "darwin" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(h, ".Trash"), nil
	}

	if sys == "linux" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p := filepath.Join(h, ".local", "share", "Trash", "files")
		err = os.MkdirAll(p, 0750)
		if err != nil {
			return "", err
		}

		return p, nil
	}

	return "", errors.New("unsupported OS: only macOS and Linux are supported")
}
