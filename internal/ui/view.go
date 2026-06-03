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

	if m.showingConfirm {
		return m.confirmView()
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
	availableSearchWidth := max(m.width-fixedContentWidth, 10)

	spin := "  "
	if m.isSearching {
		spin = m.spinner.View() + " "
	}

	// Calculate input width dynamically inside the pill shape
	middleWidth := availableSearchWidth - 2 // 2 for the rounded corners  and 
	inputWidth := middleWidth - 4
	if inputWidth < 5 {
		inputWidth = 5
	}
	m.input.Width = inputWidth
	middleWidth = inputWidth + 4

	// Build the pill-shaped search input box
	leftPill := lipgloss.NewStyle().Foreground(CurrentTheme.Base).Background(CurrentTheme.Highlight).Render("")
	rightPill := lipgloss.NewStyle().Foreground(CurrentTheme.Base).Background(CurrentTheme.Highlight).Render("")
	
	searchContent := lipgloss.NewStyle().
		Background(CurrentTheme.Base).
		Foreground(searchBorderColor).
		Width(middleWidth).
		Render(spin + searchIcon + m.input.View())

	searchView := leftPill + searchContent + rightPill

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

	totalQueued := len(m.markedInstall) + len(m.markedRemove)
	queueText := ""
	if totalQueued > 0 {
		queueText = fmt.Sprintf(" • 📥 QUEUE: %d (+%d, -%d) Press 'I' to Apply, 'C' to Clear", totalQueued, len(m.markedInstall), len(m.markedRemove))
	}

	// Dynamic Status Bar
	var helpText string
	if m.searching {
		helpText = "   SEARCHING • Enter: Search • Tab: Focus List • Esc: Cancel " + queueText
	} else if m.focusSide == 0 {
		helpText = "   LIST VIEW • h/l: Change Tab • Enter: Install/Remove • Space: Queue • Tab: Focus Details • /: Search • ?: Help " + queueText
	} else {
		helpText = "   DETAILS • j/k: Scroll • Esc: Back to List • Tab: Focus Search • ?: Help " + queueText
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
		{"Space", "Queue/unqueue package"},
		{"I", "Apply queued changes (Confirm)"},
		{"C", "Clear queue"},
		{"Enter", "Install/Remove (Confirm)"},
		{"h/l or ◄/►", "Change tab filter"},
		{"j/k or Up/Down", "Scroll details (in Details view)"},
		{"Esc", "Return to list view (in Details view)"},
		{"p", "View PKGBUILD (AUR only)"},
		{"Up/Down", "Search history (when searching)"},
		{"Mouse", "Click to focus panels or tabs"},
		{"?", "Toggle help"},
		{"q or Ctrl+C", "Quit"},
	}

	var sb strings.Builder
	sb.WriteByte('\n')
	sb.WriteString(title)
	sb.WriteString("\n\n")

	for _, r := range rows {
		key := lipgloss.NewStyle().Foreground(CurrentTheme.Focus).Bold(true).Width(15).Render(r.Key)
		desc := lipgloss.NewStyle().Foreground(CurrentTheme.Text).Render(r.Desc)
		fmt.Fprintf(&sb, "%s %s\n", key, desc)
	}

	sb.WriteByte('\n')
	sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Render("Press '?' to close help"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(CurrentTheme.Focus).
			Padding(1, 4).
			Render(sb.String()))
}

func (m Model) confirmView() string {
	var sections []string

	// Header
	header := lipgloss.NewStyle().
		Foreground(CurrentTheme.Base).
		Background(CurrentTheme.Yellow).
		Bold(true).
		Padding(0, 3).
		Render(" ACTION REQUIRED ")

	sections = append(sections, header, "")

	// Subtitle
	sections = append(sections, lipgloss.NewStyle().Foreground(CurrentTheme.Text).Bold(true).Render("The following system actions will be performed:"))
	sections = append(sections, "")

	// Format categories
	bulletColorOfficial := CurrentTheme.RepoOfficial
	bulletColorAUR := CurrentTheme.RepoAUR
	bulletColorRemove := CurrentTheme.Red

	renderPkgList := func(pkgs []string, bulletColor lipgloss.Color) string {
		if len(pkgs) == 0 {
			return ""
		}
		bullet := lipgloss.NewStyle().Foreground(bulletColor).Render("•")
		if len(pkgs) <= 5 {
			var lines []string
			for _, p := range pkgs {
				lines = append(lines, fmt.Sprintf("  %s %s", bullet, lipgloss.NewStyle().Foreground(CurrentTheme.Text).Render(p)))
			}
			return strings.Join(lines, "\n")
		} else {
			var lines []string
			for i := 0; i < 5; i++ {
				lines = append(lines, fmt.Sprintf("  %s %s", bullet, lipgloss.NewStyle().Foreground(CurrentTheme.Text).Render(pkgs[i])))
			}
			remaining := len(pkgs) - 5
			lines = append(lines, fmt.Sprintf("  %s %s", bullet, lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Italic(true).Render(fmt.Sprintf("... and %d more", remaining))))
			return strings.Join(lines, "\n")
		}
	}

	if len(m.confirmInstallOfficial) > 0 {
		label := lipgloss.NewStyle().Foreground(CurrentTheme.RepoOfficial).Bold(true).Render("📥 Install (Official):")
		list := renderPkgList(m.confirmInstallOfficial, bulletColorOfficial)
		sections = append(sections, label, list, "")
	}

	if len(m.confirmInstallAUR) > 0 {
		label := lipgloss.NewStyle().Foreground(CurrentTheme.RepoAUR).Bold(true).Render("📥 Install (AUR):")
		list := renderPkgList(m.confirmInstallAUR, bulletColorAUR)
		sections = append(sections, label, list, "")
	}

	if len(m.confirmRemove) > 0 {
		label := lipgloss.NewStyle().Foreground(CurrentTheme.Red).Bold(true).Render("🗑️ Remove:")
		list := renderPkgList(m.confirmRemove, bulletColorRemove)
		sections = append(sections, label, list, "")
	}

	// Divider
	sections = append(sections, lipgloss.NewStyle().Foreground(CurrentTheme.Highlight).Render(strings.Repeat("─", 50)))

	// Warning
	if len(m.confirmInstallAUR) > 0 {
		warningText := "⚠️  AUR warning: AUR packages are user-submitted. Make sure you trust\n   their PKGBUILDs before executing the installation."
		warningStyled := lipgloss.NewStyle().
			Foreground(CurrentTheme.Orange).
			Render(warningText)
		sections = append(sections, warningStyled, "")
	}

	// Prompt
	prompt := lipgloss.NewStyle().Foreground(CurrentTheme.Focus).Bold(true).Render("Proceed with execution? [Y/n]")
	sections = append(sections, prompt, "")

	// Keybind guide
	keybinds := lipgloss.NewStyle().
		Foreground(CurrentTheme.Gray).
		Render("y/Enter: Confirm  •  n/Esc: Cancel  •  q: Quit")
	sections = append(sections, keybinds)

	body := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(CurrentTheme.Yellow).
			Padding(1, 4).
			Render(body))
}

