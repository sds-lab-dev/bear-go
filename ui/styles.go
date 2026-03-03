package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

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

	const maxVisualLines = 5
	terminalWidth := GetTerminalSize().Width
	truncatedMsg := truncateToVisualLines(msg, maxVisualLines, terminalWidth)

	return prefixStyle.Render("● ") +
		headerStyle.Render("Tool Call Result:") +
		"\n" +
		bodyStyle.Render(truncatedMsg)
}

// truncateToVisualLines truncates text to fit within maxLines visual lines
// based on the given terminal width. A visual line is counted by considering
// both explicit newline characters and line wrapping at the terminal width.
// If the text exceeds the limit, it is truncated and a "... (truncated)"
// indicator is appended within the line budget.
func truncateToVisualLines(
	text string, maxLines int, terminalWidth int,
) string {
	if maxLines < 2 {
		maxLines = 2
	}
	if terminalWidth <= 0 {
		terminalWidth = 80
	}

	if countTotalVisualLines(text, terminalWidth) <= maxLines {
		return text
	}

	// Reserve 1 line for the truncation indicator.
	contentMaxLines := maxLines - 1
	logicalLines := strings.Split(text, "\n")
	var result []string
	visualLinesUsed := 0

	for _, line := range logicalLines {
		lineWidth := runewidth.StringWidth(line)
		visualLines := visualLineCount(lineWidth, terminalWidth)
		remaining := contentMaxLines - visualLinesUsed

		if remaining <= 0 {
			break
		}

		if visualLines <= remaining {
			result = append(result, line)
			visualLinesUsed += visualLines
			continue
		}

		maxWidth := remaining * terminalWidth
		result = append(result, truncateToVisualWidth(line, maxWidth))
		break
	}

	result = append(result, "... (truncated)")
	return strings.Join(result, "\n")
}

func countTotalVisualLines(text string, terminalWidth int) int {
	total := 0
	for _, line := range strings.Split(text, "\n") {
		lineWidth := runewidth.StringWidth(line)
		total += visualLineCount(lineWidth, terminalWidth)
	}
	return total
}

func visualLineCount(lineWidth int, terminalWidth int) int {
	if lineWidth == 0 {
		return 1
	}
	return (lineWidth + terminalWidth - 1) / terminalWidth
}

func truncateToVisualWidth(line string, maxWidth int) string {
	width := 0
	for i, r := range line {
		w := runewidth.RuneWidth(r)
		if width+w > maxWidth {
			return line[:i]
		}
		width += w
	}
	return line
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
