package ui

import (
	"strings"

	"github.com/muesli/reflow/wordwrap"
)

func wrapWords(text string, maxWidth int) []string {
	if text == "" || maxWidth <= 0 {
		return nil
	}

	wrapped := wordwrap.String(text, maxWidth)
	return strings.Split(wrapped, "\n")
}
