package ui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func readyModel(t *testing.T) UserRequestPromptModel {
	t.Helper()
	m := NewUserRequestPromptModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return updated.(UserRequestPromptModel)
}

func sendKey(m UserRequestPromptModel, key string) (UserRequestPromptModel, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(UserRequestPromptModel), cmd
}

func sendSpecialKey(m UserRequestPromptModel, keyType tea.KeyType) (UserRequestPromptModel, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyMsg{Type: keyType})
	return updated.(UserRequestPromptModel), cmd
}

func TestUserRequestPromptModel_InitReturnsBlinkCmd(t *testing.T) {
	m := NewUserRequestPromptModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a non-nil command")
	}
}

func TestUserRequestPromptModel_EmptyInputShowsError(t *testing.T) {
	m := readyModel(t)

	m, _ = sendSpecialKey(m, tea.KeyEnter)

	if m.errorMessage == "" {
		t.Error("expected error message for empty input")
	}
	if m.confirmed {
		t.Error("should not be confirmed on empty input")
	}
}

func TestUserRequestPromptModel_WhitespaceOnlyInputShowsError(t *testing.T) {
	m := readyModel(t)

	// 공백 입력
	for _, r := range "   " {
		m, _ = sendKey(m, string(r))
	}

	m, _ = sendSpecialKey(m, tea.KeyEnter)

	if m.errorMessage == "" {
		t.Error("expected error message for whitespace-only input")
	}
	if m.confirmed {
		t.Error("should not be confirmed on whitespace-only input")
	}
}

func TestUserRequestPromptModel_ValidInputConfirms(t *testing.T) {
	m := readyModel(t)

	for _, r := range "implement login feature" {
		m, _ = sendKey(m, string(r))
	}

	m, cmd := sendSpecialKey(m, tea.KeyEnter)

	if !m.confirmed {
		t.Error("should be confirmed after valid input")
	}
	if m.confirmedText != "implement login feature" {
		t.Errorf("expected 'implement login feature', got %q", m.confirmedText)
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestUserRequestPromptModel_CtrlCCancels(t *testing.T) {
	m := readyModel(t)

	m, cmd := sendSpecialKey(m, tea.KeyCtrlC)

	if m.confirmed {
		t.Error("should not be confirmed after ctrl+c")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestUserRequestPromptModel_ResultAfterConfirm(t *testing.T) {
	m := readyModel(t)

	for _, r := range "my request" {
		m, _ = sendKey(m, string(r))
	}
	m, _ = sendSpecialKey(m, tea.KeyEnter)

	result := m.Result()
	if result.Cancelled {
		t.Error("result should not be cancelled")
	}
	if result.Text != "my request" {
		t.Errorf("expected text 'my request', got %q", result.Text)
	}
}

func TestUserRequestPromptModel_ResultAfterCancel(t *testing.T) {
	m := readyModel(t)
	m, _ = sendSpecialKey(m, tea.KeyCtrlC)

	result := m.Result()
	if !result.Cancelled {
		t.Error("result should be cancelled")
	}
}

func TestUserRequestPromptModel_ErrorClearsOnInput(t *testing.T) {
	m := readyModel(t)

	// 빈 입력으로 에러 발생
	m, _ = sendSpecialKey(m, tea.KeyEnter)
	if m.errorMessage == "" {
		t.Fatal("expected error message")
	}

	// 일반 키 입력으로 에러 클리어
	m, _ = sendKey(m, "a")
	if m.errorMessage != "" {
		t.Errorf("error message should be cleared after input, got %q", m.errorMessage)
	}
}

func TestUserRequestPromptModel_ViewShowsLabelAndHelp(t *testing.T) {
	m := readyModel(t)
	view := m.View()
	plain := stripANSI(view)

	if !strings.Contains(plain, "Enter your request:") {
		t.Error("view should contain label text")
	}
	if !strings.Contains(plain, "Enter to confirm") {
		t.Error("view should contain help text about Enter")
	}
	if !strings.Contains(plain, "Ctrl+G") {
		t.Error("view should contain help text about Ctrl+G")
	}
}

func TestUserRequestPromptModel_ViewShowsInitializingBeforeReady(t *testing.T) {
	m := NewUserRequestPromptModel()
	view := m.View()

	if view != "Initializing..." {
		t.Errorf("expected 'Initializing...', got %q", view)
	}
}

func TestUserRequestPromptModel_ViewShowsErrorMessage(t *testing.T) {
	m := readyModel(t)

	m, _ = sendSpecialKey(m, tea.KeyEnter)

	view := m.View()
	plain := stripANSI(view)

	if !strings.Contains(plain, "Please enter your request.") {
		t.Error("view should contain error message")
	}
}

func TestUserRequestPromptModel_WindowSizeUpdatesWidth(t *testing.T) {
	m := NewUserRequestPromptModel()

	if m.ready {
		t.Error("should not be ready before WindowSizeMsg")
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = updated.(UserRequestPromptModel)

	if !m.ready {
		t.Error("should be ready after WindowSizeMsg")
	}
	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
}

func TestUserRequestPromptModel_EditorFinishedUpdatesTextarea(t *testing.T) {
	m := readyModel(t)

	tmpFile, err := os.CreateTemp("", "test-editor-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "edited in editor"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	m.tempFilePath = tmpFile.Name()

	updated, _ := m.Update(editorFinishedMsg{err: nil})
	m = updated.(UserRequestPromptModel)

	if m.textarea.Value() != "edited in editor" {
		t.Errorf("expected textarea to contain 'edited in editor', got %q", m.textarea.Value())
	}
	if m.errorMessage != "" {
		t.Errorf("expected no error, got %q", m.errorMessage)
	}
}

func TestUserRequestPromptModel_EditorFinishedWithError(t *testing.T) {
	m := readyModel(t)

	updated, _ := m.Update(editorFinishedMsg{err: os.ErrNotExist})
	m = updated.(UserRequestPromptModel)

	if m.errorMessage == "" {
		t.Error("expected error message when editor fails")
	}
	if !strings.Contains(m.errorMessage, "Editor failed") {
		t.Errorf("error message should mention editor failure, got %q", m.errorMessage)
	}
}
