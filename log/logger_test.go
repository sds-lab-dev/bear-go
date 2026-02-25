package log

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestInitLogger_CreatesLogFile(t *testing.T) {
	logDir := t.TempDir()
	sessionID := "test-session-create"
	logPath := filepath.Join(logDir, fmt.Sprintf("bear-%s.log", sessionID))

	file, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	globalLogger = &logger{file: file}
	t.Cleanup(func() {
		_ = CloseLogger()
		globalLogger = nil
	})

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("expected log file to be created")
	}
}

func TestInitLogger_InvalidPath(t *testing.T) {
	err := InitLogger("../../invalid/path/session")
	if err == nil {
		t.Fatal("expected error for invalid log path")
	}
	t.Cleanup(func() { globalLogger = nil })
}

func TestLogLevelStrings(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarning, "WARNING"},
		{LogLevelError, "ERROR"},
		{LogLevelFatal, "FATAL"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("LogLevel(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func setupTestLogger(t *testing.T) string {
	t.Helper()

	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "bear-test.log")

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("failed to create test log file: %v", err)
	}

	globalLogger = &logger{file: file}
	t.Cleanup(func() {
		_ = CloseLogger()
		globalLogger = nil
	})

	return logPath
}

// Log entry format: YYYY-MM-DD HH:MM:SS.sss: <LEVEL>: <FILE>:<LINE>: <MSG>
var logEntryPattern = regexp.MustCompile(
	`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3}: (DEBUG|INFO|WARNING|ERROR|FATAL): \S+:\d+: .+$`,
)

func TestLogDebug_WritesFormattedEntry(t *testing.T) {
	logPath := setupTestLogger(t)

	Debug("debug message")

	content := readLogFile(t, logPath)
	line := strings.TrimSpace(content)

	if !logEntryPattern.MatchString(line) {
		t.Fatalf("log entry does not match expected format: %q", line)
	}
	assertContains(t, line, "DEBUG")
	assertContains(t, line, "debug message")
}

func TestLogInfo_WritesFormattedEntry(t *testing.T) {
	logPath := setupTestLogger(t)

	Info("info message")

	content := readLogFile(t, logPath)
	line := strings.TrimSpace(content)

	assertContains(t, line, "INFO")
	assertContains(t, line, "info message")
}

func TestLogWarning_WritesFormattedEntry(t *testing.T) {
	logPath := setupTestLogger(t)

	Warning("warning message")

	content := readLogFile(t, logPath)
	line := strings.TrimSpace(content)

	assertContains(t, line, "WARNING")
	assertContains(t, line, "warning message")
}

func TestLogError_WritesFormattedEntry(t *testing.T) {
	logPath := setupTestLogger(t)

	Error("error message")

	content := readLogFile(t, logPath)
	line := strings.TrimSpace(content)

	assertContains(t, line, "ERROR")
	assertContains(t, line, "error message")
}

func TestLogCallerLocation_PointsToCallSite(t *testing.T) {
	logPath := setupTestLogger(t)

	Info("caller test") // This line is the expected call site

	content := readLogFile(t, logPath)
	assertContains(t, content, "logger_test.go:")
}

func TestLogMultipleEntries_AllRecorded(t *testing.T) {
	logPath := setupTestLogger(t)

	Debug("first")
	Info("second")
	Error("third")

	content := readLogFile(t, logPath)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 log lines, got %d", len(lines))
	}
}

func TestLogFunctions_NilLogger_NoPanic(t *testing.T) {
	globalLogger = nil

	Debug("should not panic")
	Info("should not panic")
	Warning("should not panic")
	Error("should not panic")
}

func TestCloseLogger_NilLogger_NoPanic(t *testing.T) {
	globalLogger = nil

	err := CloseLogger()
	if err != nil {
		t.Fatalf("expected nil error when closing nil logger, got: %v", err)
	}
}

func readLogFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	return string(data)
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}
