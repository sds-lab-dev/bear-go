package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrRelativePath = errors.New("path must be absolute; relative paths are not allowed")
	ErrPathNotExist = errors.New("path does not exist")
	ErrNotDirectory = errors.New("path is not a directory; please enter a directory path")
)

func ValidateWorkspacePath(path string) error {
	if !filepath.IsAbs(path) {
		return ErrRelativePath
	}

	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return ErrPathNotExist
	}
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return ErrNotDirectory
	}

	return nil
}
