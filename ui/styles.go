package ui

import "github.com/charmbracelet/lipgloss"

var (
	bearArtStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	sloganStyle  = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).Bold(true)
	descriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	separatorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	successStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
)

func renderAgentActivePrompt(msg string, prefixDot bool) string {
	prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1C5A8B"))
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)

	var result string
	if prefixDot {
		result = prefixStyle.Render("● ") + promptStyle.Render(msg)
	} else {
		result = promptStyle.Render(msg)
	}

	return result
}

func renderAgentInactivePrompt(msg string, prefixDot bool) string {
	prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).Bold(true)

	var result string
	if prefixDot {
		result = prefixStyle.Render("● ") + promptStyle.Render(msg)
	} else {
		result = promptStyle.Render(msg)
	}
	return result
}

func renderAgentThinking(msg string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).Italic(true)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8D8D8D"))

	return headerStyle.Render("Thinking...\n") + bodyStyle.Render(msg)
}
