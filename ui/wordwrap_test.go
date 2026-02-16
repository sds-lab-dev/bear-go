package ui

import (
	"testing"
)

func TestWrapWords_NormalText(t *testing.T) {
	lines := wrapWords("hello world foo bar", 10)
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	for _, line := range lines {
		if len(line) > 10 {
			t.Errorf("line exceeds max width: %q (len=%d)", line, len(line))
		}
	}
}

func TestWrapWords_EmptyText(t *testing.T) {
	lines := wrapWords("", 10)
	if len(lines) != 0 {
		t.Fatalf("expected empty slice for empty text, got %v", lines)
	}
}

func TestWrapWords_ZeroWidth(t *testing.T) {
	lines := wrapWords("hello world", 0)
	if len(lines) != 0 {
		t.Fatalf("expected empty slice for zero width, got %v", lines)
	}
}

func TestWrapWords_NegativeWidth(t *testing.T) {
	lines := wrapWords("hello world", -5)
	if len(lines) != 0 {
		t.Fatalf("expected empty slice for negative width, got %v", lines)
	}
}

func TestWrapWords_SingleLongWord(t *testing.T) {
	lines := wrapWords("abcdefghijklmnop", 5)
	if len(lines) == 0 {
		t.Fatal("expected non-empty result for single long word")
	}
	// reflow/wordwrap does not break single words; it keeps them intact.
	joined := ""
	for _, l := range lines {
		joined += l
	}
	if joined != "abcdefghijklmnop" {
		t.Errorf("long word content lost: got %q", joined)
	}
}

func TestWrapWords_ExactFit(t *testing.T) {
	lines := wrapWords("hello", 5)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for exact fit, got %d: %v", len(lines), lines)
	}
	if lines[0] != "hello" {
		t.Errorf("expected %q, got %q", "hello", lines[0])
	}
}
