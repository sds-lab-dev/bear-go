package claudecode

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrProcessStartFailed = errors.New("failed to start claude process")
	ErrProcessExitError   = errors.New("claude process exited with error")
)

const toolList = "AskUserQuestion,Bash,TaskOutput,Edit,ExitPlanMode,Glob,Grep," +
	"KillShell,MCPSearch,Read,Skill,Task,TaskCreate,TaskGet,TaskList,TaskUpdate," +
	"WebFetch,WebSearch,Write,LSP"

type Client struct {
	apiKey     string
	workingDir string
	binaryPath string
	sessionID  string
}

func NewClient(apiKey, workingDir string) (*Client, error) {
	binaryPath, err := findClaudeBinary(
		exec.LookPath,
		os.UserHomeDir,
		fileExists,
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:     apiKey,
		workingDir: workingDir,
		binaryPath: binaryPath,
	}, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (c *Client) Query(systemPrompt, userPrompt, jsonSchema string, callback StreamCallback) (ResultData, error) {
	tmpFile, err := os.CreateTemp("", "bear-system-prompt-*.md")
	if err != nil {
		return ResultData{}, fmt.Errorf("failed to create temp file for system prompt: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(systemPrompt); err != nil {
		tmpFile.Close()
		return ResultData{}, fmt.Errorf("failed to write system prompt to temp file: %w", err)
	}
	tmpFile.Close()

	cmd := buildCommand(c, tmpFile.Name(), jsonSchema)
	cmd.Stdin = strings.NewReader(userPrompt)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ResultData{}, fmt.Errorf("%w: %v", ErrProcessStartFailed, err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ResultData{}, fmt.Errorf("%w: %v", ErrProcessStartFailed, err)
	}

	if err := cmd.Start(); err != nil {
		return ResultData{}, fmt.Errorf("%w: %v", ErrProcessStartFailed, err)
	}

	var stderrBuf strings.Builder
	go func() {
		io.Copy(&stderrBuf, stderr)
	}()

	result, streamErr := processStream(stdout, callback)

	waitErr := cmd.Wait()

	if streamErr != nil {
		return ResultData{}, streamErr
	}

	if waitErr != nil {
		return ResultData{}, fmt.Errorf("%w: %s", ErrProcessExitError, stderrBuf.String())
	}

	return result, nil
}

func buildCommand(c *Client, systemPromptPath, jsonSchema string) *exec.Cmd {
	args := []string{
		"-p",
		"--model", "claude-opus-4-6",
		"--output-format", "stream-json",
		"--verbose",
		"--include-partial-messages",
		"--allow-dangerously-skip-permissions",
		"--permission-mode", "bypassPermissions",
		"--tools", toolList,
		"--append-system-prompt-file", systemPromptPath,
		"--json-schema", jsonSchema,
	}

	if c.sessionID == "" {
		c.sessionID = uuid.New().String()
		args = append(args, "--session-id", c.sessionID)
	} else {
		args = append(args, "--resume", c.sessionID)
	}

	cmd := exec.Command(c.binaryPath, args...)
	cmd.Dir = c.workingDir
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY="+c.apiKey,
		"CLAUDE_CODE_EFFORT_LEVEL=high",
		"CLAUDE_CODE_DISABLE_AUTO_MEMORY=0",
		"CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY=1",
	)

	return cmd
}
