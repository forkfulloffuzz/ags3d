package tui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle       = lipgloss.Style{}
	selectedStyle     = lipgloss.Style{}
	detailHeaderStyle = lipgloss.Style{}
	staleStyle        = lipgloss.Style{}
	orphanStyle       = lipgloss.Style{}
	sourceStyle       = lipgloss.Style{}
	translatedStyle   = lipgloss.Style{}
	untranslatedStyle = lipgloss.Style{}
	editFieldStyle    = lipgloss.Style{}
	statusStyle       = lipgloss.Style{}
	changedStyle      = lipgloss.Style{}
)

func init() {
	headerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true)

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("25")).
		Bold(true)

	detailHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true)

	staleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	orphanStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	sourceStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	translatedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))

	untranslatedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	editFieldStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	changedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))
}
