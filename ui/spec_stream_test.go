package ui

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sds-lab-dev/bear-go/claudecode"
)

func readySpecStreamModel(t *testing.T, queryFunc func(claudecode.StreamCallback) (claudecode.ResultData, error)) SpecStreamModel {
	t.Helper()
	m := NewSpecStreamModel(queryFunc)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return updated.(SpecStreamModel)
}

func noopQueryFunc(_ claudecode.StreamCallback) (claudecode.ResultData, error) {
	return claudecode.ResultData{}, nil
}

func TestSpecStreamModel_InitStartsQuery(t *testing.T) {
	m := NewSpecStreamModel(noopQueryFunc)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a non-nil command")
	}
}

func TestSpecStreamModel_StreamEventUpdatesMessages(t *testing.T) {
	m := readySpecStreamModel(t, noopQueryFunc)

	msg := streamEventMsg{
		message: claudecode.StreamMessage{
			Type: claudecode.StreamEventTypeAssistant,
			Content: []claudecode.ContentBlock{
				{Type: "text", Text: "Analyzing requirements"},
			},
		},
	}

	updated, cmd := m.Update(msg)
	m = updated.(SpecStreamModel)

	if len(m.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(m.messages))
	}
	if m.messages[0] != "Analyzing requirements" {
		t.Errorf("expected 'Analyzing requirements', got %q", m.messages[0])
	}
	if cmd == nil {
		t.Error("expected command to wait for next message")
	}
}

func TestSpecStreamModel_MessagesTruncatedToMaxLines(t *testing.T) {
	m := readySpecStreamModel(t, noopQueryFunc)

	// maxDisplayLines(5)를 초과하는 메시지를 전송
	for i := range 7 {
		msg := streamEventMsg{
			message: claudecode.StreamMessage{
				Type: claudecode.StreamEventTypeAssistant,
				Content: []claudecode.ContentBlock{
					{Type: "text", Text: strings.Repeat("x", i+1)},
				},
			},
		}
		updated, _ := m.Update(msg)
		m = updated.(SpecStreamModel)
	}

	if len(m.messages) != maxDisplayLines {
		t.Errorf("expected %d messages, got %d", maxDisplayLines, len(m.messages))
	}
}

func TestSpecStreamModel_DoneWithQuestionsShowsQuestions(t *testing.T) {
	m := readySpecStreamModel(t, noopQueryFunc)

	questions := []string{"What is the scope?", "Who is the primary user?"}
	output, _ := json.Marshal(map[string][]string{"questions": questions})

	msg := streamDoneMsg{
		result: claudecode.ResultData{Output: json.RawMessage(output)},
	}

	updated, cmd := m.Update(msg)
	m = updated.(SpecStreamModel)

	if !m.done {
		t.Error("model should be done")
	}
	if m.result.Err != nil {
		t.Fatalf("unexpected error: %v", m.result.Err)
	}
	if len(m.result.Questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(m.result.Questions))
	}
	if m.result.Questions[0] != "What is the scope?" {
		t.Errorf("expected first question 'What is the scope?', got %q", m.result.Questions[0])
	}
	if cmd == nil {
		t.Error("expected quit command")
	}

	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "Clarification questions:") {
		t.Error("view should contain questions header")
	}
	if !strings.Contains(plain, "1. What is the scope?") {
		t.Error("view should contain first question")
	}
}

func TestSpecStreamModel_DoneWithEmptyQuestionsShowsInfo(t *testing.T) {
	m := readySpecStreamModel(t, noopQueryFunc)

	output, _ := json.Marshal(map[string][]string{"questions": {}})

	msg := streamDoneMsg{
		result: claudecode.ResultData{Output: json.RawMessage(output)},
	}

	updated, _ := m.Update(msg)
	m = updated.(SpecStreamModel)

	if !m.done {
		t.Error("model should be done")
	}
	if len(m.result.Questions) != 0 {
		t.Errorf("expected 0 questions, got %d", len(m.result.Questions))
	}

	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "No additional clarification questions needed.") {
		t.Error("view should contain info message for empty questions")
	}
}

func TestSpecStreamModel_ErrorShowsErrorMessage(t *testing.T) {
	m := readySpecStreamModel(t, noopQueryFunc)

	testErr := errors.New("connection failed")
	msg := streamErrorMsg{err: testErr}

	updated, cmd := m.Update(msg)
	m = updated.(SpecStreamModel)

	if !m.done {
		t.Error("model should be done")
	}
	if m.result.Err == nil {
		t.Fatal("expected error in result")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}

	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "connection failed") {
		t.Error("view should contain error message")
	}
}

func TestSpecStreamModel_CtrlCCancels(t *testing.T) {
	m := readySpecStreamModel(t, noopQueryFunc)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(SpecStreamModel)

	if !m.result.Cancelled {
		t.Error("result should be cancelled")
	}
	if !m.done {
		t.Error("model should be done")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestSpecStreamModel_ViewBeforeReady(t *testing.T) {
	m := NewSpecStreamModel(noopQueryFunc)
	view := m.View()

	if view != "Initializing..." {
		t.Errorf("expected 'Initializing...', got %q", view)
	}
}
