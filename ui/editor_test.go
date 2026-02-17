package ui

import (
	"errors"
	"testing"
)

func TestResolveEditor_EditorEnvSet(t *testing.T) {
	lookupEnv := func(key string) (string, bool) {
		if key == "EDITOR" {
			return "vim", true
		}
		return "", false
	}
	commandExists := func(string) bool { return false }

	cmd, err := resolveEditor(lookupEnv, commandExists)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Executable != "vim" {
		t.Errorf("expected executable 'vim', got %q", cmd.Executable)
	}
	if len(cmd.Args) != 0 {
		t.Errorf("expected no args, got %v", cmd.Args)
	}
}

func TestResolveEditor_EditorEnvWithArgs(t *testing.T) {
	lookupEnv := func(key string) (string, bool) {
		if key == "EDITOR" {
			return "code --wait", true
		}
		return "", false
	}
	commandExists := func(string) bool { return false }

	cmd, err := resolveEditor(lookupEnv, commandExists)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Executable != "code" {
		t.Errorf("expected executable 'code', got %q", cmd.Executable)
	}
	if len(cmd.Args) != 1 || cmd.Args[0] != "--wait" {
		t.Errorf("expected args [--wait], got %v", cmd.Args)
	}
}

func TestResolveEditor_EditorEnvEmpty(t *testing.T) {
	lookupEnv := func(key string) (string, bool) {
		if key == "EDITOR" {
			return "", true
		}
		return "", false
	}
	commandExists := func(name string) bool {
		return name == "vi"
	}

	cmd, err := resolveEditor(lookupEnv, commandExists)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Executable != "vi" {
		t.Errorf("expected fallback to 'vi', got %q", cmd.Executable)
	}
}

func TestResolveEditor_FallbackToCode(t *testing.T) {
	lookupEnv := func(string) (string, bool) { return "", false }
	commandExists := func(name string) bool {
		return name == "code"
	}

	cmd, err := resolveEditor(lookupEnv, commandExists)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Executable != "code" {
		t.Errorf("expected executable 'code', got %q", cmd.Executable)
	}
	if len(cmd.Args) != 1 || cmd.Args[0] != "--wait" {
		t.Errorf("expected args [--wait], got %v", cmd.Args)
	}
}

func TestResolveEditor_FallbackToVi(t *testing.T) {
	lookupEnv := func(string) (string, bool) { return "", false }
	commandExists := func(name string) bool {
		return name == "vi"
	}

	cmd, err := resolveEditor(lookupEnv, commandExists)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Executable != "vi" {
		t.Errorf("expected executable 'vi', got %q", cmd.Executable)
	}
	if len(cmd.Args) != 0 {
		t.Errorf("expected no args, got %v", cmd.Args)
	}
}

func TestResolveEditor_NoEditorAvailable(t *testing.T) {
	lookupEnv := func(string) (string, bool) { return "", false }
	commandExists := func(string) bool { return false }

	_, err := resolveEditor(lookupEnv, commandExists)
	if !errors.Is(err, ErrNoEditorFound) {
		t.Errorf("expected ErrNoEditorFound, got %v", err)
	}
}
