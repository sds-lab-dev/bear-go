package claudecode

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestProcessStream_AssistantMessageCallsCallback(t *testing.T) {
	input := `{"type":"assistant","content":[{"type":"text","text":"hello"}]}
{"type":"result","subtype":"success","structured_output":{"questions":[]}}
`
	var called bool
	var received StreamMessage
	callback := func(msg StreamMessage) {
		called = true
		received = msg
	}

	_, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("callback was not called for assistant message")
	}
	if received.Type != StreamEventTypeAssistant {
		t.Errorf("expected type assistant, got %q", received.Type)
	}
	if len(received.Content) != 1 || received.Content[0].Text != "hello" {
		t.Errorf("unexpected content: %+v", received.Content)
	}
}

func TestProcessStream_UserMessageCallsCallback(t *testing.T) {
	input := `{"type":"user","content":[{"type":"tool_result","content":"done"}]}
{"type":"result","subtype":"success","structured_output":{"questions":[]}}
`
	var called bool
	var received StreamMessage
	callback := func(msg StreamMessage) {
		called = true
		received = msg
	}

	_, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("callback was not called for user message")
	}
	if received.Type != StreamEventTypeUser {
		t.Errorf("expected type user, got %q", received.Type)
	}
}

func TestProcessStream_ResultMessageNotPassedToCallback(t *testing.T) {
	input := `{"type":"result","subtype":"success","structured_output":{"questions":["q1"]}}
`
	callCount := 0
	callback := func(msg StreamMessage) {
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
	callback := func(msg StreamMessage) {
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
	callback := func(msg StreamMessage) {
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
	callback := func(msg StreamMessage) {}

	result, err := processStream(strings.NewReader(input), callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output struct {
		Questions []string `json:"questions"`
	}
	if err := json.Unmarshal(result.Output, &output); err != nil {
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
	callback := func(msg StreamMessage) {}

	_, err := processStream(strings.NewReader(input), callback)
	if !errors.Is(err, ErrResultError) {
		t.Fatalf("expected ErrResultError, got: %v", err)
	}
}

func TestProcessStream_InvalidJSONReturnsError(t *testing.T) {
	input := `not valid json
`
	callback := func(msg StreamMessage) {}

	_, err := processStream(strings.NewReader(input), callback)
	if !errors.Is(err, ErrStreamParseFailed) {
		t.Fatalf("expected ErrStreamParseFailed, got: %v", err)
	}
}

func TestProcessStream_NoResultReturnsError(t *testing.T) {
	input := `{"type":"assistant","content":[{"type":"text","text":"thinking..."}]}
`
	callback := func(msg StreamMessage) {}

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
	callback := func(msg StreamMessage) {
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
