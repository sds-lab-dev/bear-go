package ui

import "github.com/charmbracelet/lipgloss"

var (
	BearArtStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	SloganStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	DescriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	SeparatorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	ErrorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	PromptLabelStyle = lipgloss.NewStyle().Bold(true)
	SuccessStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	AgentActivityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	QuestionStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	InfoStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)
