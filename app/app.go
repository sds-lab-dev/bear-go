package app

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sds-lab-dev/bear-go/ui"
)

type mainState int

const (
	stateWorkspace mainState = iota
	stateUserRequest
	stateSpecDrafting
	stateDone
)

type mainModel struct {
	state         mainState
	currentModel  tea.Model
	mainHeaderCmd tea.Cmd
	workspacePath string
}

func newMainModel() (mainModel, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return mainModel{}, fmt.Errorf("failed to get current working directory: %w", err)
	}

	mainHeaderCmd, err := buildMainHeader()
	if err != nil {
		return mainModel{}, fmt.Errorf("failed to build main header: %w", err)
	}

	return mainModel{
		state: stateWorkspace,
		currentModel: ui.NewWorkspacePromptModel(
			cwd,
			ValidateWorkspacePath,
		),
		mainHeaderCmd: mainHeaderCmd,
	}, nil
}

func buildMainHeader() (tea.Cmd, error) {
	var b strings.Builder

	apiKey := getAPIKeyFromEnvVar(nil)
	if apiKey == "" {
		b.WriteString(ui.ErrorStyle.Render("ANTHROPIC_API_KEY environment variable is not set or empty; trying to use a subscription plan, but this may fail if the key is required for authentication"))
		b.WriteByte('\n')
	}

	terminalSize, err := ui.GetTerminalSize()
	if err != nil {
		return nil, fmt.Errorf("failed to query terminal size: %w", err)
	}
	if terminalSize.Width < ui.MinTerminalWidth {
		return nil, fmt.Errorf("terminal too narrow (%d columns); at least %d columns required", terminalSize.Width, ui.MinTerminalWidth)
	}
	b.WriteString(ui.RenderBanner(terminalSize.Width))

	return tea.Println(b.String()), nil
}

func (m mainModel) Init() tea.Cmd {
	return tea.Sequence(m.mainHeaderCmd, m.currentModel.Init())
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Bubbletea message handling
	switch msg := msg.(type) {
	// Global key handling (e.g., Ctrl+C to quit)
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	// Terminal resize handling
	case tea.WindowSizeMsg:

	}

	// Internal message handling
	switch msg := msg.(type) {
	case ui.WorkspacePromptResult:
		m.workspacePath = msg.Path
		m.state = stateUserRequest
		m.currentModel = ui.NewUserRequestPromptModel()
		return m, tea.Sequence(tea.Printf("%v\n", msg.View), m.currentModel.Init())
	case ui.UserRequestPromptResult:
		// TODO: next state and model
		return m, tea.Sequence(tea.Printf("%v\n", msg.View), tea.Quit)
	}

	// Delegate to the current sub-model
	updatedModel, cmd := m.currentModel.Update(msg)
	m.currentModel = updatedModel

	return m, cmd
}

func (m mainModel) View() string {
	return m.currentModel.View()
}

func Run() error {
	mainModel, err := newMainModel()
	if err != nil {
		return fmt.Errorf("failed to initialize main model: %w", err)
	}

	mainProgram := tea.NewProgram(mainModel)
	_, err = mainProgram.Run()
	if err != nil {
		return fmt.Errorf("application main loop failed: %w", err)
	}

	// workspaceModel := ui.NewWorkspacePromptModel(cwd, ValidateWorkspacePath)
	// workspaceProgram := tea.NewProgram(workspaceModel, tea.WithOutput(stderr))
	// finalWorkspaceModel, err := workspaceProgram.Run()
	// if err != nil {
	// 	return fmt.Errorf("workspace prompt failed: %w", err)
	// }

	// workspaceResult := finalWorkspaceModel.(ui.WorkspacePromptModel).Result()
	// if workspaceResult.Cancelled {
	// 	return nil
	// }

	// requestModel := ui.NewUserRequestPromptModel()
	// requestProgram := tea.NewProgram(requestModel, tea.WithOutput(stderr))
	// finalRequestModel, err := requestProgram.Run()
	// if err != nil {
	// 	return fmt.Errorf("user request prompt failed: %w", err)
	// }

	// requestResult := finalRequestModel.(ui.UserRequestPromptModel).Result()
	// if requestResult.Cancelled {
	// 	return nil
	// }

	// client, err := claudecode.NewClient(apiKey, workspaceResult.Path)
	// if err != nil {
	// 	return fmt.Errorf("failed to create claude client: %w", err)
	// }

	// userPrompt := BuildSpecClarificationUserPrompt(requestResult.Text, "")

	// queryFunc := func(callback claudecode.StreamCallback) (claudecode.ResultData, error) {
	// 	return client.Query(SpecAgentSystemPrompt, userPrompt, SpecClarificationJSONSchema, callback)
	// }

	// specStreamModel := ui.NewSpecStreamModel(queryFunc)
	// specProgram := tea.NewProgram(specStreamModel, tea.WithOutput(stderr))
	// finalSpecModel, err := specProgram.Run()
	// if err != nil {
	// 	return fmt.Errorf("spec stream failed: %w", err)
	// }

	// specResult := finalSpecModel.(ui.SpecStreamModel).Result()
	// if specResult.Cancelled {
	// 	return nil
	// }
	// if specResult.Err != nil {
	// 	return fmt.Errorf("spec agent failed: %w", specResult.Err)
	// }

	return nil
}
