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

func renderStreamMessageThinking(msg string) string {
	prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).Italic(true)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8D8D8D"))

	return prefixStyle.Render("● ") +
		headerStyle.Render("Thinking:") +
		"\n" +
		bodyStyle.Render(msg)
}

func renderStreamMessageToolCall(msg string) string {
	prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).Italic(true)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8D8D8D"))

	return prefixStyle.Render("● ") +
		headerStyle.Render("Tool Call:") +
		"\n" +
		bodyStyle.Render(msg)
}

func renderStreamMessageToolCallResult(msg string) string {
	prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).Italic(true)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8D8D8D"))

	return prefixStyle.Render("● ") +
		headerStyle.Render("Tool Call Result:") +
		"\n" +
		bodyStyle.Render(msg)
}

func renderStreamMessageText(msg string) string {
	prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).Italic(true)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8D8D8D"))

	return prefixStyle.Render("● ") +
		headerStyle.Render("Text:") +
		"\n" +
		bodyStyle.Render(msg)
}
