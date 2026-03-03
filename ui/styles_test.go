package ui

import (
	"strings"
	"testing"
)

func TestTruncateToVisualLines_NoTruncationNeeded(t *testing.T) {
	text := "line1\nline2\nline3"
	result := truncateToVisualLines(text, 5, 80)
	if result != text {
		t.Errorf("expected no truncation, got %q", result)
	}
}

func TestTruncateToVisualLines_ExactlyAtLimit(t *testing.T) {
	text := "line1\nline2\nline3\nline4\nline5"
	result := truncateToVisualLines(text, 5, 80)
	if result != text {
		t.Errorf("expected no truncation for exactly 5 lines, got %q", result)
	}
}

func TestTruncateToVisualLines_TruncatesExcessLines(t *testing.T) {
	text := "line1\nline2\nline3\nline4\nline5\nline6"
	result := truncateToVisualLines(text, 5, 80)

	lines := strings.Split(result, "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d: %q", len(lines), result)
	}
	if lines[len(lines)-1] != "... (truncated)" {
		t.Errorf("expected truncation indicator, got %q", lines[len(lines)-1])
	}
}

func TestTruncateToVisualLines_LongLineWrapping(t *testing.T) {
	// A single line of 160 characters wraps to 2 visual lines at width 80.
	longLine := strings.Repeat("a", 160)
	text := longLine + "\nshort"
	// longLine = 2 visual lines, "short" = 1 visual line → total 3, within limit.
	result := truncateToVisualLines(text, 5, 80)
	if result != text {
		t.Errorf("expected no truncation for 3 visual lines, got %q", result)
	}
}

func TestTruncateToVisualLines_LongLineExceedsLimit(t *testing.T) {
	// A single line of 480 characters wraps to 6 visual lines at width 80.
	longLine := strings.Repeat("a", 480)
	result := truncateToVisualLines(longLine, 5, 80)

	lines := strings.Split(result, "\n")
	if lines[len(lines)-1] != "... (truncated)" {
		t.Errorf("expected truncation indicator, got %q", lines[len(lines)-1])
	}

	// Content part should be truncated to 4 visual lines = 320 characters.
	contentLine := lines[0]
	if len(contentLine) != 320 {
		t.Errorf("expected content truncated to 320 chars, got %d", len(contentLine))
	}
}

func TestTruncateToVisualLines_MixedWrappingAndNewlines(t *testing.T) {
	// 160 chars = 2 visual lines at width 80.
	longLine := strings.Repeat("b", 160)
	// longLine(2) + "c"(1) + "d"(1) = 4 visual lines, plus "e" would be 5.
	text := longLine + "\nc\nd\ne"
	result := truncateToVisualLines(text, 5, 80)
	if result != text {
		t.Errorf("expected no truncation for exactly 5 visual lines, got %q", result)
	}

	// Adding one more line should trigger truncation.
	text = longLine + "\nc\nd\ne\nf"
	result = truncateToVisualLines(text, 5, 80)
	if !strings.HasSuffix(result, "... (truncated)") {
		t.Errorf("expected truncation indicator, got %q", result)
	}
}

func TestTruncateToVisualLines_EmptyLines(t *testing.T) {
	// Each empty line counts as 1 visual line.
	text := "\n\n\n\n\n\n"
	result := truncateToVisualLines(text, 5, 80)

	lines := strings.Split(result, "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d: %q", len(lines), result)
	}
	if lines[len(lines)-1] != "... (truncated)" {
		t.Errorf("expected truncation indicator, got %q", lines[len(lines)-1])
	}
}

func TestTruncateToVisualLines_EmptyString(t *testing.T) {
	result := truncateToVisualLines("", 5, 80)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestTruncateToVisualLines_ZeroTerminalWidth(t *testing.T) {
	text := "line1\nline2"
	result := truncateToVisualLines(text, 5, 0)
	if result != text {
		t.Errorf("expected fallback to default width, got %q", result)
	}
}

func TestVisualLineCount(t *testing.T) {
	tests := []struct {
		name          string
		lineWidth     int
		terminalWidth int
		expected      int
	}{
		{"empty line", 0, 80, 1},
		{"short line", 40, 80, 1},
		{"exact width", 80, 80, 1},
		{"wraps once", 81, 80, 2},
		{"wraps twice", 161, 80, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visualLineCount(tt.lineWidth, tt.terminalWidth)
			if got != tt.expected {
				t.Errorf("visualLineCount(%d, %d) = %d, want %d",
					tt.lineWidth, tt.terminalWidth, got, tt.expected)
			}
		})
	}
}

func TestTruncateToVisualWidth(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		maxWidth int
		expected string
	}{
		{"within limit", "hello", 10, "hello"},
		{"exact limit", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello"},
		{"empty", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateToVisualWidth(tt.line, tt.maxWidth)
			if got != tt.expected {
				t.Errorf("truncateToVisualWidth(%q, %d) = %q, want %q",
					tt.line, tt.maxWidth, got, tt.expected)
			}
		})
	}
}
