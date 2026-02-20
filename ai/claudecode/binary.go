package claudecode

import (
	"errors"
	"path/filepath"
)

var ErrBinaryNotFound = errors.New("claude binary not found in PATH or any fallback location")

var homeRelativeFallbackPaths = []string{
	".local/bin/claude",
	".npm-global/bin/claude",
	"node_modules/.bin/claude",
	".yarn/bin/claude",
	".claude/local/claude",
}

var absoluteFallbackPaths = []string{
	"/usr/local/bin/claude",
	"/usr/bin/claude",
}

func findClaudeBinary(
	lookPath func(string) (string, error),
	homeDir func() (string, error),
	fileExists func(string) bool,
) (string, error) {
	if path, err := lookPath("claude"); err == nil {
		return path, nil
	}

	if home, err := homeDir(); err == nil {
		for _, rel := range homeRelativeFallbackPaths {
			candidate := filepath.Join(home, rel)
			if fileExists(candidate) {
				return candidate, nil
			}
		}
	}

	for _, abs := range absoluteFallbackPaths {
		if fileExists(abs) {
			return abs, nil
		}
	}

	return "", ErrBinaryNotFound
}
