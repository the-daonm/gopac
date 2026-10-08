package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

	header := m.headerView()

	statusBar := m.statusBarView()

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
		listContent = lipgloss.Place(max(listViewWidth-4, 0), listViewHeight, lipgloss.Center, lipgloss.Center, m.emptyListView())
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
	sections := []struct {
		title string
		rows  [][2]string
	}{
		{"Navigation", [][2]string{
			{"/", "Search packages"},
			{"Tab / S-Tab", "Cycle focus: search, list, details"},
			{"h/l  ←/→", "Switch tab"},
			{"j/k  ↑/↓", "Move in list or scroll details"},
			{"Ctrl+d/u", "Page details down/up"},
			{"↑/↓", "Search history (while searching)"},
			{"Mouse", "Click tabs, panels and packages"},
		}},
		{"Packages", [][2]string{
			{"Enter", "Install / remove selected package"},
			{"Space", "Queue / unqueue selected package"},
			{"a", "Queue / unqueue all visible packages"},
			{"I", "Apply queued changes"},
			{"C", "Clear the queue"},
			{"p", "Toggle PKGBUILD (AUR only)"},
		}},
		{"System", [][2]string{
			{"U", "Upgrade the whole system"},
			{"r", "Refresh package data"},
			{"?", "Toggle this help"},
			{"q  Ctrl+C", "Quit"},
		}},
	}

	keyStyle := lipgloss.NewStyle().Foreground(CurrentTheme.Focus).Bold(true).Width(14)
	descStyle := lipgloss.NewStyle().Foreground(CurrentTheme.Text)
	titleStyle := lipgloss.NewStyle().Foreground(CurrentTheme.Header).Bold(true).Underline(true)

	var sb strings.Builder
	sb.WriteString(HeaderStyle.Render(" GOPAC HELP "))
	sb.WriteString("\n")
	for _, sec := range sections {
		sb.WriteString("\n" + titleStyle.Render(sec.title) + "\n")
		for _, r := range sec.rows {
			fmt.Fprintf(&sb, "  %s %s\n", keyStyle.Render(r[0]), descStyle.Render(r[1]))
		}
	}
	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Render("Press ? or Esc to close"))

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

// tabLabel is the text shown for tab i in the header, with the number of
// packages it holds once known.
func (m Model) tabLabel(i int) string {
	n := 0
	if i == updatesTab {
		if !m.updatesLoaded {
			return tabs[i]
		}
		n = len(m.updates)
	} else {
		for _, it := range m.allItems {
			if matchesTab(tabs[i], it.Pkg) {
				n++
			}
		}
	}
	if n == 0 {
		return tabs[i]
	}
	return fmt.Sprintf("%s %d", tabs[i], n)
}

func (m Model) tabsView() string {
	var views []string
	for i := range tabs {
		style := lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Background(CurrentTheme.Highlight).Padding(0, 1)
		if i == m.activeTab {
			style = style.Foreground(CurrentTheme.Base).Background(CurrentTheme.Focus).Bold(true)
		}
		views = append(views, style.Render(m.tabLabel(i)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

// headerView renders the logo, search pill and tabs on exactly one line of
// m.width cells, so mouse hit-testing can rely on the layout.
func (m Model) headerView() string {
	bg := lipgloss.NewStyle().Background(CurrentTheme.Highlight)
	logo := HeaderStyle.Render(" GOPAC ")
	tabsView := m.tabsView()
	gap := bg.Render(" ")

	searchWidth := m.width - lipgloss.Width(logo) - lipgloss.Width(tabsView) - 2*lipgloss.Width(gap)
	if searchWidth < 12 {
		// Not enough room for the tabs; the active tab is still shown in the status bar.
		tabsView = ""
		searchWidth = m.width - lipgloss.Width(logo) - lipgloss.Width(gap)
	}

	fg := CurrentTheme.Gray
	if m.searching {
		fg = CurrentTheme.Focus
	}
	spin := "  "
	if m.isSearching {
		spin = m.spinner.View() + " "
	}

	// Rounded pill caps take one cell each; the rest is the input field.
	inner := max(searchWidth-2, 0)
	input := m.input
	input.Width = max(inner-5, 1) // spinner (2) + icon (2) + cursor (1)
	content := ansi.Truncate(spin+"\uf002 "+input.View(), inner, "")
	pillCap := lipgloss.NewStyle().Foreground(CurrentTheme.Base).Background(CurrentTheme.Highlight)
	search := pillCap.Render("\ue0b6") +
		lipgloss.NewStyle().Background(CurrentTheme.Base).Foreground(fg).Width(inner).Render(content) +
		pillCap.Render("\ue0b4")

	header := logo + gap + search
	if tabsView != "" {
		header += gap + tabsView
	}
	return bg.Width(m.width).Render(ansi.Truncate(header, m.width, ""))
}

func (m Model) emptyListView() string {
	bold := lipgloss.NewStyle().Bold(true)
	hint := lipgloss.NewStyle().Foreground(CurrentTheme.Gray)
	switch {
	case m.isSearching:
		return bold.Foreground(CurrentTheme.Focus).Render(m.spinner.View() + " Searching...")
	case m.activeTab == updatesTab && m.loadingUpdates:
		return bold.Foreground(CurrentTheme.Focus).Render(m.spinner.View() + " Checking for updates...")
	case m.activeTab == updatesTab:
		return bold.Foreground(CurrentTheme.Green).Render("✓ System is up to date") + "\n\n" + hint.Render("press r to check again")
	case tabs[m.activeTab] == "ORPHANS":
		return bold.Foreground(CurrentTheme.Green).Render("✓ No orphaned packages")
	case m.currentQuery == "":
		return bold.Foreground(CurrentTheme.Focus).Render("Nothing here yet") + "\n\n" + hint.Render("press / to search")
	default:
		return bold.Foreground(CurrentTheme.Red).Render("No packages found") + "\n\n" + hint.Render("try another tab or query")
	}
}

type keyHint struct{ key, desc string }

func (m Model) keyHints() (mode string, hints []keyHint) {
	switch {
	case m.searching:
		return "SEARCH", []keyHint{{"enter", "search"}, {"↑↓", "history"}, {"tab", "list"}, {"esc", "cancel"}}
	case m.focusSide == 1:
		hints = []keyHint{{"j/k", "scroll"}, {"esc", "back"}}
		if i, ok := m.list.SelectedItem().(Item); ok && i.Pkg.IsAUR {
			hints = append(hints, keyHint{"p", "PKGBUILD"})
		}
		return "DETAILS", append(hints, keyHint{"?", "help"})
	case m.activeTab == updatesTab:
		return "UPDATES", []keyHint{{"enter/U", "upgrade all"}, {"r", "recheck"}, {"h/l", "tabs"}, {"/", "search"}, {"?", "help"}}
	default:
		return "LIST", []keyHint{{"enter", "install/remove"}, {"space", "queue"}, {"a", "queue all"}, {"h/l", "tabs"}, {"/", "search"}, {"r", "refresh"}, {"?", "help"}}
	}
}

// statusBarView renders a single line: mode pill, status message and key
// hints on the left; queue summary and position on the right.
func (m Model) statusBarView() string {
	t := CurrentTheme
	base := lipgloss.NewStyle().Background(t.Base)

	mode, hints := m.keyHints()
	left := lipgloss.NewStyle().Foreground(t.Base).Background(t.Focus).Bold(true).Padding(0, 1).Render(mode)

	if m.statusMsg != "" {
		color := t.Green
		if m.statusIsErr {
			color = t.Red
		}
		left += base.Foreground(color).Bold(true).Render(" " + m.statusMsg)
	} else {
		for _, h := range hints {
			left += base.Foreground(t.Text).Bold(true).Render("  "+h.key) + base.Foreground(t.Gray).Render(" "+h.desc)
		}
	}

	var right string
	if n := len(m.markedInstall) + len(m.markedRemove); n > 0 {
		q := fmt.Sprintf(" QUEUE +%d -%d  I apply  C clear ", len(m.markedInstall), len(m.markedRemove))
		right += lipgloss.NewStyle().Foreground(t.Base).Background(t.Yellow).Bold(true).Render(q)
	}
	if total := len(m.list.Items()); total > 0 {
		right += base.Foreground(t.Gray).Render(fmt.Sprintf(" %d/%d ", m.list.Index()+1, total))
	}

	// Keep the bar on one line; wrapping would push the header off-screen.
	left = ansi.Truncate(left, max(m.width-lipgloss.Width(right), 0), "…")
	gap := base.Render(strings.Repeat(" ", max(m.width-lipgloss.Width(left)-lipgloss.Width(right), 0)))
	return ansi.Truncate(left+gap+right, m.width, "")
}
