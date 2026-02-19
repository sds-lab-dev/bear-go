package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type UserRequestPromptResult struct {
	Text string
	View string
}

type editorFinishedMsg struct {
	err error
}

type startEditorMsg struct{}

type UserRequestPromptModel struct {
	textarea        textarea.Model
	errorMessage    string
	confirmed       bool
	confirmedText   string
	resolveEditor   func() (EditorCommand, error)
	tempFilePath    string
	launchingEditor bool
}

func NewUserRequestPromptModel() UserRequestPromptModel {
	terminalSize, err := GetTerminalSize()
	if err != nil {
		// If we fail to get the terminal size, we can still proceed; we'll just set a default width.
		terminalSize = TerminalSize{Width: 80, Height: 24}
	}

	ta := textarea.New()
	ta.Placeholder = ""
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(terminalSize.Width)
	ta.SetHeight(terminalSize.Height / 2)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	ta.Focus()

	return UserRequestPromptModel{
		textarea: ta,
		resolveEditor: func() (EditorCommand, error) {
			return resolveEditor(os.LookupEnv, commandExistsOnSystem)
		},
	}
}

func commandExistsOnSystem(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func (m UserRequestPromptModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m UserRequestPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height / 2)
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case startEditorMsg:
		return m.handleEditorLaunch()
	case editorFinishedMsg:
		return m.handleEditorFinished(msg)
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m UserRequestPromptModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleEnter()
	case "shift+enter", "alt+enter":
		m.textarea.InsertString("\n")
		m.errorMessage = ""
		return m, nil
	case "ctrl+g":
		return m.prepareEditorLaunch()
	}

	m.errorMessage = ""
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m UserRequestPromptModel) handleEnter() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.textarea.Value())
	if value == "" {
		m.errorMessage = "Please enter your request."
		return m, nil
	}

	// We're done here.
	m.confirmed = true
	m.confirmedText = value
	m.errorMessage = ""
	var b strings.Builder
	b.WriteString(renderAgentInactivePrompt("You requested as follows:"))
	b.WriteByte('\n')
	b.WriteString(m.textarea.Value())
	return m, func() tea.Msg {
		return UserRequestPromptResult{
			Text: m.confirmedText,
			View: b.String(),
		}
	}
}

func (m UserRequestPromptModel) prepareEditorLaunch() (tea.Model, tea.Cmd) {
	m.launchingEditor = true
	return m, func() tea.Msg { return startEditorMsg{} }
}

func (m UserRequestPromptModel) handleEditorLaunch() (tea.Model, tea.Cmd) {
	editorCmd, err := m.resolveEditor()
	if err != nil {
		m.launchingEditor = false
		m.errorMessage = err.Error()
		return m, nil
	}

	tmpFile, err := os.CreateTemp("", "bear-request-*.md")
	if err != nil {
		m.launchingEditor = false
		m.errorMessage = fmt.Sprintf("Failed to create temp file: %v", err)
		return m, nil
	}

	content := m.textarea.Value()
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		m.launchingEditor = false
		m.errorMessage = fmt.Sprintf("Failed to write temp file: %v", err)
		return m, nil
	}
	tmpFile.Close()

	m.tempFilePath = tmpFile.Name()

	args := append(editorCmd.Args, m.tempFilePath)
	c := exec.Command(editorCmd.Executable, args...)

	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m UserRequestPromptModel) handleEditorFinished(msg editorFinishedMsg) (tea.Model, tea.Cmd) {
	m.launchingEditor = false

	if msg.err != nil {
		m.errorMessage = fmt.Sprintf("Editor failed: %v", msg.err)
		m.cleanupTempFile()
		return m, nil
	}

	content, err := os.ReadFile(m.tempFilePath)
	if err != nil {
		m.errorMessage = fmt.Sprintf("Failed to read editor output: %v", err)
		m.cleanupTempFile()
		return m, nil
	}

	m.textarea.SetValue(string(content))
	m.errorMessage = ""
	m.cleanupTempFile()
	return m, nil
}

func (m *UserRequestPromptModel) cleanupTempFile() {
	if m.tempFilePath != "" {
		os.Remove(m.tempFilePath)
		m.tempFilePath = ""
	}
}

func (m UserRequestPromptModel) View() string {
	if m.confirmed || m.launchingEditor {
		return ""
	}

	var b strings.Builder

	b.WriteString(renderAgentActivePrompt("Enter your request:"))
	b.WriteByte('\n')
	b.WriteString("Press Enter to confirm, Shift+Enter or Alt+Enter for newline, Ctrl+G for external editor.")
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
