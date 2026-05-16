package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width < 20 || m.height < 10 {
		return "Terminal too small"
	}

	if m.showingHelp {
		return m.helpView()
	}

	// Header
	logo := HeaderStyle.Render(" GOPAC ")

	// Render Tabs
	var tabViews []string
	for i, t := range tabs {
		style := lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Padding(0, 1)
		if i == m.activeTab {
			style = lipgloss.NewStyle().
				Foreground(CurrentTheme.Base).
				Background(CurrentTheme.Focus).
				Bold(true).
				Padding(0, 1)
		}
		tabViews = append(tabViews, style.Render(t))
	}
	tabsView := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)

	// Search Styling
	var searchBorderColor lipgloss.Color
	if m.searching {
		searchBorderColor = CurrentTheme.Focus
	} else {
		searchBorderColor = CurrentTheme.Gray
	}
	searchIcon := " "

	gap := lipgloss.NewStyle().Background(CurrentTheme.Highlight).Render("   ")
	gapWidth := lipgloss.Width(gap)

	fixedContentWidth := lipgloss.Width(logo) + lipgloss.Width(tabsView) + (gapWidth * 2)
	availableSearchWidth := m.width - fixedContentWidth

	if availableSearchWidth < 5 {
		availableSearchWidth = 5
	}

	spin := ""
	if m.isSearching {
		spin = m.spinner.View() + " "
	}

	searchView := InputStyle.
		Width(availableSearchWidth).
		Background(CurrentTheme.Highlight).
		Foreground(searchBorderColor).
		Render(spin + searchIcon + m.input.View())

	// Join Header Elements
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		logo,
		gap,
		searchView,
		gap,
		tabsView,
	)

	header = lipgloss.NewStyle().
		Width(m.width).
		Background(CurrentTheme.Highlight).
		Render(header)

	// Dynamic Status Bar
	var helpText string
	if m.searching {
		helpText = "   SEARCHING • Enter: Confirm • Tab: Focus List • Esc: Cancel "
	} else if m.focusSide == 0 {
		helpText = "   LIST VIEW • ◄/►: Change Filter • Enter: Install • U: Update System • /: Search • ?: Help "
	} else {
		helpText = "   DETAILS • Tab: Focus Search • Esc: Back to List • ?: Help "
	}

	statusBar := lipgloss.NewStyle().
		Width(m.width).
		Foreground(CurrentTheme.Gray).
		Background(CurrentTheme.Base).
		Render(helpText)

	// Content Layout
	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(statusBar)

	listStyle := BlurredStyle
	descStyle := BlurredStyle

	if m.focusSide == 0 {
		listStyle = FocusedStyle
	} else {
		descStyle = FocusedStyle
	}

	const borderThickness = 2

	listViewWidth := m.listWidth - borderThickness
	descViewWidth := m.descWidth - borderThickness
	listViewHeight := contentHeight - borderThickness

	// Safety checks
	if listViewWidth < 0 {
		listViewWidth = 0
	}
	if descViewWidth < 0 {
		descViewWidth = 0
	}
	if listViewHeight < 0 {
		listViewHeight = 0
	}

	var listContent string
	if len(m.list.Items()) == 0 {
		msg := lipgloss.NewStyle().Foreground(CurrentTheme.Red).Bold(true).Render("No Packages Found")
		if m.isSearching {
			msg = lipgloss.NewStyle().Foreground(CurrentTheme.Focus).Bold(true).Render("Searching...")
		}
		listContent = lipgloss.Place(listViewWidth-4, listViewHeight, lipgloss.Center, lipgloss.Center, msg)
	} else {
		listContent = m.list.View()
	}

	listView := listStyle.
		Width(listViewWidth).
		Height(listViewHeight).
		Render(listContent)

	descView := descStyle.
		Width(descViewWidth).
		Height(listViewHeight).
		Render(m.viewport.View())

	content := lipgloss.JoinHorizontal(lipgloss.Top, listView, descView)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		statusBar,
	)
}

func (m Model) helpView() string {
	title := HeaderStyle.Render(" GOPAC HELP ")
	
	rows := []struct {
		Key  string
		Desc string
	}{
		{"/", "Search packages"},
		{"U", "Update system packages"},
		{"Tab", "Cycle focus (Search/List/Details)"},
		{"Enter", "Install/Remove package"},
		{"h/l or ◄/►", "Change tab filter"},
		{"p", "View PKGBUILD (AUR only)"},
		{"Up/Down", "Search history (when searching)"},
		{"Mouse", "Click to focus panels or tabs"},
		{"?", "Toggle help"},
		{"q or Ctrl+C", "Quit"},
	}

	var sb strings.Builder
	sb.WriteString("\n" + title + "\n\n")
	
	for _, r := range rows {
		key := lipgloss.NewStyle().Foreground(CurrentTheme.Focus).Bold(true).Width(15).Render(r.Key)
		desc := lipgloss.NewStyle().Foreground(CurrentTheme.Text).Render(r.Desc)
		sb.WriteString(fmt.Sprintf("%s %s\n", key, desc))
	}

	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Render("Press '?' to close help"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, 
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(CurrentTheme.Focus).
			Padding(1, 4).
			Render(sb.String()))
}
