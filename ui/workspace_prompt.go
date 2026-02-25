package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sds-lab-dev/bear-go/log"
)

type WorkspacePromptResult struct {
	Path string
}

type WorkspacePromptModel struct {
	textarea      textarea.Model
	currentDir    string
	validatePath  func(string) error
	errorMessage  string
	confirmedPath string
}

func NewWorkspacePromptModel(currentDir string, validatePath func(string) error) WorkspacePromptModel {
	terminalSize := GetTerminalSize()

	ta := textarea.New()
	ta.Placeholder = currentDir
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(terminalSize.Width)
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

// Update handles incoming messages for the workspace prompt.
//
// It processes key events for confirming the workspace path, updates the
// textarea state, and listens for window size changes to set the initial width
// of the textarea.
//
// The function should return the updated model and any commands to execute
// (e.g., tea.Quit if the user presses Ctrl+C).
func (m WorkspacePromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("received update message in WorkspacePromptModel: %#v", msg))

	// WindowSizeMsg is sent by the Bubble Tea runtime when the terminal size
	// changes, including when the program first starts. We use it to set the
	// width of our textarea and mark ourselves as ready to render.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Debug(fmt.Sprintf("received window size message: width=%d, height=%d", msg.Width, msg.Height))
		m.textarea.SetWidth(msg.Width)
		return m, nil
	}

	// If the message is a reserved key event, handle it with our custom key
	// handling logic.
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		// Go to next step on Enter
		case "enter":
			log.Debug("Enter key pressed, handling workspace confirmation")
			return m.handleEnter()
		}
		// Fall through.
	}

	var cmd tea.Cmd
	// For all other messages, we pass them to the textarea component to handle
	// them in the text area.
	log.Debug(fmt.Sprintf("passing message of type %T to textarea component", msg))
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m WorkspacePromptModel) handleEnter() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.textarea.Value())

	// If the user just presses Enter without typing anything, we treat it as
	// confirming the current directory. Otherwise, we validate the entered path
	// and set it as the confirmed workspace if it's valid.
	if value == "" {
		m.confirmedPath = m.currentDir
	} else {
		if err := m.validatePath(value); err != nil {
			m.errorMessage = err.Error()
			return m, nil
		}
		m.confirmedPath = value
	}

	// We're done here.
	m.errorMessage = ""
	var b strings.Builder
	b.WriteString(renderAgentInactivePrompt(successStyle.Render(fmt.Sprintf("Workspace set to: %s", m.confirmedPath)), true))
	log.Info(fmt.Sprintf("Workspace confirmed: %s", m.confirmedPath))
	cmd := tea.Sequence(
		tea.Printf("%v\n", b.String()),
		func() tea.Msg {
			return WorkspacePromptResult{
				Path: m.confirmedPath,
			}
		},
	)

	return m, cmd
}

func (m WorkspacePromptModel) View() string {
	var b strings.Builder

	log.Debug(fmt.Sprintf("rendering workspace prompt view: errorMessage=%v", m.errorMessage))

	b.WriteString(renderAgentActivePrompt(fmt.Sprintf("Current directory: %s", m.currentDir), true))
	b.WriteByte('\n')
	b.WriteString("Press Enter to confirm, or type an absolute path.")
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(m.textarea.View())
	if m.errorMessage != "" {
		b.WriteByte('\n')
		b.WriteString(errorStyle.Render(m.errorMessage))
	}
	b.WriteByte('\n')

	return b.String()
}
