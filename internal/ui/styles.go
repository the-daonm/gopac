package ui

import "github.com/charmbracelet/lipgloss"

var (
	GruvBg     = lipgloss.Color("#282828")
	GruvFg     = lipgloss.Color("#ebdbb2")
	GruvRed    = lipgloss.Color("#cc241d")
	GruvGreen  = lipgloss.Color("#98971a")
	GruvYellow = lipgloss.Color("#d79921")
	GruvOrange = lipgloss.Color("#d65d0e")
	GruvGray   = lipgloss.Color("#7c6f64")
	GruvBlue   = lipgloss.Color("#458588")

	AppStyle = lipgloss.NewStyle().Margin(2, 2, 0, 2)

	ContainerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GruvGray).
			Padding(0)

	ListStyle = lipgloss.NewStyle().Padding(1)
	DescStyle = lipgloss.NewStyle().Padding(1, 2, 0, 2)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(GruvBg).
			Background(GruvGreen).
			Bold(true).
			Padding(0, 2)

	InputStyle = lipgloss.NewStyle().
			Foreground(GruvYellow).
			Padding(0, 2)

	LabelStyle = lipgloss.NewStyle().Foreground(GruvGray).Width(12).Bold(true)
	ValueStyle = lipgloss.NewStyle().Foreground(GruvFg)
	LinkStyle  = lipgloss.NewStyle().Foreground(GruvBlue).Underline(true)
)

func GetRepoColor(isAUR bool) lipgloss.Color {
	if isAUR {
		return GruvOrange
	}
	return GruvGreen
}
