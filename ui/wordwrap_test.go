package ui

import (
	"strings"
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

	// 모든 단어가 순서대로 보존되어야 한다.
	joined := strings.Join(lines, " ")
	for _, word := range []string{"hello", "world", "foo", "bar"} {
		if !strings.Contains(joined, word) {
			t.Errorf("word %q lost after wrapping", word)
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

	lines = wrapWords("hello world", -5)
	if len(lines) != 0 {
		t.Fatalf("expected empty slice for negative width, got %v", lines)
	}
}

func TestWrapWords_SingleLongWord(t *testing.T) {
	lines := wrapWords("abcdefghijklmnop", 5)
	// reflow/wordwrap은 단일 단어를 분리하지 않으므로 1줄이어야 한다.
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for single long word, got %d: %v", len(lines), lines)
	}
	if lines[0] != "abcdefghijklmnop" {
		t.Errorf("expected %q, got %q", "abcdefghijklmnop", lines[0])
	}
}

func TestWrapWords_ExactFit(t *testing.T) {
	// 단일 단어가 정확히 폭에 맞는 경우
	lines := wrapWords("hello", 5)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for exact fit, got %d: %v", len(lines), lines)
	}
	if lines[0] != "hello" {
		t.Errorf("expected %q, got %q", "hello", lines[0])
	}

	// 여러 단어가 정확히 폭에 맞는 경우
	lines = wrapWords("hi there", 8)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for exact multi-word fit, got %d: %v", len(lines), lines)
	}
	if lines[0] != "hi there" {
		t.Errorf("expected %q, got %q", "hi there", lines[0])
	}
}
