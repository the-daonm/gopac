package ui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"gopac/internal/manager"

	"github.com/charmbracelet/bubbles/list"
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
	baseColor := GruvGreen
	if i.Pkg.IsAUR {
		baseColor = GruvOrange
	}

	name := i.Pkg.Name
	var titleSB strings.Builder

	if i.Query != "" && strings.Contains(strings.ToLower(name), strings.ToLower(i.Query)) {
		lowerName := strings.ToLower(name)
		lowerQuery := strings.ToLower(i.Query)
		idx := strings.Index(lowerName, lowerQuery)

		if idx >= 0 {
			titleSB.WriteString(lipgloss.NewStyle().Foreground(baseColor).Bold(true).Render(name[:idx]))
			titleSB.WriteString(lipgloss.NewStyle().Foreground(GruvYellow).Background(GruvBg).Bold(true).Underline(true).Render(name[idx : idx+len(lowerQuery)]))
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
	aurTag := lipgloss.NewStyle().Foreground(GruvOrange).Render("AUR")
	offTag := lipgloss.NewStyle().Foreground(GruvGreen).Render("Official")
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
	list                                             list.Model
	input                                            textinput.Model
	viewport                                         viewport.Model
	searching                                        bool
	allItems                                         []Item
	activeTab                                        int
	width, height, listWidth, descWidth, panelHeight int
	lastID                                           int
	currentQuery                                     string
	lastSelectedPkg                                  string
	showingPKGBUILD                                  bool
	focusSide                                        int // 0: List, 1: Detail
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(GruvYellow)
	ti.TextStyle = lipgloss.NewStyle().Foreground(GruvYellow)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(GruvYellow).BorderLeftForeground(GruvYellow)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(GruvFg).BorderLeftForeground(GruvYellow)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.DisableQuitKeybindings()
	l.SetFilteringEnabled(false)

	return Model{
		list: l, input: ti, viewport: viewport.New(0, 0), searching: true, allItems: []Item{}, activeTab: 0,
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		m.panelHeight = m.height - 11 // Adjusted for extra borders

		if m.panelHeight < 5 {
			m.panelHeight = 5
		}

		innerW := m.width - 6
		if innerW < 10 {
			innerW = 10
		}

		m.listWidth = int(float64(innerW) * 0.35)
		m.descWidth = innerW - m.listWidth - 2

		m.list.SetSize(m.listWidth-2, m.panelHeight)
		m.viewport.Height = m.panelHeight
		m.viewport.Width = m.descWidth - 2

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+l":
			m.activeTab = (m.activeTab + 1) % len(tabs)
			m.updateListItems()
		case "ctrl+h":
			m.activeTab = (m.activeTab - 1 + len(tabs)) % len(tabs)
			m.updateListItems()

		case "tab":
			// Toggle focus between List (0) and Detail (1)
			if m.focusSide == 0 {
				m.focusSide = 1
			} else {
				m.focusSide = 0
			}
			return m, nil

		case "/":
			// Enter search mode
			m.searching = true
			m.input.Focus()
			return m, textinput.Blink

		case "esc":
			if m.searching {
				m.searching = false
				m.input.Blur()
			}
			return m, nil

		case "p":
			if !m.searching {
				if i, ok := m.list.SelectedItem().(Item); ok && i.Pkg.IsAUR {
					m.showingPKGBUILD = !m.showingPKGBUILD
					
					// Auto-focus logic
					if m.showingPKGBUILD {
						m.focusSide = 1
					} else {
						m.focusSide = 0
					}

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
		}

		if m.searching {
			if msg.String() == "enter" {
				m.searching = false
				m.input.Blur()
				m.currentQuery = m.input.Value()
				return m, performSearch(m.input.Value())
			}
			if msg.String() == "esc" {
				m.searching = false
				m.input.Blur()
				return m, nil
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "j", "down":
			if m.focusSide == 0 {
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.viewport.LineDown(1)
			}

		case "k", "up":
			if m.focusSide == 0 {
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.viewport.LineUp(1)
			}

		case "ctrl+d":
			if m.focusSide == 0 {
				m.list.CursorDown()
				for i := 0; i < 5; i++ { m.list.CursorDown() }
			} else {
				m.viewport.HalfViewDown()
			}

		case "ctrl+u":
			if m.focusSide == 0 {
				m.list.CursorUp()
				for i := 0; i < 5; i++ { m.list.CursorUp() }
			} else {
				m.viewport.HalfViewUp()
			}

		case "g", "home":
			if m.focusSide == 0 {
				m.list.Select(0)
			} else {
				m.viewport.GotoTop()
			}

		case "G", "end":
			if m.focusSide == 0 {
				m.list.Select(len(m.list.Items()) - 1)
			} else {
				m.viewport.GotoBottom()
			}

		case "h", "left":
			m.list, cmd = m.list.Update(msg) // Pagination
			cmds = append(cmds, cmd)

		case "l", "right":
			// No action or pagination if list supported horizontal scrolling
			m.list, cmd = m.list.Update(msg) 
			cmds = append(cmds, cmd)

		case "pgup":
			if m.focusSide == 0 {
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.viewport.ViewUp()
			}

		case "pgdown":
			if m.focusSide == 0 {
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.viewport.ViewDown()
			}

		case "q":
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(Item); ok {
				c := manager.InstallOrRemove(i.Pkg.Name, i.Pkg.IsAUR, i.Pkg.IsInstalled)
				return m, tea.ExecProcess(c, func(err error) tea.Msg { return refreshInstalledStatus() })
			}
		}

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			if msg.X < m.listWidth+5 {
				m.focusSide = 0
			} else {
				m.focusSide = 1
			}
		}

		if msg.Type == tea.MouseWheelUp || msg.Type == tea.MouseWheelDown {
			if msg.X < m.listWidth+5 {
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		} else {
			if msg.X < m.listWidth+5 {
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}

	case TickMsg:
		if m.searching && m.input.Value() != "" && m.input.Value() != m.currentQuery {
			m.currentQuery = m.input.Value()
			return m, performSearch(m.input.Value())
		}

	case []manager.Package:
		items := make([]Item, len(msg))
		for i, pkg := range msg {
			items[i] = Item{Pkg: pkg, Query: m.currentQuery}
		}
		m.allItems = items
		m.updateListItems()

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

func performSearch(query string) tea.Cmd {
	if query == "" {
		return nil
	}
	return func() tea.Msg {
		res, _ := manager.Search(query)
		return res
	}
}

func refreshInstalledStatus() tea.Msg {
	out, err := exec.Command("pacman", "-Qq").Output()
	if err != nil {
		return InstalledMapMsg{}
	}
	installed := make(InstalledMapMsg)
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" {
			installed[line] = true
		}
	}
	return installed
}

func renderDescription(p manager.Package, width int) string {
	if !p.Detailed {
		header := lipgloss.NewStyle().Foreground(GruvGreen).Bold(true).Render(p.Name)
		if p.IsAUR {
			header = lipgloss.NewStyle().Foreground(GruvOrange).Bold(true).Render(p.Name)
		}
		return lipgloss.NewStyle().Width(width).Render(fmt.Sprintf("\n%s\n\nLoading details...", header))
	}

	var sb strings.Builder

	keyStyle := LabelStyle.Width(16)
	valStyle := ValueStyle
	headerStyle := lipgloss.NewStyle().Foreground(GruvGreen).Bold(true).Background(GruvBg).Padding(0, 1)

	if p.IsAUR {
		headerStyle = headerStyle.Foreground(GruvOrange)

		// AUR Website Style
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
			sb.WriteString(fmt.Sprintf("\n%s\n%s\n", lipgloss.NewStyle().Foreground(GruvYellow).Bold(true).Render("Dependencies"), valStyle.Render(strings.Join(p.Depends, "  "))))
		}
		if len(p.MakeDepends) > 0 {
			sb.WriteString(fmt.Sprintf("%s : %s\n", keyStyle.Render("Make Deps"), valStyle.Render(strings.Join(p.MakeDepends, "  "))))
		}

		sb.WriteString(lipgloss.NewStyle().Foreground(GruvGray).Render("\n[ PKGBUILD ]"))

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
	sb.WriteString(lipgloss.NewStyle().Foreground(GruvOrange).Bold(true).Render("PKGBUILD for " + p.Name))
	sb.WriteString(lipgloss.NewStyle().Foreground(GruvGray).Render("  (Press 'p' to go back)\n\n"))

	if p.PKGBUILD == "" {
		sb.WriteString("Loading PKGBUILD or not available...")
	} else {
		lines := strings.Split(p.PKGBUILD, "\n")
		for _, line := range lines {
			// Basic syntax highlighting or just render
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				sb.WriteString(lipgloss.NewStyle().Foreground(GruvGray).Render(line) + "\n")
			} else if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				sb.WriteString(lipgloss.NewStyle().Foreground(GruvBlue).Render(parts[0]) + "=" + lipgloss.NewStyle().Foreground(GruvFg).Render(parts[1]) + "\n")
			} else {
				sb.WriteString(lipgloss.NewStyle().Foreground(GruvFg).Render(line) + "\n")
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
		if build, err := manager.GetPKGBUILD(p.Name); err == nil {
			p.PKGBUILD = build
			return PackageDetailMsg(p)
		}
		return nil
	}
}
