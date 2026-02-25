package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sds-lab-dev/bear-go/ai"
)

type mockSpecWriter struct{}

func (m *mockSpecWriter) GetInitialClarifyingQuestions(
	_ string,
) ([]string, error) {
	return nil, nil
}

func (m *mockSpecWriter) GetNextClarifyingQuestions(
	userAnswer string,
) ([]string, error) {
	return nil, nil
}

func (m *mockSpecWriter) DraftSpec() (string, error) {
	return "", nil
}

func (m *mockSpecWriter) ReviseSpec(_ string) (string, error) {
	return "", nil
}

func (m *mockSpecWriter) SetStreamCallbackHandler(_ func(ai.StreamMessage)) {
	// no-op for mock
}

func readySpecPromptModel(t *testing.T) SpecPromptModel {
	t.Helper()
	m := NewSpecPromptModel("test request", &mockSpecWriter{})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return updated.(SpecPromptModel)
}

func TestSpecPromptModel_InitReturnsCommand(t *testing.T) {
	m := NewSpecPromptModel("test request", &mockSpecWriter{})
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a non-nil command")
	}
}

func TestSpecPromptModel_StreamEventReturnsCommand(t *testing.T) {
	m := readySpecPromptModel(t)

	msg := streamEventMsg{
		StreamMessage: ai.StreamMessage{
			Type:    ai.StreamMessageTypeText,
			Content: "Analyzing requirements",
		},
	}

	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("expected command after stream event")
	}
}

func TestSpecPromptModel_ClarifyingQuestionsTransitionsState(t *testing.T) {
	m := readySpecPromptModel(t)

	msg := clarifyingQuestionsMsg{
		questions: []string{"What is the scope?", "Who is the primary user?"},
	}

	updated, cmd := m.Update(msg)
	m = updated.(SpecPromptModel)

	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "Please answer the clarifying questions") {
		t.Error("view should prompt user to answer clarifying questions")
	}
	if cmd == nil {
		t.Error("expected command for printing questions")
	}
}

func TestSpecPromptModel_ClarifyingQuestionsDoneTransitionsState(t *testing.T) {
	m := readySpecPromptModel(t)

	msg := clarifyingQuestionsDoneMsg{}

	updated, cmd := m.Update(msg)
	m = updated.(SpecPromptModel)

	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "Drafting the spec") {
		t.Error("view should show spec drafting message")
	}
	if cmd == nil {
		t.Error("expected command for printing no more questions message")
	}
}

func TestSpecPromptModel_ErrorReturnsResult(t *testing.T) {
	m := readySpecPromptModel(t)

	testErr := errors.New("connection failed")
	msg := streamErrorMsg{err: testErr}

	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected command after error")
	}

	resultMsg := cmd()
	result, ok := resultMsg.(SpecPromptResult)
	if !ok {
		t.Fatalf("expected SpecPromptResult, got %T", resultMsg)
	}
	if result.Err == nil {
		t.Fatal("expected error in result")
	}
	if result.Err.Error() != "connection failed" {
		t.Errorf("expected 'connection failed', got %q", result.Err.Error())
	}
}

func TestSpecPromptModel_ViewInInitialState(t *testing.T) {
	m := NewSpecPromptModel("test request", &mockSpecWriter{})
	view := m.View()
	plain := stripANSI(view)

	if !strings.Contains(plain, "Analyzing your request") {
		t.Errorf("expected view to contain 'Analyzing your request', got %q", plain)
	}
}

func TestSpecPromptModel_EnterWithEmptyTextShowsError(t *testing.T) {
	m := readySpecPromptModel(t)

	// Transition to WaitUserAnswers state
	updated, _ := m.Update(clarifyingQuestionsMsg{
		questions: []string{"What is the scope?"},
	})
	m = updated.(SpecPromptModel)

	// Press Enter with empty textarea
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(SpecPromptModel)

	if cmd != nil {
		t.Error("expected nil command when textarea is empty")
	}

	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "Please enter your feedback") {
		t.Error("view should show error message for empty input")
	}
}
