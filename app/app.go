package app

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/sds-lab-dev/bear-go/ui"
)

func Run(stdout, stderr io.Writer) error {
	terminalWidth, err := queryTerminalWidth(stdout)
	if err != nil {
		return fmt.Errorf("failed to query terminal width: %w", err)
	}

	if terminalWidth < ui.MinTerminalWidth {
		return fmt.Errorf("terminal too narrow (%d columns); at least %d columns required", terminalWidth, ui.MinTerminalWidth)
	}

	banner := ui.RenderBanner(terminalWidth)
	fmt.Fprint(stdout, banner)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	model := ui.NewWorkspacePromptModel(cwd, ValidateWorkspacePath)
	program := tea.NewProgram(model, tea.WithOutput(stderr))
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("workspace prompt failed: %w", err)
	}

	result := finalModel.(ui.WorkspacePromptModel).Result()
	if result.Cancelled {
		return nil
	}

	requestModel := ui.NewUserRequestPromptModel()
	requestProgram := tea.NewProgram(requestModel, tea.WithOutput(stderr))
	finalRequestModel, err := requestProgram.Run()
	if err != nil {
		return fmt.Errorf("user request prompt failed: %w", err)
	}

	requestResult := finalRequestModel.(ui.UserRequestPromptModel).Result()
	if requestResult.Cancelled {
		return nil
	}

	return nil
}

func queryTerminalWidth(w io.Writer) (int, error) {
	f, ok := w.(*os.File)
	if !ok {
		return 0, fmt.Errorf("stdout is not a file descriptor")
	}

	width, _, err := term.GetSize(int(f.Fd()))
	if err != nil {
		return 0, err
	}

	return width, nil
}
