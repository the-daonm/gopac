package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	// Layout
	innerWidth := m.width - 6
	if innerWidth < 10 {
		innerWidth = 10
	}

	// Header & Search
	var searchBorderColor lipgloss.Color
	if m.searching {
		searchBorderColor = GruvYellow
	} else {
		searchBorderColor = GruvGray
	}
	searchIcon := " "
	if m.searching {
		searchIcon = " "
	}

	headerLeft := HeaderStyle.Render(" GOPAC ")

	searchWidth := innerWidth - lipgloss.Width(headerLeft)
	if searchWidth < 10 {
		searchWidth = 10
	}

	headerRight := InputStyle.
		Width(searchWidth).
		Background(lipgloss.Color("#32302f")).
		Foreground(searchBorderColor).
		Render(searchIcon + m.input.View())

	header := lipgloss.JoinHorizontal(lipgloss.Top, headerLeft, headerRight)
	header = lipgloss.NewStyle().
		Width(innerWidth).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(GruvGray).
		Render(header)

	// Tab Bar
	var tabViews []string
	for i, t := range tabs {
		style := lipgloss.NewStyle().Foreground(GruvGray).Padding(0, 1)
		if i == m.activeTab {
			style = lipgloss.NewStyle().Foreground(GruvBg).Background(GruvYellow).Bold(true).Padding(0, 1)
		}
		tabViews = append(tabViews, style.Render(t))
	}
	tabBar := lipgloss.NewStyle().
		Width(innerWidth).
		Background(GruvBg).
		MarginBottom(1). // <--- THIS ADDS THE BLANK SPACE
		Render(lipgloss.JoinHorizontal(lipgloss.Top, tabViews...))

	// Content
	var listContent string
	if len(m.list.Items()) == 0 {
		msg := lipgloss.NewStyle().Foreground(GruvRed).Bold(true).Render("No Packages Found")
		listContent = lipgloss.Place(m.listWidth, m.panelHeight, lipgloss.Center, lipgloss.Center, msg)
	} else {
		listContent = m.list.View()
	}

	listView := ListStyle.
		Width(m.listWidth).
		Height(m.panelHeight).
		Render(listContent)

	descView := DescStyle.
		Width(m.descWidth).
		Height(m.panelHeight).
		Render(m.viewport.View())

	separator := lipgloss.NewStyle().
		Height(m.panelHeight).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(GruvGray).
		MarginLeft(0).
		Render("")

	content := lipgloss.JoinHorizontal(lipgloss.Top, listView, separator, descView)

	// Wrap
	innerView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabBar,
		content,
	)

	return AppStyle.Render(ContainerStyle.Render(innerView))
}
