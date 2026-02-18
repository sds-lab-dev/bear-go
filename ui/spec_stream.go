package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sds-lab-dev/bear-go/claudecode"
)

const maxDisplayLines = 5

type SpecStreamResult struct {
	Questions []string
	Err       error
	Cancelled bool
}

type streamEventMsg struct {
	message claudecode.StreamMessage
}

type streamDoneMsg struct {
	result claudecode.ResultData
}

type streamErrorMsg struct {
	err error
}

type SpecStreamModel struct {
	messages  []string
	result    SpecStreamResult
	done      bool
	width     int
	ready     bool
	queryFunc func(claudecode.StreamCallback) (claudecode.ResultData, error)
	eventCh   chan tea.Msg
}

func NewSpecStreamModel(queryFunc func(claudecode.StreamCallback) (claudecode.ResultData, error)) SpecStreamModel {
	return SpecStreamModel{
		queryFunc: queryFunc,
		eventCh:   make(chan tea.Msg, 64),
	}
}

func (m SpecStreamModel) Init() tea.Cmd {
	return func() tea.Msg {
		go func() {
			callback := func(msg claudecode.StreamMessage) {
				m.eventCh <- streamEventMsg{message: msg}
			}

			result, err := m.queryFunc(callback)
			if err != nil {
				m.eventCh <- streamErrorMsg{err: err}
			} else {
				m.eventCh <- streamDoneMsg{result: result}
			}
		}()

		return <-m.eventCh
	}
}

func (m SpecStreamModel) waitForNext() tea.Cmd {
	return func() tea.Msg {
		return <-m.eventCh
	}
}

func (m SpecStreamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.result = SpecStreamResult{Cancelled: true}
			m.done = true
			return m, tea.Quit
		}

	case streamEventMsg:
		formatted := formatStreamMessage(msg.message)
		m.messages = append(m.messages, formatted...)
		if len(m.messages) > maxDisplayLines {
			m.messages = m.messages[len(m.messages)-maxDisplayLines:]
		}
		return m, m.waitForNext()

	case streamDoneMsg:
		questions, err := parseQuestions(msg.result.Output)
		if err != nil {
			m.result = SpecStreamResult{Err: err}
		} else {
			m.result = SpecStreamResult{Questions: questions}
		}
		m.done = true
		return m, tea.Quit

	case streamErrorMsg:
		m.result = SpecStreamResult{Err: msg.err}
		m.done = true
		return m, tea.Quit
	}

	return m, nil
}

func (m SpecStreamModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder

	if !m.done {
		b.WriteString(AgentActivityStyle.Render("Spec agent is analyzing your request..."))
		b.WriteByte('\n')
		for _, line := range m.messages {
			b.WriteString(DescriptionStyle.Render(line))
			b.WriteByte('\n')
		}
		return b.String()
	}

	if m.result.Err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", m.result.Err)))
		b.WriteByte('\n')
		return b.String()
	}

	if len(m.result.Questions) == 0 {
		b.WriteString(InfoStyle.Render("No additional clarification questions needed."))
		b.WriteByte('\n')
		return b.String()
	}

	b.WriteString(PromptLabelStyle.Render("Clarification questions:"))
	b.WriteByte('\n')
	for i, q := range m.result.Questions {
		b.WriteString(QuestionStyle.Render(fmt.Sprintf("%d. %s", i+1, q)))
		b.WriteByte('\n')
	}

	return b.String()
}

func (m SpecStreamModel) Result() SpecStreamResult {
	return m.result
}

func formatStreamMessage(msg claudecode.StreamMessage) []string {
	var parts []string

	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			if block.Text != "" {
				parts = append(parts, block.Text)
			}
		case "tool_use":
			parts = append(parts, fmt.Sprintf("[tool: %s]", block.Name))
		case "tool_result":
			if block.Content != "" {
				parts = append(parts, block.Content)
			}
		default:
			if block.Text != "" {
				parts = append(parts, block.Text)
			}
		}
	}

	joined := strings.Join(parts, " ")
	if joined == "" {
		return nil
	}

	lines := strings.Split(joined, "\n")
	if len(lines) > maxDisplayLines {
		lines = lines[:maxDisplayLines]
	}

	return lines
}

type clarificationOutput struct {
	Questions []string `json:"questions"`
}

func parseQuestions(data json.RawMessage) ([]string, error) {
	var output clarificationOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("failed to parse clarification output: %w", err)
	}
	return output.Questions, nil
}
