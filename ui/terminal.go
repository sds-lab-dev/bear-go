package ui

import (
	"os"

	"github.com/charmbracelet/x/term"
)

var MinTerminalWidth = bearArtWidth + bearArtGap + 1

type TerminalSize struct {
	Width  int
	Height int
}

func GetTerminalSize() TerminalSize {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		// Fallback to default size if we can't get the terminal size.
		return TerminalSize{Width: MinTerminalWidth, Height: 24}
	}

	return TerminalSize{Width: width, Height: height}
}
