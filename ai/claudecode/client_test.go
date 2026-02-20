package claudecode

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewClient_BinaryNotFound(t *testing.T) {
	lookPath := func(string) (string, error) {
		return "", errors.New("not found")
	}
	homeDir := func() (string, error) {
		return "/nonexistent-home-dir", nil
	}
	noFileExists := func(string) bool { return false }

	_, err := findClaudeBinary(lookPath, homeDir, noFileExists)
	if !errors.Is(err, ErrBinaryNotFound) {
		t.Fatalf("expected ErrBinaryNotFound, got: %v", err)
	}
}

func TestQuery_SessionIDGeneratedOnFirstCall(t *testing.T) {
	c := &Client{
		apiKey:     "test-key",
		workingDir: t.TempDir(),
		binaryPath: "/bin/echo",
	}

	promptFile := filepath.Join(t.TempDir(), "prompt.md")
	if err := os.WriteFile(promptFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create prompt file: %v", err)
	}

	cmd := buildCommand(c, promptFile, `{"type":"object"}`)

	if c.sessionID == "" {
		t.Fatal("sessionID should be generated after buildCommand")
	}

	hasSessionID := false
	for i, arg := range cmd.Args {
		if arg == "--session-id" && i+1 < len(cmd.Args) {
			hasSessionID = true
			break
		}
	}
	if !hasSessionID {
		t.Error("expected --session-id in first call arguments")
	}
}

func TestQuery_SessionIDReusedOnSubsequentCall(t *testing.T) {
	c := &Client{
		apiKey:     "test-key",
		workingDir: t.TempDir(),
		binaryPath: "/bin/echo",
	}

	promptFile := filepath.Join(t.TempDir(), "prompt.md")
	if err := os.WriteFile(promptFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create prompt file: %v", err)
	}

	buildCommand(c, promptFile, `{"type":"object"}`)
	firstSessionID := c.sessionID

	cmd2 := buildCommand(c, promptFile, `{"type":"object"}`)

	if c.sessionID != firstSessionID {
		t.Errorf("sessionID changed: %q -> %q", firstSessionID, c.sessionID)
	}

	hasResume := false
	for i, arg := range cmd2.Args {
		if arg == "--resume" && i+1 < len(cmd2.Args) && cmd2.Args[i+1] == firstSessionID {
			hasResume = true
			break
		}
	}
	if !hasResume {
		t.Error("expected --resume with existing sessionID in subsequent call")
	}
}

func TestQuery_EnvironmentVariablesSet(t *testing.T) {
	c := &Client{
		apiKey:     "sk-test-api-key",
		workingDir: t.TempDir(),
		binaryPath: "/bin/echo",
	}

	promptFile := filepath.Join(t.TempDir(), "prompt.md")
	if err := os.WriteFile(promptFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create prompt file: %v", err)
	}

	cmd := buildCommand(c, promptFile, `{"type":"object"}`)

	envMap := make(map[string]string)
	for _, env := range cmd.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	expected := map[string]string{
		"ANTHROPIC_API_KEY":                   "sk-test-api-key",
		"CLAUDE_CODE_EFFORT_LEVEL":            "high",
		"CLAUDE_CODE_DISABLE_AUTO_MEMORY":     "0",
		"CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY": "1",
	}

	for key, want := range expected {
		got, ok := envMap[key]
		if !ok {
			t.Errorf("expected environment variable %s to be set", key)
			continue
		}
		if got != want {
			t.Errorf("expected %s=%q, got %q", key, want, got)
		}
	}
}

func TestQuery_CLIArgumentsCorrect(t *testing.T) {
	c := &Client{
		apiKey:     "test-key",
		workingDir: t.TempDir(),
		binaryPath: "/usr/bin/claude",
	}

	promptFile := filepath.Join(t.TempDir(), "prompt.md")
	if err := os.WriteFile(promptFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create prompt file: %v", err)
	}

	cmd := buildCommand(c, promptFile, `{"type":"object"}`)

	requiredArgs := []string{
		"-p",
		"--model",
		"claude-opus-4-6",
		"--output-format",
		"stream-json",
		"--verbose",
		"--include-partial-messages",
		"--allow-dangerously-skip-permissions",
		"--permission-mode",
		"bypassPermissions",
		"--tools",
		"--append-system-prompt-file",
		"--json-schema",
	}

	argStr := strings.Join(cmd.Args, " ")
	for _, required := range requiredArgs {
		if !strings.Contains(argStr, required) {
			t.Errorf("expected argument %q in command args, got: %v", required, cmd.Args)
		}
	}

	if cmd.Dir != c.workingDir {
		t.Errorf("expected working dir %q, got %q", c.workingDir, cmd.Dir)
	}
}

func TestQuery_SystemPromptTempFileCleanup(t *testing.T) {
	c := &Client{
		apiKey:     "test-key",
		workingDir: t.TempDir(),
		binaryPath: "/bin/echo",
	}

	// Query 실행 전 기존 bear-system-prompt 파일 수를 기록한다.
	countBefore := countTempFiles(t, "bear-system-prompt-")

	// echo는 빈 출력으로 ErrNoResultReceived를 반환하지만, 임시 파일은 정리되어야 한다.
	type dummyOutput struct{}
	_, _ = query[dummyOutput](c, "test prompt", "test user prompt")

	countAfter := countTempFiles(t, "bear-system-prompt-")

	if countAfter > countBefore {
		t.Errorf("temp files increased from %d to %d; cleanup failed", countBefore, countAfter)
	}
}

func countTempFiles(t *testing.T, prefix string) int {
	t.Helper()
	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}
	count := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) {
			count++
		}
	}
	return count
}

func TestQuery_StdinReceivesUserPrompt(t *testing.T) {
	// 셸 스크립트를 생성하여 stdin을 파일로 캡처한다.
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "stdin_capture.txt")
	scriptFile := filepath.Join(tmpDir, "capture_stdin.sh")

	scriptContent := "#!/bin/sh\ncat > " + outputFile + "\n"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	c := &Client{
		apiKey:     "test-key",
		workingDir: tmpDir,
		binaryPath: scriptFile,
	}

	type dummyOutput struct{}
	userPrompt := "my user prompt text"
	_, _ = query[dummyOutput](c, "system prompt", userPrompt)

	captured, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read captured stdin: %v", err)
	}
	if string(captured) != userPrompt {
		t.Errorf("expected stdin %q, got %q", userPrompt, string(captured))
	}
}
