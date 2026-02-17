package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type WorkspacePromptResult struct {
	Path      string
	Cancelled bool
}

type WorkspacePromptModel struct {
	textarea      textarea.Model
	currentDir    string
	validatePath  func(string) error
	errorMessage  string
	confirmed     bool
	confirmedPath string
	width         int
	ready         bool
}

func NewWorkspacePromptModel(currentDir string, validatePath func(string) error) WorkspacePromptModel {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetHeight(1)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	ta.Focus()

	return WorkspacePromptModel{
		textarea:     ta,
		currentDir:   currentDir,
		validatePath: validatePath,
	}
}

func (m WorkspacePromptModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m WorkspacePromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.ready = true
		return m, nil
	}

	if !m.ready {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m WorkspacePromptModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		return m.handleEnter()
	case "alt+enter":
		m.textarea.InsertString("\n")
		m.errorMessage = ""
		m.textarea.SetHeight(m.textarea.LineCount())
		return m, nil
	}

	m.errorMessage = ""
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	m.textarea.SetHeight(m.textarea.LineCount())
	return m, cmd
}

func (m WorkspacePromptModel) handleEnter() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.textarea.Value())

	if value == "" {
		m.confirmedPath = m.currentDir
	} else {
		if err := m.validatePath(value); err != nil {
			m.errorMessage = err.Error()
			return m, nil
		}
		m.confirmedPath = value
	}

	m.confirmed = true
	m.errorMessage = ""
	return m, tea.Quit
}

func (m WorkspacePromptModel) View() string {
	if !m.ready {
		return "Initializing..."
	}
	if m.confirmed {
		var b strings.Builder
		b.WriteString(PromptLabelStyle.Render(fmt.Sprintf("Current directory: %s", m.currentDir)))
		b.WriteByte('\n')
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("Workspace set to: %s", m.confirmedPath)))
		b.WriteString("\n\n")
		return b.String()
	}

	var b strings.Builder

	b.WriteString(PromptLabelStyle.Render(fmt.Sprintf("Current directory: %s", m.currentDir)))
	b.WriteByte('\n')
	b.WriteString("Press Enter to confirm, or type an absolute path. (Alt+Enter for newline)")
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(m.textarea.View())

	if m.errorMessage != "" {
		b.WriteByte('\n')
		b.WriteString(ErrorStyle.Render(m.errorMessage))
	}

	b.WriteByte('\n')

	return b.String()
}

func (m WorkspacePromptModel) Result() WorkspacePromptResult {
	return WorkspacePromptResult{
		Path:      m.confirmedPath,
		Cancelled: !m.confirmed,
	}
}
