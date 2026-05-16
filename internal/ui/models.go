package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gopac/internal/manager"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var tabs = []string{"ALL", "AUR", "OFFICIAL", "INSTALLED"}

type Item struct {
	Pkg   manager.Package
	Query string
}

func (i Item) Title() string {
	icon := " "
	if i.Pkg.IsInstalled {
		icon = "✓"
	}
	baseColor := CurrentTheme.RepoOfficial
	if i.Pkg.IsAUR {
		baseColor = CurrentTheme.RepoAUR
	}

	name := i.Pkg.Name
	var titleSB strings.Builder

	if i.Query != "" && strings.Contains(strings.ToLower(name), strings.ToLower(i.Query)) {
		lowerName := strings.ToLower(name)
		lowerQuery := strings.ToLower(i.Query)
		idx := strings.Index(lowerName, lowerQuery)

		if idx >= 0 {
			titleSB.WriteString(lipgloss.NewStyle().Foreground(baseColor).Bold(true).Render(name[:idx]))
			titleSB.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Focus).Background(CurrentTheme.Highlight).Bold(true).Underline(true).Render(name[idx : idx+len(lowerQuery)]))
			titleSB.WriteString(lipgloss.NewStyle().Foreground(baseColor).Bold(true).Render(name[idx+len(lowerQuery):]))
		} else {
			titleSB.WriteString(lipgloss.NewStyle().Foreground(baseColor).Bold(true).Render(name))
		}
	} else {
		titleSB.WriteString(lipgloss.NewStyle().Foreground(baseColor).Bold(true).Render(name))
	}

	return fmt.Sprintf("%s %s",
		lipgloss.NewStyle().Foreground(baseColor).Render(icon),
		titleSB.String(),
	)
}

func (i Item) Description() string {
	aurTag := lipgloss.NewStyle().Foreground(CurrentTheme.RepoAUR).Render("AUR")
	offTag := lipgloss.NewStyle().Foreground(CurrentTheme.RepoOfficial).Render("Official")
	tag := offTag
	if i.Pkg.IsAUR {
		tag = aurTag
	}
	return fmt.Sprintf("%s | %s", tag, i.Pkg.Version)
}

func (i Item) FilterValue() string { return i.Pkg.Name }

type (
	InstalledMapMsg  map[string]bool
	PackageDetailMsg manager.Package
	TickMsg          time.Time
)

type Model struct {
	list            list.Model
	input           textinput.Model
	viewport        viewport.Model
	spinner         spinner.Model
	searching       bool
	isSearching     bool
	allItems        []Item
	activeTab       int
	width, height   int
	listWidth       int
	descWidth       int
	panelHeight     int
	lastID          int
	currentQuery    string
	lastSelectedPkg string
	showingPKGBUILD bool
	showingHelp     bool
	focusSide       int // 0: List, 1: Detail, 2: Search
	searchCancel    context.CancelFunc
	searchHistory   []string
	historyIdx      int
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Search packages..."
	ti.CharLimit = 156
	ti.Width = 30
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(CurrentTheme.Focus)
	ti.TextStyle = lipgloss.NewStyle().Foreground(CurrentTheme.Focus)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(CurrentTheme.Focus)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(CurrentTheme.Focus).BorderLeftForeground(CurrentTheme.Focus)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(CurrentTheme.Text).BorderLeftForeground(CurrentTheme.Focus)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.DisableQuitKeybindings()
	l.SetFilteringEnabled(false)

	ti.Focus()
	return Model{
		list: l, input: ti, viewport: viewport.New(0, 0), spinner: s, searching: true, allItems: []Item{}, activeTab: 0, focusSide: 2,
		searchHistory: []string{}, historyIdx: -1,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) Init() tea.Cmd { return tea.Batch(textinput.Blink, tickCmd(), m.spinner.Tick) }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		m.panelHeight = m.height - 4
		if m.panelHeight < 5 {
			m.panelHeight = 5
		}

		m.listWidth = int(float64(m.width) * 0.35)
		m.descWidth = m.width - m.listWidth

		const chromeWidth = 6
		m.list.SetSize(m.listWidth-chromeWidth, m.panelHeight)
		m.viewport.Width = m.descWidth - chromeWidth
		m.viewport.Height = m.panelHeight

		if m.width > 60 {
			m.input.Width = m.width - 60
		} else {
			m.input.Width = 20
		}

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			if msg.Y == 0 {
				// Check if click was in the search bar area or tabs area
				// Simple approximation: tabs are on the right
				if msg.X > m.width-20 {
					m.activeTab = (m.activeTab + 1) % len(tabs)
					m.updateListItems()
				} else if msg.X > m.listWidth && msg.X < m.width-20 {
					m.focusSide = 2
					m.searching = true
					m.input.Focus()
				}
			} else {
				if msg.X < m.listWidth {
					m.focusSide = 0
					m.searching = false
					m.input.Blur()
				} else {
					m.focusSide = 1
					m.searching = false
					m.input.Blur()
				}
			}
		}
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Cycle Focus: List(0) -> Detail(1) -> Search(2)
		if msg.String() == "tab" {
			m.focusSide = (m.focusSide + 1) % 3
			if m.focusSide == 2 {
				m.searching = true
				m.input.Focus()
			} else {
				m.searching = false
				m.input.Blur()
			}
			return m, nil
		}

		if msg.String() == "shift+tab" {
			m.focusSide = (m.focusSide - 1 + 3) % 3
			if m.focusSide == 2 {
				m.searching = true
				m.input.Focus()
			} else {
				m.searching = false
				m.input.Blur()
			}
			return m, nil
		}

		if m.searching {
			if msg.String() == "enter" {
				m.searching = false
				m.input.Blur()
				m.focusSide = 0 // Auto focus list
				m.currentQuery = m.input.Value()

				if m.input.Value() != "" {
					// Add to history if not same as last
					if len(m.searchHistory) == 0 || m.searchHistory[len(m.searchHistory)-1] != m.input.Value() {
						m.searchHistory = append(m.searchHistory, m.input.Value())
					}
					m.historyIdx = len(m.searchHistory)
				}

				if m.searchCancel != nil {
					m.searchCancel()
				}
				ctx, cancel := context.WithCancel(context.Background())
				m.searchCancel = cancel
				m.isSearching = true
				return m, performSearch(ctx, m.input.Value())
			}
			if msg.String() == "esc" {
				m.searching = false
				m.input.Blur()
				m.focusSide = 0
				return m, nil
			}
			if msg.String() == "up" && len(m.searchHistory) > 0 {
				if m.historyIdx > 0 {
					m.historyIdx--
					m.input.SetValue(m.searchHistory[m.historyIdx])
					m.input.SetCursor(len(m.input.Value()))
				}
				return m, nil
			}
			if msg.String() == "down" && len(m.searchHistory) > 0 {
				if m.historyIdx < len(m.searchHistory)-1 {
					m.historyIdx++
					m.input.SetValue(m.searchHistory[m.historyIdx])
					m.input.SetCursor(len(m.input.Value()))
				} else {
					m.historyIdx = len(m.searchHistory)
					m.input.SetValue("")
				}
				return m, nil
			}

			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "?":
			m.showingHelp = !m.showingHelp
			return m, nil

		case "/":
			m.focusSide = 2
			m.searching = true
			m.input.Focus()
			return m, textinput.Blink

		case "q":
			return m, tea.Quit

		case "U":
			c := manager.UpdateSystem()
			return m, tea.ExecProcess(c, func(err error) tea.Msg { return refreshInstalledStatus() })

		case "p":
			if i, ok := m.list.SelectedItem().(Item); ok && i.Pkg.IsAUR {
				m.showingPKGBUILD = !m.showingPKGBUILD
				var fetchCmd tea.Cmd
				if m.showingPKGBUILD && i.Pkg.PKGBUILD == "" {
					fetchCmd = fetchPKGBUILD(i.Pkg)
				}
				if m.showingPKGBUILD {
					m.viewport.SetContent(renderPKGBUILD(i.Pkg, m.viewport.Width))
				} else {
					m.viewport.SetContent(renderDescription(i.Pkg, m.viewport.Width))
				}
				return m, fetchCmd
			}
		}

		if m.focusSide == 0 {
			switch msg.String() {
			case "left", "h":
				m.activeTab = (m.activeTab - 1 + len(tabs)) % len(tabs)
				m.updateListItems()
			case "right", "l":
				m.activeTab = (m.activeTab + 1) % len(tabs)
				m.updateListItems()
			case "enter":
				if i, ok := m.list.SelectedItem().(Item); ok {
					c := manager.InstallOrRemove(i.Pkg.Name, i.Pkg.IsAUR, i.Pkg.IsInstalled)
					return m, tea.ExecProcess(c, func(err error) tea.Msg { return refreshInstalledStatus() })
				}
			}
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.focusSide == 1 {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case TickMsg:
		cmds = append(cmds, tickCmd())
		if m.searching && m.input.Value() != "" && m.input.Value() != m.currentQuery {
			m.currentQuery = m.input.Value()
			if m.searchCancel != nil {
				m.searchCancel()
			}
			ctx, cancel := context.WithCancel(context.Background())
			m.searchCancel = cancel
			m.isSearching = true
			cmds = append(cmds, performSearch(ctx, m.input.Value()))
		}

	case []manager.Package:
		m.isSearching = false
		if msg != nil {
			items := make([]Item, len(msg))
			for i, pkg := range msg {
				items[i] = Item{Pkg: pkg, Query: m.currentQuery}
			}
			m.allItems = items
			m.updateListItems()
		}

	case InstalledMapMsg:
		for i := range m.allItems {
			if _, ok := msg[m.allItems[i].Pkg.Name]; ok {
				m.allItems[i].Pkg.IsInstalled = true
			} else {
				m.allItems[i].Pkg.IsInstalled = false
			}
		}
		m.updateListItems()

	case PackageDetailMsg:
		for i := range m.allItems {
			if m.allItems[i].Pkg.Name == msg.Name {
				m.allItems[i].Pkg = manager.Package(msg)
			}
		}
		m.updateListItems()
	}

	if i, ok := m.list.SelectedItem().(Item); ok {
		if i.Pkg.Name != m.lastSelectedPkg {
			m.lastSelectedPkg = i.Pkg.Name
			m.showingPKGBUILD = false
			m.viewport.GotoTop()
		}

		if m.showingPKGBUILD {
			m.viewport.SetContent(renderPKGBUILD(i.Pkg, m.viewport.Width))
		} else {
			m.viewport.SetContent(renderDescription(i.Pkg, m.viewport.Width))
		}

		if !i.Pkg.Detailed && m.lastSelectedPkg == i.Pkg.Name {
			cmds = append(cmds, fetchDetails(i.Pkg))
		}
	} else {
		m.viewport.SetContent("")
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) updateListItems() {
	var filtered []list.Item
	mode := tabs[m.activeTab]
	for _, item := range m.allItems {
		item.Query = m.currentQuery
		switch mode {
		case "ALL":
			filtered = append(filtered, item)
		case "AUR":
			if item.Pkg.IsAUR {
				filtered = append(filtered, item)
			}
		case "OFFICIAL":
			if !item.Pkg.IsAUR {
				filtered = append(filtered, item)
			}
		case "INSTALLED":
			if item.Pkg.IsInstalled {
				filtered = append(filtered, item)
			}
		}
	}
	m.list.SetItems(filtered)
}

func performSearch(ctx context.Context, query string) tea.Cmd {
	if query == "" {
		return nil
	}
	return func() tea.Msg {
		res, _ := manager.SearchContext(ctx, query)
		return res
	}
}

func refreshInstalledStatus() tea.Msg {
	manager.RefreshInstalledCache()
	return InstalledMapMsg(manager.GetInstalledCache())
}

func renderDescription(p manager.Package, width int) string {
	if !p.Detailed {
		header := lipgloss.NewStyle().Foreground(CurrentTheme.RepoOfficial).Bold(true).Render(p.Name)
		if p.IsAUR {
			header = lipgloss.NewStyle().Foreground(CurrentTheme.RepoAUR).Bold(true).Render(p.Name)
		}
		return lipgloss.NewStyle().Width(width).Render(fmt.Sprintf("\n%s\n\nLoading details...", header))
	}

	var sb strings.Builder

	keyStyle := LabelStyle.Width(16)
	valStyle := ValueStyle
	headerStyle := lipgloss.NewStyle().Foreground(CurrentTheme.RepoOfficial).Bold(true).Background(CurrentTheme.Highlight).Padding(0, 1)

	if p.IsAUR {
		headerStyle = headerStyle.Foreground(CurrentTheme.RepoAUR)
		sb.WriteString(fmt.Sprintf("\n%s\n\n", headerStyle.Render(p.Name)))

		row := func(k, v string) {
			if v == "" {
				return
			}
			sb.WriteString(fmt.Sprintf("%s : %s\n", keyStyle.Render(k), valStyle.Render(v)))
		}

		row("Repository", "AUR")
		row("Version", p.Version)
		row("Description", p.Description)
		row("URL", p.URL)
		row("Maintainer", p.Maintainer)
		row("Votes", fmt.Sprintf("%d (Pop: %.2f)", p.Votes, p.Popularity))
		row("Keywords", strings.Join(p.Keywords, "  "))
		row("Licenses", strings.Join(p.Licenses, "  "))

		if p.FirstSubmitted > 0 {
			row("Submitted", time.Unix(p.FirstSubmitted, 0).Format("2006-01-02"))
		}
		if p.LastModified > 0 {
			row("Last Modified", time.Unix(p.LastModified, 0).Format("2006-01-02"))
		}

		if len(p.Depends) > 0 {
			sb.WriteString(fmt.Sprintf("\n%s\n%s\n", lipgloss.NewStyle().Foreground(CurrentTheme.Focus).Bold(true).Render("Dependencies"), valStyle.Render(strings.Join(p.Depends, "  "))))
		}
		if len(p.MakeDepends) > 0 {
			sb.WriteString(fmt.Sprintf("%s : %s\n", keyStyle.Render("Make Deps"), valStyle.Render(strings.Join(p.MakeDepends, "  "))))
		}

		sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Render("\n[ PKGBUILD ]"))

	} else {
		sb.WriteString(fmt.Sprintf("\n%s\n\n", headerStyle.Render(p.Name)))

		row := func(k, v string) {
			sb.WriteString(fmt.Sprintf("%s : %s\n", keyStyle.Render(k), valStyle.Render(v)))
		}
		listRow := func(k string, v []string) {
			if len(v) == 0 {
				row(k, "None")
			} else {
				row(k, strings.Join(v, "  "))
			}
		}

		row("Name", p.Name)
		row("Version", p.Version)
		row("Description", p.Description)
		row("Architecture", p.Architecture)
		row("URL", p.URL)
		listRow("Licenses", p.Licenses)
		listRow("Groups", p.Groups)
		listRow("Provides", p.Provides)
		listRow("Depends On", p.Depends)
		listRow("Optional Deps", p.OptDepends)
		listRow("Required By", p.RequiredBy)
		listRow("Conflicts With", p.Conflicts)
		listRow("Replaces", p.Replaces)
		row("Download Size", p.DownloadSize)
		row("Installed Size", p.InstalledSize)
		row("Packager", p.Packager)

		dateStr := func(t int64) string {
			if t == 0 {
				return "None"
			}
			return time.Unix(t, 0).Format("Mon 02 Jan 2006 03:04:05 PM MST")
		}

		row("Build Date", dateStr(p.BuildDate))
		row("Install Date", dateStr(p.InstallDate))
		row("Install Reason", p.InstallReason)
		row("Validated By", p.ValidatedBy)
	}

	return lipgloss.NewStyle().Width(width).Render(sb.String())
}

func renderPKGBUILD(p manager.Package, width int) string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.RepoAUR).Bold(true).Render("PKGBUILD for " + p.Name))
	sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Render("  (Press 'p' to go back)\n\n"))

	if p.PKGBUILD == "" {
		sb.WriteString("Loading PKGBUILD or not available...")
	} else {
		lines := strings.Split(p.PKGBUILD, "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Render(line) + "\n")
			} else if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Blue).Render(parts[0]) + "=" + lipgloss.NewStyle().Foreground(CurrentTheme.Text).Render(parts[1]) + "\n")
			} else {
				sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Text).Render(line) + "\n")
			}
		}
	}
	return lipgloss.NewStyle().Width(width).Render(sb.String())
}

func fetchDetails(p manager.Package) tea.Cmd {
	return func() tea.Msg {
		if err := manager.GetPackageDetails(&p); err != nil {
			return nil
		}
		return PackageDetailMsg(p)
	}
}

func fetchPKGBUILD(p manager.Package) tea.Cmd {
	return func() tea.Msg {
		build, err := manager.GetPKGBUILD(p.Name)
		if err != nil {
			p.PKGBUILD = fmt.Sprintf("Error fetching PKGBUILD: %v", err)
		} else {
			p.PKGBUILD = build
		}
		return PackageDetailMsg(p)
	}
}
