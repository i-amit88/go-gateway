package session

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func ValidateWorkingDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return ErrInvalidWorkingDir
	}

	if !filepath.IsAbs(path) {
		return ErrInvalidWorkingDir
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrInvalidWorkingDir
		}

		return err
	}

	if !info.IsDir() {
		return ErrInvalidWorkingDir
	}

	return nil
}
