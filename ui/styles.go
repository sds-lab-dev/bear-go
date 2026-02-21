package ui

import "github.com/charmbracelet/lipgloss"

var (
	BearArtStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	SloganStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	DescriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	SeparatorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	ErrorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	SuccessStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	QuestionStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	InfoStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
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
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Bold(true)

	var result string
	if prefixDot {
		result = prefixStyle.Render("● ") + promptStyle.Render(msg)
	} else {
		result = promptStyle.Render(msg)
	}
	return result
}

func renderAgentThinking(msg string) string {
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Italic(true)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8D8D8D"))

	return headerStyle.Render("Thinking...\n") + bodyStyle.Render(msg)
}
