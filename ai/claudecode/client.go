package claudecode

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/kaptinlin/jsonschema"
	"github.com/sds-lab-dev/bear-go/ai"
	"github.com/sds-lab-dev/bear-go/log"
)

var (
	ErrProcessStartFailed = errors.New("failed to start claude process")
	ErrProcessExitError   = errors.New("claude process exited with error")
)

const toolList = "AskUserQuestion,Bash,TaskOutput,Edit,ExitPlanMode,Glob,Grep,KillShell,MCPSearch,Read,Skill,Task,TaskCreate,TaskGet,TaskList,TaskUpdate,WebFetch,WebSearch,Write,LSP"

type Client struct {
	apiKey         string
	workingDir     string
	binaryPath     string
	sessionID      string
	sessionState   clientSessionState
	streamCallback func(ai.StreamMessage)
}

type clientSessionState int

const (
	sessionStateBegin clientSessionState = iota
	sessionStateWaitUserAnswers
	sessionStateNoClarifyingQuestions
	sessionStateWaitUserFeedback
	sessionStateSpecApproved
)

func NewClient(apiKey, workingDir string) (*Client, error) {
	binaryPath, err := findClaudeBinary(
		exec.LookPath,
		os.UserHomeDir,
		fileExists,
	)
	if err != nil {
		return nil, err
	}

	// Fallback to current working directory if no working directory is provided.
	if workingDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		workingDir = cwd
	}

	return &Client{
		apiKey:         apiKey,
		workingDir:     workingDir,
		binaryPath:     binaryPath,
		sessionState:   sessionStateBegin,
		streamCallback: nil,
	}, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (c *Client) GetInitialClarifyingQuestions(
	initialUserRequest string,
) ([]string, error) {
	if c.sessionState != sessionStateBegin {
		return nil, fmt.Errorf("unexpected session state for GetInitialClarifyingQuestions: %v", c.sessionState)
	}

	log.Debug(fmt.Sprintf("getting initial clarifying questions for user request: %v", initialUserRequest))
	type outputSchema struct {
		Questions []string `json:"questions" jsonschema:"required,minItems=0,maxItems=5"`
	}
	output, err := query[outputSchema](
		c,
		ai.ClarificationSystemPrompt(),
		ai.ClarificationUserPromptForInitialRequest(initialUserRequest),
	)
	if err != nil {
		err = fmt.Errorf("failed to get initial clarifying questions: %w", err)
		log.Error(err.Error())
		return nil, err
	}
	log.Debug(fmt.Sprintf("received initial clarifying questions: %#v", output))

	if len(output.Questions) == 0 {
		c.sessionState = sessionStateNoClarifyingQuestions
	} else {
		c.sessionState = sessionStateWaitUserAnswers
	}

	return output.Questions, nil
}

func (c *Client) GetNextClarifyingQuestions(userAnswer string) ([]string, error) {
	if c.sessionState != sessionStateWaitUserAnswers {
		return nil, fmt.Errorf("unexpected session state for GetNextClarifyingQuestions: %v", c.sessionState)
	}

	// TODO
	questions, err := []string{}, errors.New("GetNextClarifyingQuestions not implemented yet")
	if err != nil {
		return nil, err
	}

	if len(questions) == 0 {
		c.sessionState = sessionStateNoClarifyingQuestions
	} else {
		c.sessionState = sessionStateWaitUserAnswers
	}

	return questions, nil
}

func (c *Client) DraftSpec() (string, error) {
	if c.sessionState != sessionStateNoClarifyingQuestions {
		return "", fmt.Errorf("unexpected session state for DraftSpec: %v", c.sessionState)
	}

	// TODO
	return "", errors.New("DraftSpec not implemented yet")
}

func (c *Client) ReviseSpec(userFeedback string) (string, error) {
	if c.sessionState != sessionStateWaitUserFeedback {
		return "", fmt.Errorf("unexpected session state for ReviseSpec: %v", c.sessionState)
	}

	// TODO
	return "", errors.New("ReviseSpec not implemented yet")
}

func (c *Client) SetStreamCallbackHandler(handler func(ai.StreamMessage)) {
	c.streamCallback = handler
}

func query[T any](client *Client, systemPrompt, userPrompt string) (T, error) {
	var zeroValue T

	opts := &jsonschema.StructTagOptions{
		// Remove `$schema` from the generated schema.
		SchemaVersion: "",
		SchemaProperties: map[string]any{
			"additionalProperties": false,
		},
	}
	schema, err := jsonschema.FromStructWithOptions[T](opts)
	if err != nil {
		return zeroValue, fmt.Errorf("failed to generate JSON schema for type %T: %w", zeroValue, err)
	}
	// Remove `$defs` from the generated schema.
	schema.Defs = nil

	schemaString, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return zeroValue, fmt.Errorf("failed to marshal JSON schema for type %T: %w", zeroValue, err)
	}
	log.Debug(fmt.Sprintf("generated JSON schema for type %T: %s", zeroValue, string(schemaString)))

	tmpFile, err := os.CreateTemp("", "bear-system-prompt-*.md")
	if err != nil {
		return zeroValue, fmt.Errorf("failed to create temp file for system prompt: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(systemPrompt); err != nil {
		tmpFile.Close()
		return zeroValue, fmt.Errorf("failed to write system prompt to temp file: %w", err)
	}
	tmpFile.Close()

	cmd := buildCommand(client, tmpFile.Name(), string(schemaString))
	cmd.Stdin = strings.NewReader(userPrompt)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return zeroValue, fmt.Errorf("%w: %v", ErrProcessStartFailed, err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return zeroValue, fmt.Errorf("%w: %v", ErrProcessStartFailed, err)
	}

	if err := cmd.Start(); err != nil {
		return zeroValue, fmt.Errorf("%w: %v", ErrProcessStartFailed, err)
	}

	var stderrBuf strings.Builder
	go func() {
		io.Copy(&stderrBuf, stderr)
	}()

	log.Debug("starting Claude Code CLI process...")
	result, streamErr := processStream(stdout, client.streamCallback)
	if streamErr != nil {
		return zeroValue, streamErr
	}
	log.Debug(fmt.Sprintf("waiting Claude Code CLI process: result=%v, streamErr=%v", string(result), streamErr))
	waitErr := cmd.Wait()
	log.Debug(fmt.Sprintf("finished Claude Code CLI process: waitErr=%v", waitErr))
	if waitErr != nil {
		return zeroValue,
			fmt.Errorf("%w: %s", ErrProcessExitError, stderrBuf.String())
	}
	log.Debug(fmt.Sprintf("raw result from processStream: %v", string(result)))

	if validator := schema.Validate(result); !validator.IsValid() {
		err := validator.DetailedErrors()
		return zeroValue,
			fmt.Errorf("JSON schema validation failed for type %T: %v", zeroValue, err)
	}

	var finalResult T
	if err := json.Unmarshal(result, &finalResult); err != nil {
		return zeroValue,
			fmt.Errorf("failed to unmarshal processStream result into type %T: %w", zeroValue, err)
	}

	return finalResult, nil
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
		"CLAUDE_CODE_EFFORT_LEVEL=high",
		"CLAUDE_CODE_DISABLE_AUTO_MEMORY=0",
		"CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY=1",
	)
	// If an API key is provided, use the key to authenticate with the Anthropic
	// API.
	// If no API key is provided, the claude binary will attempt to use a
	// subscription plan if available, but this may fail if the key is required
	// for authentication, in which case an error will be returned.
	if c.apiKey != "" {
		cmd.Env = append(cmd.Env, "ANTHROPIC_API_KEY="+c.apiKey)
	}

	return cmd
}
