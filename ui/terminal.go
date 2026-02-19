package ui

import (
	"os"

	"github.com/charmbracelet/x/term"
)

type TerminalSize struct {
	Width  int
	Height int
}

func GetTerminalSize() (TerminalSize, error) {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		return TerminalSize{}, err
	}

	return TerminalSize{Width: width, Height: height}, nil
}
