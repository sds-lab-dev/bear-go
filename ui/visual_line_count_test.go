package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCountVisualLines_EmptyString(t *testing.T) {
	got := countVisualLines("", 80)
	if got != 1 {
		t.Errorf("empty string: expected 1, got %d", got)
	}
}

func TestCountVisualLines_SingleShortLine(t *testing.T) {
	got := countVisualLines("hello", 80)
	if got != 1 {
		t.Errorf("short line: expected 1, got %d", got)
	}
}

func TestCountVisualLines_ExactWidth(t *testing.T) {
	got := countVisualLines("abcde", 5)
	if got < 1 {
		t.Errorf("exact width line: expected at least 1, got %d", got)
	}
}

func TestCountVisualLines_LongLineWraps(t *testing.T) {
	// "hello world" with width 6: "hello" (5) fits on line 1,
	// "world" wraps to line 2
	got := countVisualLines("hello world", 6)
	if got < 2 {
		t.Errorf("'hello world' at width 6: expected at least 2, got %d", got)
	}
}

func TestCountVisualLines_MultipleLogicalLines(t *testing.T) {
	got := countVisualLines("line1\nline2\nline3", 80)
	if got != 3 {
		t.Errorf("3 logical lines: expected 3, got %d", got)
	}
}

func TestCountVisualLines_LogicalLinesWithWrapping(t *testing.T) {
	// Two logical lines, each wrapping
	text := "hello world foo\nbar baz qux"
	got := countVisualLines(text, 6)
	if got < 4 {
		t.Errorf("two long logical lines at width 6: expected at least 4, got %d", got)
	}
}

func TestCountVisualLines_ZeroWidth(t *testing.T) {
	got := countVisualLines("hello\nworld", 0)
	if got != 2 {
		t.Errorf("zero width: expected 2 (logical lines only), got %d", got)
	}
}

func TestCountVisualLines_VeryLongWord(t *testing.T) {
	// A single word longer than the width should still produce multiple lines
	got := countVisualLines("abcdefghij", 5)
	if got < 2 {
		t.Errorf("long word at narrow width: expected at least 2, got %d", got)
	}
}

func TestCountVisualLines_GreaterThanLineCount(t *testing.T) {
	// This test verifies the core fix: visual lines >= logical lines
	text := "this is a very long line that definitely exceeds twenty characters"
	logicalLines := 1
	visualLines := countVisualLines(text, 20)
	if visualLines < logicalLines {
		t.Errorf("visual lines (%d) should be >= logical lines (%d)", visualLines, logicalLines)
	}
	if visualLines <= 1 {
		t.Errorf("long text at width 20 should wrap to multiple visual lines, got %d", visualLines)
	}
}

func TestCountWrappedLines_EmptyRunes(t *testing.T) {
	got := countWrappedLines([]rune{}, 80)
	if got != 1 {
		t.Errorf("empty runes: expected 1, got %d", got)
	}
}

func TestCountWrappedLines_SingleWord(t *testing.T) {
	got := countWrappedLines([]rune("hello"), 80)
	if got != 1 {
		t.Errorf("single word: expected 1, got %d", got)
	}
}

func TestCountWrappedLines_WordWrapAtBoundary(t *testing.T) {
	// "abc def" at width 4: "abc" fits, "def" wraps
	got := countWrappedLines([]rune("abc def"), 4)
	if got < 2 {
		t.Errorf("'abc def' at width 4: expected at least 2, got %d", got)
	}
}

func TestUserRequestPrompt_LongTextShowsAllLines(t *testing.T) {
	const termWidth = 20
	m := NewUserRequestPromptModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: termWidth, Height: 24})
	m = updated.(UserRequestPromptModel)

	// Type text that exceeds the terminal width, causing soft-wrap.
	longText := "the quick brown fox jumps over the lazy dog"
	for _, r := range longText {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(UserRequestPromptModel)
	}

	view := m.View()
	plain := stripANSI(view)

	// Every word from the input should appear in the rendered view.
	for _, word := range strings.Fields(longText) {
		if !strings.Contains(plain, word) {
			t.Errorf("view should contain %q but doesn't.\nFull view:\n%s", word, plain)
		}
	}

	// The textarea height should be greater than 1 for wrapped text.
	if m.textarea.Height() <= 1 {
		t.Errorf("textarea height should be > 1 for wrapped text, got %d", m.textarea.Height())
	}
}

func TestWorkspacePrompt_LongTextShowsAllLines(t *testing.T) {
	const termWidth = 20
	m := NewWorkspacePromptModel("/tmp", func(s string) error { return nil })
	updated, _ := m.Update(tea.WindowSizeMsg{Width: termWidth, Height: 24})
	m = updated.(WorkspacePromptModel)

	longText := "the quick brown fox jumps over the lazy dog"
	for _, r := range longText {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(WorkspacePromptModel)
	}

	view := m.View()
	plain := stripANSI(view)

	for _, word := range strings.Fields(longText) {
		if !strings.Contains(plain, word) {
			t.Errorf("view should contain %q but doesn't.\nFull view:\n%s", word, plain)
		}
	}

	if m.textarea.Height() <= 1 {
		t.Errorf("textarea height should be > 1 for wrapped text, got %d", m.textarea.Height())
	}
}
