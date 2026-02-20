package claudecode

import (
	"errors"
	"testing"
)

func TestFindClaudeBinary_FoundInPath(t *testing.T) {
	lookPath := func(name string) (string, error) {
		return "/usr/bin/claude", nil
	}
	homeDir := func() (string, error) {
		return "/home/user", nil
	}
	fileExists := func(string) bool { return false }

	path, err := findClaudeBinary(lookPath, homeDir, fileExists)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if path != "/usr/bin/claude" {
		t.Errorf("expected /usr/bin/claude, got %q", path)
	}
}

func TestFindClaudeBinary_FoundInHomeFallback(t *testing.T) {
	lookPath := func(string) (string, error) {
		return "", errors.New("not found in PATH")
	}
	homeDir := func() (string, error) {
		return "/home/user", nil
	}
	fileExists := func(path string) bool {
		return path == "/home/user/.local/bin/claude"
	}

	path, err := findClaudeBinary(lookPath, homeDir, fileExists)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if path != "/home/user/.local/bin/claude" {
		t.Errorf("expected /home/user/.local/bin/claude, got %q", path)
	}
}

func TestFindClaudeBinary_FoundInAbsoluteFallback(t *testing.T) {
	lookPath := func(string) (string, error) {
		return "", errors.New("not found in PATH")
	}
	homeDir := func() (string, error) {
		return "/home/user", nil
	}
	fileExists := func(path string) bool {
		return path == "/usr/local/bin/claude"
	}

	path, err := findClaudeBinary(lookPath, homeDir, fileExists)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if path != "/usr/local/bin/claude" {
		t.Errorf("expected /usr/local/bin/claude, got %q", path)
	}
}

func TestFindClaudeBinary_FallbackOrder(t *testing.T) {
	lookPath := func(string) (string, error) {
		return "", errors.New("not found in PATH")
	}
	homeDir := func() (string, error) {
		return "/home/user", nil
	}

	// .npm-global/bin/claude와 .yarn/bin/claude 둘 다 존재하지만,
	// .npm-global이 먼저 탐색되므로 해당 경로가 반환되어야 한다.
	fileExists := func(path string) bool {
		return path == "/home/user/.npm-global/bin/claude" ||
			path == "/home/user/.yarn/bin/claude"
	}

	path, err := findClaudeBinary(lookPath, homeDir, fileExists)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if path != "/home/user/.npm-global/bin/claude" {
		t.Errorf("expected /home/user/.npm-global/bin/claude, got %q", path)
	}
}

func TestFindClaudeBinary_NotFound(t *testing.T) {
	lookPath := func(string) (string, error) {
		return "", errors.New("not found in PATH")
	}
	homeDir := func() (string, error) {
		return "/home/user", nil
	}
	fileExists := func(string) bool { return false }

	_, err := findClaudeBinary(lookPath, homeDir, fileExists)
	if !errors.Is(err, ErrBinaryNotFound) {
		t.Fatalf("expected ErrBinaryNotFound, got: %v", err)
	}
}

func TestFindClaudeBinary_NoHomeDir(t *testing.T) {
	lookPath := func(string) (string, error) {
		return "", errors.New("not found in PATH")
	}
	homeDir := func() (string, error) {
		return "", errors.New("home dir not available")
	}
	fileExists := func(path string) bool {
		return path == "/usr/local/bin/claude"
	}

	path, err := findClaudeBinary(lookPath, homeDir, fileExists)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if path != "/usr/local/bin/claude" {
		t.Errorf("expected /usr/local/bin/claude, got %q", path)
	}
}
