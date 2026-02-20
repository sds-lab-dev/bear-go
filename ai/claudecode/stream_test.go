package claudecode

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/sds-lab-dev/bear-go/ai"
)

func TestProcessStream_AssistantMessageCallsCallback(t *testing.T) {
	input := `{"type":"assistant","content":[{"type":"text","text":"hello"}]}
{"type":"result","subtype":"success","structured_output":{"questions":[]}}
`
	var called bool
	callback := func(msg ai.StreamMessage) {
		called = true
	}

	_, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("callback was not called for assistant message")
	}
	// TODO: toStreamMessage 구현 후 Type, Content 검증 추가
}

func TestProcessStream_UserMessageCallsCallback(t *testing.T) {
	input := `{"type":"user","content":[{"type":"tool_result","content":"done"}]}
{"type":"result","subtype":"success","structured_output":{"questions":[]}}
`
	var called bool
	callback := func(msg ai.StreamMessage) {
		called = true
	}

	_, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("callback was not called for user message")
	}
}

func TestProcessStream_ResultMessageNotPassedToCallback(t *testing.T) {
	input := `{"type":"result","subtype":"success","structured_output":{"questions":["q1"]}}
`
	callCount := 0
	callback := func(msg ai.StreamMessage) {
		callCount++
	}

	_, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 0 {
		t.Errorf("callback should not be called for result messages, called %d times", callCount)
	}
}

func TestProcessStream_SystemMessageIgnored(t *testing.T) {
	input := `{"type":"system","subtype":"init"}
{"type":"result","subtype":"success","structured_output":{"questions":[]}}
`
	callCount := 0
	callback := func(msg ai.StreamMessage) {
		callCount++
	}

	_, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 0 {
		t.Errorf("callback should not be called for system messages, called %d times", callCount)
	}
}

func TestProcessStream_StreamEventIgnored(t *testing.T) {
	input := `{"type":"stream_event"}
{"type":"result","subtype":"success","structured_output":{"questions":[]}}
`
	callCount := 0
	callback := func(msg ai.StreamMessage) {
		callCount++
	}

	_, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 0 {
		t.Errorf("callback should not be called for stream_event messages, called %d times", callCount)
	}
}

func TestProcessStream_SuccessResultReturnsStructuredOutput(t *testing.T) {
	input := `{"type":"result","subtype":"success","structured_output":{"questions":["What is the scope?","Who is the user?"]}}
`
	callback := func(msg ai.StreamMessage) {}

	result, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output struct {
		Questions []string `json:"questions"`
	}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("failed to unmarshal result output: %v", err)
	}
	if len(output.Questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(output.Questions))
	}
	if output.Questions[0] != "What is the scope?" {
		t.Errorf("expected first question 'What is the scope?', got %q", output.Questions[0])
	}
}

func TestProcessStream_ErrorResultReturnsError(t *testing.T) {
	input := `{"type":"result","subtype":"error_max_turns","is_error":true}
`
	callback := func(msg ai.StreamMessage) {}

	_, err := processStream(strings.NewReader(input), callback)
	if !errors.Is(err, ErrResultError) {
		t.Fatalf("expected ErrResultError, got: %v", err)
	}
}

func TestProcessStream_InvalidJSONReturnsError(t *testing.T) {
	input := `not valid json
`
	callback := func(msg ai.StreamMessage) {}

	_, err := processStream(strings.NewReader(input), callback)
	if !errors.Is(err, ErrStreamParseFailed) {
		t.Fatalf("expected ErrStreamParseFailed, got: %v", err)
	}
}

func TestProcessStream_NoResultReturnsError(t *testing.T) {
	input := `{"type":"assistant","content":[{"type":"text","text":"thinking..."}]}
`
	callback := func(msg ai.StreamMessage) {}

	_, err := processStream(strings.NewReader(input), callback)
	if !errors.Is(err, ErrNoResultReceived) {
		t.Fatalf("expected ErrNoResultReceived, got: %v", err)
	}
}

func TestProcessStream_EmptyLinesSkipped(t *testing.T) {
	input := `
{"type":"assistant","content":[{"type":"text","text":"hello"}]}

{"type":"result","subtype":"success","structured_output":{"questions":[]}}

`
	callCount := 0
	callback := func(msg ai.StreamMessage) {
		callCount++
	}

	_, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected callback called once for assistant message, called %d times", callCount)
	}
}
