package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelFatal
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

var globalLogger *logger

type logger struct {
	mu      sync.Mutex
	file    *os.File
	logPath string
}

func InitLogger(sessionID string) error {
	logPath := fmt.Sprintf("/var/log/bear-%s.log", sessionID)

	file, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}

	globalLogger = &logger{file: file, logPath: logPath}
	return nil
}

func CloseLogger() error {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.file.Close()
}

// callerSkip: runtime.Caller skip count from this function.
// Public log functions call writeLog with callerSkip=2 so that
// runtime.Caller resolves to the actual call site.
func (l *logger) writeLog(callerSkip int, level LogLevel, msg string) {
	_, file, line, ok := runtime.Caller(callerSkip)
	callerLocation := "unknown:0"
	if ok {
		callerLocation = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	entry := fmt.Sprintf("%s: %s: %s: %s\n", timestamp, level, callerLocation, msg)

	l.mu.Lock()
	defer l.mu.Unlock()

	_, _ = l.file.WriteString(entry)
}

func GetLogPath() string {
	if globalLogger == nil {
		return ""
	}
	return globalLogger.logPath
}

func Debug(msg string) {
	if globalLogger == nil {
		return
	}
	globalLogger.writeLog(2, LogLevelDebug, msg)
}

func Info(msg string) {
	if globalLogger == nil {
		return
	}
	globalLogger.writeLog(2, LogLevelInfo, msg)
}

func Warning(msg string) {
	if globalLogger == nil {
		return
	}
	globalLogger.writeLog(2, LogLevelWarning, msg)
}

func Error(msg string) {
	if globalLogger == nil {
		return
	}
	globalLogger.writeLog(2, LogLevelError, msg)
}

func Fatal(msg string) {
	if globalLogger == nil {
		os.Exit(1)
		return
	}
	globalLogger.writeLog(2, LogLevelFatal, msg)
	os.Exit(1)
}
