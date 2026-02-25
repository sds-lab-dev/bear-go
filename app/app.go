package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sds-lab-dev/bear-go/ai"
	"github.com/sds-lab-dev/bear-go/log"
	"github.com/sds-lab-dev/bear-go/ui"
)

type mainModelState int

const (
	mainStateWorkspaceDir mainModelState = iota
	mainStateUserRequest
	mainStateSpecDrafting
	mainStateDone
	mainStateSwitching
)

type stateSwitchMsg struct {
	newState mainModelState
	newModel tea.Model
}

type mainModel struct {
	sessionID     string
	state         mainModelState
	currentModel  tea.Model
	mainHeaderCmd tea.Cmd
	workspacePath string
	aiPorts       ai.Ports
	err           error
}

func newMainModel(sessionID string, aiPorts ai.Ports) (mainModel, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return mainModel{}, fmt.Errorf("failed to get current working directory: %w", err)
	}

	mainHeaderCmd, err := buildMainHeader()
	if err != nil {
		return mainModel{}, fmt.Errorf("failed to build main header: %w", err)
	}

	return mainModel{
		sessionID: sessionID,
		state:     mainStateWorkspaceDir,
		currentModel: ui.NewWorkspacePromptModel(
			cwd,
			ValidateWorkspacePath,
		),
		mainHeaderCmd: mainHeaderCmd,
		aiPorts:       aiPorts,
		err:           nil,
	}, nil
}

func buildMainHeader() (tea.Cmd, error) {
	terminalSize := ui.GetTerminalSize()
	if terminalSize.Width < ui.MinTerminalWidth {
		return nil, fmt.Errorf("terminal too narrow (%d columns); at least %d columns required", terminalSize.Width, ui.MinTerminalWidth)
	}

	return tea.Println(ui.RenderBanner(terminalSize.Width)), nil
}

func (m mainModel) Init() tea.Cmd {
	if m.currentModel == nil {
		panic("currentModel must be initialized before calling Init()")
	}

	return tea.Sequence(m.mainHeaderCmd, m.currentModel.Init())
}

func (m mainModel) switchModel(newState mainModelState, newModel tea.Model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("switching main model state from %v to %v", m.state, newState))

	m.state = mainStateSwitching
	cmd = tea.Sequence(
		func() tea.Msg {
			return stateSwitchMsg{
				newState: newState,
				newModel: newModel,
			}
		},
		cmd,
	)
	return m, cmd
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("main model received update message of type %#v in state %v", msg, m.state))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key handling (e.g., Ctrl+C to quit)
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case stateSwitchMsg:
		// This internal message is used to switch between sub-models to ensure
		// that the mainModel clears the terminal and re-renders the new sub-model
		// cleanly when transitioning between states.
		//
		// Sub-models DO NOT need to worry about clearing the terminal when they
		// finish, they can just send a message to the mainModel to inform the
		// final result of the sub-model.
		m.state = msg.newState
		m.currentModel = msg.newModel
		return m, m.currentModel.Init()
	case ui.WorkspacePromptResult:
		m.workspacePath = msg.Path
		return m.switchModel(mainStateUserRequest, ui.NewUserRequestPromptModel(), nil)
	case ui.UserRequestPromptResult:
		session, err := m.aiPorts.NewSession(m.workspacePath)
		if err != nil {
			m.err = fmt.Errorf("failed to create AI session: %w", err)
			return m, tea.Quit
		}
		return m.switchModel(mainStateSpecDrafting, ui.NewSpecPromptModel(msg.Text, session), nil)
	case ui.SpecPromptResult:
		if msg.Err != nil {
			m.err = fmt.Errorf("spec prompt failed: %w", msg.Err)
			return m, tea.Quit
		}
		return m.switchModel(mainStateDone, nil, tea.Quit)
	}

	// For all other messages, delegate them to the current sub-model.
	//
	// IMPORTANT:
	// You MUST NOT return the sub-model directly here, as that would lose the
	// mainModel and break the main loop. Instead, you MUST update the
	// currentModel field of the mainModel with the updated sub-model returned
	// from the Update call, and then return the mainModel itself.
	log.Debug(fmt.Sprintf("delegating message to current sub-model of type %T: %#v", m.currentModel, msg))
	updatedModel, cmd := m.currentModel.Update(msg)
	m.currentModel = updatedModel

	return m, cmd
}

func (m mainModel) View() string {
	log.Debug(fmt.Sprintf("main model rendering view in state %v with current sub-model of type %T", m.state, m.currentModel))

	// m.currentModel is allowed to be nil here that means we don't have a
	// sub-model to render anymore (e.g., we've finished all the steps).
	if m.currentModel == nil {
		return ""
	}

	return m.currentModel.View()
}

func appMain(sessionID string, aiPorts ai.Ports) error {
	model, err := newMainModel(sessionID, aiPorts)
	if err != nil {
		return fmt.Errorf("failed to initialize main model: %v", err)
	}

	mainProgram := tea.NewProgram(model)
	finalModel, err := mainProgram.Run()
	if err != nil {
		return fmt.Errorf("application main loop failed: %v", err)
	}

	if model, ok := finalModel.(mainModel); ok && model.err != nil {
		return fmt.Errorf("failed to run main model: %v", model.err)
	}

	return nil
}

func Run(buildVersion string, sessionID string, aiPorts ai.Ports) {
	log.InitLogger(sessionID)
	defer log.CloseLogger()

	fmt.Printf("Log file initialized at %s\n", log.GetLogPath())
	log.Info(fmt.Sprintf("Starting application: sessionID=%v, buildVersion=%v", sessionID, buildVersion))

	if err := appMain(sessionID, aiPorts); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		log.Fatal(err.Error())
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
	//
	// return nil
}
