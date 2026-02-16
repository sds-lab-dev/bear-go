package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateWorkspacePath_AbsoluteDirectoryPath(t *testing.T) {
	dir := t.TempDir()

	err := ValidateWorkspacePath(dir)
	if err != nil {
		t.Fatalf("expected nil error for valid absolute directory, got: %v", err)
	}
}

func TestValidateWorkspacePath_RelativePath(t *testing.T) {
	err := ValidateWorkspacePath("./some/path")
	if !errors.Is(err, ErrRelativePath) {
		t.Fatalf("expected ErrRelativePath, got: %v", err)
	}
}

func TestValidateWorkspacePath_NonExistentPath(t *testing.T) {
	err := ValidateWorkspacePath("/nonexistent/path/abc123")
	if !errors.Is(err, ErrPathNotExist) {
		t.Fatalf("expected ErrPathNotExist, got: %v", err)
	}
}

func TestValidateWorkspacePath_FilePath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "testfile.txt")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	f.Close()

	err = ValidateWorkspacePath(filePath)
	if !errors.Is(err, ErrNotDirectory) {
		t.Fatalf("expected ErrNotDirectory, got: %v", err)
	}
}

func TestValidateWorkspacePath_EmptyPath(t *testing.T) {
	err := ValidateWorkspacePath("")
	if !errors.Is(err, ErrRelativePath) {
		t.Fatalf("expected ErrRelativePath for empty path, got: %v", err)
	}
}
