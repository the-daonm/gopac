package ui

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"gopac/internal/history"
	"gopac/internal/manager"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var tabs = []string{"ALL", "AUR", "OFFICIAL", "INSTALLED", "UPDATES", "ORPHANS"}

const updatesTab = 4

type Item struct {
	Pkg        manager.Package
	Query      string
	MarkedInst bool
	MarkedRem  bool
	DetailErr  string
}

func (i Item) Title() string {
	icon := " "
	baseColor := CurrentTheme.RepoOfficial
	if i.Pkg.IsAUR {
		baseColor = CurrentTheme.RepoAUR
	}
	iconColor := baseColor

	if i.MarkedInst {
		icon = ""
		iconColor = CurrentTheme.Green
	} else if i.MarkedRem {
		icon = ""
		iconColor = CurrentTheme.Red
	} else if i.Pkg.IsInstalled {
		icon = "✓"
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
		lipgloss.NewStyle().Foreground(iconColor).Render(icon),
		titleSB.String(),
	)
}

func (i Item) Description() string {
	t := CurrentTheme
	sep := lipgloss.NewStyle().Foreground(t.Gray).Render(" · ")
	ver := lipgloss.NewStyle().Foreground(t.Text)

	parts := []string{lipgloss.NewStyle().Foreground(GetRepoColor(i.Pkg.IsAUR)).Render(repoLabel(i.Pkg))}
	if i.Pkg.OldVersion != "" {
		parts = append(parts, lipgloss.NewStyle().Foreground(t.Gray).Render(i.Pkg.OldVersion)+
			lipgloss.NewStyle().Foreground(t.Green).Render(" → "+i.Pkg.Version))
	} else {
		parts = append(parts, ver.Render(i.Pkg.Version))
	}
	if i.Pkg.IsAUR && i.Pkg.Votes > 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(t.Yellow).Render(fmt.Sprintf("★ %d", i.Pkg.Votes)))
	}
	if i.Pkg.OutOfDate > 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(t.Red).Render("out of date"))
	}
	if i.Pkg.IsOrphan {
		parts = append(parts, lipgloss.NewStyle().Foreground(t.Orange).Render("orphan"))
	}
	return strings.Join(parts, sep)
}

func repoLabel(p manager.Package) string {
	if p.IsAUR {
		return "AUR"
	}
	return "Official"
}

func (i Item) FilterValue() string { return i.Pkg.Name }

type (
	TickMsg time.Time
)

type installedStateMsg struct {
	installed map[string]bool
	orphans   map[string]bool
}

type detailsMsg struct {
	pkg manager.Package
	err error
}

type pkgbuildMsg struct {
	name    string
	content string
}

// execDoneMsg reports the result of an external pacman/helper command.
type execDoneMsg struct {
	bulk bool
	err  error
}

type updatesMsg struct {
	pkgs []manager.Package
	err  error
}

type installedListMsg struct {
	pkgs []manager.Package
	err  error
}

type searchResultsMsg struct {
	query string
	pkgs  []manager.Package
	err   error
}

type Model struct {
	list                   list.Model
	input                  textinput.Model
	viewport               viewport.Model
	spinner                spinner.Model
	searching              bool
	isSearching            bool
	allItems               []Item
	activeTab              int
	width, height          int
	listWidth              int
	descWidth              int
	panelHeight            int
	currentQuery           string
	lastSelectedPkg        string
	showingPKGBUILD        bool
	showingHelp            bool
	showingConfirm         bool
	confirmIsBulk          bool
	confirmInstallOfficial []string
	confirmInstallAUR      []string
	confirmRemove          []string
	focusSide              int // 0: List, 1: Detail, 2: Search
	searchCancel           context.CancelFunc
	searchHistory          []string
	historyIdx             int
	markedInstall          map[string]manager.Package
	markedRemove           map[string]manager.Package
	loadingDetailsFor      string
	statusMsg              string
	updates                []Item
	updatesLoaded          bool
	loadingUpdates         bool
	statusIsErr            bool
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Search packages..."
	ti.CharLimit = 156
	ti.Width = 30
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(CurrentTheme.Focus)
	ti.TextStyle = lipgloss.NewStyle().Foreground(CurrentTheme.Focus)

	s := spinner.New()
	s.Spinner = spinner.Line
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
		searchHistory: history.Load(), historyIdx: -1,
		markedInstall:     make(map[string]manager.Package),
		markedRemove:      make(map[string]manager.Package),
		loadingDetailsFor: "",
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickCmd(), m.spinner.Tick, loadInstalled)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		m.panelHeight = max(m.height-4, 5)

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
		switch {
		case msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft:
			if msg.Y == 0 {
				if t := m.tabAt(msg.X); t >= 0 {
					cmds = append(cmds, m.setTab(t))
				} else if msg.X >= lipgloss.Width(HeaderStyle.Render(" GOPAC ")) {
					m.focusSide = 2
					m.searching = true
					m.input.Focus()
					m.historyIdx = len(m.searchHistory)
				}
			} else if msg.X < m.listWidth {
				m.focusSide = 0
				m.searching = false
				m.input.Blur()
				if idx, ok := m.listIndexAt(msg.Y); ok {
					m.list.Select(idx)
				}
			} else {
				m.focusSide = 1
				m.searching = false
				m.input.Blur()
			}
		case msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown:
			// Scroll whichever panel is under the pointer.
			if msg.X < m.listWidth {
				if msg.Button == tea.MouseButtonWheelUp {
					m.list.CursorUp()
				} else {
					m.list.CursorDown()
				}
			} else {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.statusMsg = ""

		if m.showingHelp {
			switch msg.String() {
			case "?", "esc":
				m.showingHelp = false
			}
			return m, nil
		}

		if m.showingConfirm {
			switch msg.String() {
			case "y", "Y", "enter":
				m.showingConfirm = false
				if len(m.confirmInstallAUR) > 0 && !manager.HasAURHelper() {
					m.setStatus("No AUR helper found: install paru or yay, or set aur_helper in the config", true)
					return m, nil
				}
				if len(m.confirmInstallOfficial) > 0 || len(m.confirmInstallAUR) > 0 || len(m.confirmRemove) > 0 {
					c := manager.BulkActionCmd(m.confirmInstallOfficial, m.confirmInstallAUR, m.confirmRemove)
					if c != nil {
						if m.confirmIsBulk {
							m.confirmInstallOfficial = nil
							m.confirmInstallAUR = nil
							m.confirmRemove = nil
							return m, execCmd(c, true)
						} else {
							var singlePkg string
							if len(m.confirmInstallOfficial) > 0 {
								singlePkg = m.confirmInstallOfficial[0]
							} else if len(m.confirmInstallAUR) > 0 {
								singlePkg = m.confirmInstallAUR[0]
							} else if len(m.confirmRemove) > 0 {
								singlePkg = m.confirmRemove[0]
							}
							delete(m.markedInstall, singlePkg)
							delete(m.markedRemove, singlePkg)

							m.confirmInstallOfficial = nil
							m.confirmInstallAUR = nil
							m.confirmRemove = nil
							return m, execCmd(c, false)
						}
					}
				}
			case "n", "N", "esc":
				m.showingConfirm = false
				m.confirmInstallOfficial = nil
				m.confirmInstallAUR = nil
				m.confirmRemove = nil
			}
			return m, nil
		}

		// Cycle Focus: List(0) -> Detail(1) -> Search(2)
		if msg.String() == "tab" {
			m.focusSide = (m.focusSide + 1) % 3
			if m.focusSide == 2 {
				m.searching = true
				m.input.Focus()
				m.historyIdx = len(m.searchHistory)
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
				m.historyIdx = len(m.searchHistory)
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
				m.currentQuery = strings.TrimSpace(m.input.Value())

				if m.searchCancel != nil {
					m.searchCancel()
					m.searchCancel = nil
				}

				if m.currentQuery == "" {
					m.isSearching = false
					return m, loadInstalled
				}

				if m.input.Value() != "" {
					// Add to history if not same as last
					if len(m.searchHistory) == 0 || m.searchHistory[len(m.searchHistory)-1] != m.input.Value() {
						m.searchHistory = append(m.searchHistory, m.input.Value())
						cmds = append(cmds, saveHistory(m.searchHistory))
					}
					m.historyIdx = len(m.searchHistory)
				}

				ctx, cancel := context.WithCancel(context.Background())
				m.searchCancel = cancel
				m.isSearching = true
				cmds = append(cmds, performSearch(ctx, m.currentQuery))
				return m, tea.Batch(cmds...)
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
			m.historyIdx = len(m.searchHistory)
			return m, textinput.Blink

		case "U":
			c := manager.UpdateSystem()
			return m, execCmd(c, false)

		case "I":
			if len(m.markedInstall) > 0 || len(m.markedRemove) > 0 {
				var toInstallOfficial []string
				var toInstallAUR []string
				var toRemove []string

				for name, pkg := range m.markedInstall {
					if pkg.IsAUR {
						toInstallAUR = append(toInstallAUR, name)
					} else {
						toInstallOfficial = append(toInstallOfficial, name)
					}
				}
				for name := range m.markedRemove {
					toRemove = append(toRemove, name)
				}

				m.confirmInstallOfficial = toInstallOfficial
				m.confirmInstallAUR = toInstallAUR
				m.confirmRemove = toRemove
				m.confirmIsBulk = true
				m.showingConfirm = true
			}
			return m, nil

		case "r":
			// Drop caches and reload everything from pacman and the AUR.
			m.updatesLoaded = false
			cmds = append(cmds, refreshInstalledStatus, m.loadUpdatesIfNeeded())
			if m.currentQuery != "" {
				if m.searchCancel != nil {
					m.searchCancel()
				}
				ctx, cancel := context.WithCancel(context.Background())
				m.searchCancel = cancel
				m.isSearching = true
				cmds = append(cmds, performSearch(ctx, m.currentQuery))
			}
			m.setStatus("Refreshing...", false)
			return m, tea.Batch(cmds...)

		case "C":
			m.markedInstall = make(map[string]manager.Package)
			m.markedRemove = make(map[string]manager.Package)
			m.updateListItems()
			return m, nil

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
					m.viewport.SetContent(renderDescription(i, m.viewport.Width))
				}
				return m, fetchCmd
			}
		}

		switch m.focusSide {
		case 0:
			switch msg.String() {
			case "left", "h":
				cmds = append(cmds, m.setTab((m.activeTab-1+len(tabs))%len(tabs)))
			case "right", "l":
				cmds = append(cmds, m.setTab((m.activeTab+1)%len(tabs)))
			case "enter":
				if m.activeTab == updatesTab {
					// Arch does not support partial upgrades.
					return m, execCmd(manager.UpdateSystem(), false)
				}
				if i, ok := m.list.SelectedItem().(Item); ok {
					m.confirmIsBulk = false
					m.confirmRemove = nil
					m.confirmInstallOfficial = nil
					m.confirmInstallAUR = nil

					if i.Pkg.IsInstalled {
						m.confirmRemove = []string{i.Pkg.Name}
					} else {
						if i.Pkg.IsAUR {
							m.confirmInstallAUR = []string{i.Pkg.Name}
						} else {
							m.confirmInstallOfficial = []string{i.Pkg.Name}
						}
					}
					m.showingConfirm = true
					return m, nil
				}
			case "a":
				if m.activeTab == updatesTab {
					return m, nil
				}
				m.toggleQueueAll()
				return m, nil
			case " ":
				if m.activeTab == updatesTab {
					m.setStatus("Partial upgrades are unsupported on Arch; press Enter or U to upgrade everything", true)
					return m, nil
				}
				if i, ok := m.list.SelectedItem().(Item); ok {
					name := i.Pkg.Name
					if i.Pkg.IsInstalled {
						if _, exists := m.markedRemove[name]; exists {
							delete(m.markedRemove, name)
						} else {
							m.markedRemove[name] = i.Pkg
						}
					} else {
						if _, exists := m.markedInstall[name]; exists {
							delete(m.markedInstall, name)
						} else {
							m.markedInstall[name] = i.Pkg
						}
					}
					m.updateListItems()
				}
				return m, nil
			}
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		case 1:
			switch msg.String() {
			case "esc":
				m.focusSide = 0
				return m, nil
			case "up", "k":
				m.viewport.LineUp(1)
			case "down", "j":
				m.viewport.LineDown(1)
			case "ctrl+u":
				m.viewport.LineUp(10)
			case "ctrl+d":
				m.viewport.LineDown(10)
			}
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case TickMsg:
		cmds = append(cmds, tickCmd())
		if m.searching {
			trimmedVal := strings.TrimSpace(m.input.Value())
			if trimmedVal != m.currentQuery {
				m.currentQuery = trimmedVal
				if m.searchCancel != nil {
					m.searchCancel()
					m.searchCancel = nil
				}
				if m.currentQuery == "" {
					m.isSearching = false
					cmds = append(cmds, loadInstalled)
				} else {
					ctx, cancel := context.WithCancel(context.Background())
					m.searchCancel = cancel
					m.isSearching = true
					cmds = append(cmds, performSearch(ctx, trimmedVal))
				}
			}
		}

	case searchResultsMsg:
		if msg.query != m.currentQuery || errors.Is(msg.err, context.Canceled) {
			// Outdated or superseded search result, ignore it!
			return m, nil
		}
		m.isSearching = false
		if msg.err != nil {
			m.setStatus("Search failed: "+msg.err.Error(), true)
		}
		// Results may be partial (e.g. AUR unreachable), still show them.
		items := make([]Item, len(msg.pkgs))
		for i, pkg := range msg.pkgs {
			items[i] = Item{Pkg: pkg, Query: msg.query}
		}
		m.allItems = items
		m.updateListItems()

	case installedStateMsg:
		for i := range m.allItems {
			m.allItems[i].Pkg.IsInstalled = msg.installed[m.allItems[i].Pkg.Name]
			m.allItems[i].Pkg.IsOrphan = msg.orphans[m.allItems[i].Pkg.Name]
			// Versions and install metadata may have changed; refetch lazily.
			m.allItems[i].Pkg.Detailed = false
			m.allItems[i].DetailErr = ""
		}
		m.loadingDetailsFor = ""
		m.updateListItems()
		if m.currentQuery == "" {
			cmds = append(cmds, loadInstalled)
		}

	case installedListMsg:
		if m.currentQuery != "" {
			// The user started searching meanwhile.
			return m, nil
		}
		if msg.err != nil {
			m.setStatus("Failed to list installed packages: "+msg.err.Error(), true)
		}
		items := make([]Item, len(msg.pkgs))
		for i, pkg := range msg.pkgs {
			items[i] = Item{Pkg: pkg}
		}
		m.allItems = items
		m.updateListItems()

	case detailsMsg:
		if msg.pkg.Name == m.loadingDetailsFor {
			m.loadingDetailsFor = ""
		}
		for _, it := range m.items() {
			// The same name can exist both in the repos and in the AUR.
			if it.Pkg.Name != msg.pkg.Name || it.Pkg.IsAUR != msg.pkg.IsAUR {
				continue
			}
			if msg.err != nil {
				it.DetailErr = msg.err.Error()
				continue
			}
			// Keep state that may have changed while the fetch was running.
			pkg := msg.pkg
			pkg.IsInstalled = it.Pkg.IsInstalled
			pkg.PKGBUILD = it.Pkg.PKGBUILD
			if it.Pkg.OldVersion != "" {
				// Pending update: keep showing installed -> new version.
				pkg.OldVersion, pkg.Version = it.Pkg.OldVersion, it.Pkg.Version
			}
			it.Pkg = pkg
			it.DetailErr = ""
		}
		m.updateListItems()

	case pkgbuildMsg:
		for _, it := range m.items() {
			if it.Pkg.Name == msg.name && it.Pkg.IsAUR {
				it.Pkg.PKGBUILD = msg.content
			}
		}
		m.updateListItems()

	case updatesMsg:
		m.loadingUpdates = false
		m.updatesLoaded = true
		if msg.err != nil {
			m.setStatus("Checking updates: "+msg.err.Error(), true)
		}
		m.updates = make([]Item, len(msg.pkgs))
		for i, pkg := range msg.pkgs {
			m.updates[i] = Item{Pkg: pkg}
		}
		m.updateListItems()

	case execDoneMsg:
		if msg.err != nil {
			// Keep the queue so the user can retry after fixing the problem.
			m.setStatus("Command failed: "+msg.err.Error(), true)
		} else {
			if msg.bulk {
				m.markedInstall = make(map[string]manager.Package)
				m.markedRemove = make(map[string]manager.Package)
			}
			m.setStatus("Done", false)
		}
		m.updatesLoaded = false
		loadUpdates := m.loadUpdatesIfNeeded()
		return m, tea.Batch(refreshInstalledStatus, loadUpdates)
	}

	if i, ok := m.list.SelectedItem().(Item); ok {
		if i.Pkg.Name != m.lastSelectedPkg {
			m.lastSelectedPkg = i.Pkg.Name
			m.showingPKGBUILD = false
			m.loadingDetailsFor = ""
			m.viewport.GotoTop()
		}

		if m.showingPKGBUILD {
			m.viewport.SetContent(renderPKGBUILD(i.Pkg, m.viewport.Width))
		} else {
			m.viewport.SetContent(renderDescription(i, m.viewport.Width))
		}

		if !i.Pkg.Detailed && i.DetailErr == "" && m.loadingDetailsFor != i.Pkg.Name {
			m.loadingDetailsFor = i.Pkg.Name
			cmds = append(cmds, fetchDetails(i.Pkg))
		}
	} else {
		m.viewport.SetContent("")
	}
	return m, tea.Batch(cmds...)
}

// tabAt returns the index of the header tab at column x, or -1. Tabs are
// right-aligned in the header (see headerView).
func (m Model) tabAt(x int) int {
	tabsWidth := lipgloss.Width(m.tabsView())
	start := m.width - tabsWidth
	for i := range tabs {
		w := lipgloss.Width(m.tabLabel(i)) + 2
		if x >= start && x < start+w {
			return i
		}
		start += w
	}
	return -1
}

// listIndexAt maps a screen row to the index of the list item drawn there.
// The list starts below the header and the panel's top border; the default
// delegate draws each item as two lines followed by one spacer line.
func (m Model) listIndexAt(y int) (int, bool) {
	const listTop, rowsPerItem = 2, 3
	if y < listTop || (y-listTop)%rowsPerItem == rowsPerItem-1 {
		return 0, false
	}
	idx := m.list.Paginator.Page*m.list.Paginator.PerPage + (y-listTop)/rowsPerItem
	if idx >= len(m.list.Items()) || (y-listTop)/rowsPerItem >= m.list.Paginator.PerPage {
		return 0, false
	}
	return idx, true
}

// toggleQueueAll queues every visible package (install or remove depending on
// its state), or unqueues them all if they are already queued.
func (m *Model) toggleQueueAll() {
	visible := m.list.Items()
	allQueued := len(visible) > 0
	for _, li := range visible {
		i := li.(Item)
		if !i.MarkedInst && !i.MarkedRem {
			allQueued = false
			break
		}
	}
	for _, li := range visible {
		p := li.(Item).Pkg
		switch {
		case allQueued:
			delete(m.markedInstall, p.Name)
			delete(m.markedRemove, p.Name)
		case p.IsInstalled:
			m.markedRemove[p.Name] = p
		default:
			m.markedInstall[p.Name] = p
		}
	}
	m.updateListItems()
}

// items returns pointers to every item the model holds, across data sets.
func (m *Model) items() []*Item {
	all := make([]*Item, 0, len(m.allItems)+len(m.updates))
	for i := range m.allItems {
		all = append(all, &m.allItems[i])
	}
	for i := range m.updates {
		all = append(all, &m.updates[i])
	}
	return all
}

func (m *Model) setTab(t int) tea.Cmd {
	m.activeTab = t
	m.updateListItems()
	return m.loadUpdatesIfNeeded()
}

// loadUpdatesIfNeeded starts an update check when the updates tab is shown
// and its data is missing or stale.
func (m *Model) loadUpdatesIfNeeded() tea.Cmd {
	if m.activeTab != updatesTab || m.updatesLoaded || m.loadingUpdates {
		return nil
	}
	m.loadingUpdates = true
	return func() tea.Msg {
		pkgs, err := manager.ListUpdates(context.Background())
		return updatesMsg{pkgs: pkgs, err: err}
	}
}

func (m *Model) updateListItems() {
	var filtered []list.Item
	mode := tabs[m.activeTab]

	if mode == "UPDATES" {
		q := strings.ToLower(m.currentQuery)
		for i := range m.updates {
			m.updates[i].Query = m.currentQuery
			if strings.Contains(strings.ToLower(m.updates[i].Pkg.Name), q) {
				filtered = append(filtered, m.updates[i])
			}
		}
		m.list.SetItems(filtered)
		return
	}

	for i := range m.allItems {
		m.allItems[i].Query = m.currentQuery
		_, m.allItems[i].MarkedInst = m.markedInstall[m.allItems[i].Pkg.Name]
		_, m.allItems[i].MarkedRem = m.markedRemove[m.allItems[i].Pkg.Name]
		if matchesTab(mode, m.allItems[i].Pkg) {
			filtered = append(filtered, m.allItems[i])
		}
	}
	m.list.SetItems(filtered)
}

// matchesTab reports whether p belongs in the given (non-updates) tab.
func matchesTab(tab string, p manager.Package) bool {
	switch tab {
	case "AUR":
		return p.IsAUR
	case "OFFICIAL":
		return !p.IsAUR
	case "INSTALLED":
		return p.IsInstalled
	case "ORPHANS":
		return p.IsOrphan
	default:
		return true
	}
}

func performSearch(ctx context.Context, query string) tea.Cmd {
	if query == "" {
		return nil
	}
	return func() tea.Msg {
		res, err := manager.SearchContext(ctx, query)
		return searchResultsMsg{query: query, pkgs: res, err: err}
	}
}

func (m *Model) setStatus(msg string, isErr bool) {
	m.statusMsg = msg
	m.statusIsErr = isErr
}

func execCmd(c *exec.Cmd, bulk bool) tea.Cmd {
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return execDoneMsg{bulk: bulk, err: err}
	})
}

func saveHistory(entries []string) tea.Cmd {
	entries = append([]string(nil), entries...)
	return func() tea.Msg {
		_ = history.Save(entries) // best effort; history is a convenience
		return nil
	}
}

func loadInstalled() tea.Msg {
	pkgs, err := manager.ListInstalled()
	return installedListMsg{pkgs: pkgs, err: err}
}

func refreshInstalledStatus() tea.Msg {
	manager.InvalidateCaches()
	manager.RefreshInstalledCache()
	return installedStateMsg{
		installed: manager.GetInstalledCache(),
		orphans:   manager.GetOrphanCache(),
	}
}

func renderDescription(item Item, width int) string {
	t := CurrentTheme
	p := item.Pkg
	var sb strings.Builder

	// Title line: name pill and version.
	name := lipgloss.NewStyle().Foreground(t.Base).Background(GetRepoColor(p.IsAUR)).Bold(true).Padding(0, 1).Render(p.Name)
	version := lipgloss.NewStyle().Foreground(t.Gray).Render(p.Version)
	if p.OldVersion != "" {
		version = lipgloss.NewStyle().Foreground(t.Gray).Render(p.OldVersion) +
			lipgloss.NewStyle().Foreground(t.Green).Bold(true).Render(" → "+p.Version)
	}
	sb.WriteString("\n" + name + "  " + version + "\n\n")

	// Badges summarise the package state at a glance.
	badge := func(text string, color lipgloss.Color) string {
		return lipgloss.NewStyle().Foreground(color).Background(t.Highlight).Bold(true).Padding(0, 1).Render(text)
	}
	badges := []string{badge(repoLabel(p), GetRepoColor(p.IsAUR))}
	if p.IsInstalled {
		badges = append(badges, badge("✓ installed", t.Green))
	}
	if p.OldVersion != "" {
		badges = append(badges, badge("↑ update", t.Blue))
	}
	if p.OutOfDate > 0 {
		badges = append(badges, badge("out of date", t.Red))
	}
	if p.IsOrphan {
		badges = append(badges, badge("orphan", t.Orange))
	}
	if p.IsAUR && p.Detailed && p.Maintainer == "" {
		badges = append(badges, badge("unmaintained", t.Red))
	}
	sb.WriteString(strings.Join(badges, " ") + "\n")

	if p.Description != "" {
		sb.WriteString("\n" + lipgloss.NewStyle().Foreground(t.Text).Italic(true).Width(width).Render(p.Description) + "\n")
	}

	if !p.Detailed {
		body := lipgloss.NewStyle().Foreground(t.Gray).Render("Loading details...")
		if item.DetailErr != "" {
			body = lipgloss.NewStyle().Foreground(t.Red).Render("Failed to load details: " + item.DetailErr)
		}
		sb.WriteString("\n" + body)
		return lipgloss.NewStyle().Width(width).Render(sb.String())
	}

	const labelWidth = 16
	labelStyle := lipgloss.NewStyle().Foreground(t.Gray).Width(labelWidth)
	valueStyle := lipgloss.NewStyle().Foreground(t.Text).Width(max(width-labelWidth, 10))
	section := func(title string) {
		rule := strings.Repeat("─", max(width-lipgloss.Width(title)-1, 0))
		sb.WriteString("\n" + lipgloss.NewStyle().Foreground(t.Focus).Bold(true).Render(title) + " " +
			lipgloss.NewStyle().Foreground(t.Highlight).Render(rule) + "\n")
	}
	row := func(k, v string) {
		if v == "" {
			return
		}
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(k), valueStyle.Render(v)) + "\n")
	}
	listRow := func(k string, v []string, sep string) {
		row(k, strings.Join(v, sep))
	}
	date := func(ts int64, layout string) string {
		if ts == 0 {
			return ""
		}
		return time.Unix(ts, 0).Format(layout)
	}

	section("Info")
	row("URL", lipgloss.NewStyle().Foreground(t.Blue).Underline(true).Render(p.URL))
	if p.IsAUR {
		maintainer := p.Maintainer
		if maintainer == "" {
			maintainer = lipgloss.NewStyle().Foreground(t.Red).Render("none (orphaned)")
		}
		row("Maintainer", maintainer)
		row("Votes", fmt.Sprintf("%d  (popularity %.2f)", p.Votes, p.Popularity))
		listRow("Keywords", p.Keywords, "  ")
	} else {
		row("Architecture", p.Architecture)
		row("Packager", p.Packager)
	}
	listRow("Licenses", p.Licenses, "  ")
	listRow("Groups", p.Groups, "  ")
	row("Download Size", p.DownloadSize)
	row("Installed Size", p.InstalledSize)
	row("Install Reason", p.InstallReason)
	row("Validated By", p.ValidatedBy)
	row("Submitted", date(p.FirstSubmitted, "2006-01-02"))
	row("Last Modified", date(p.LastModified, "2006-01-02"))
	row("Build Date", date(p.BuildDate, "2006-01-02 15:04"))
	row("Install Date", date(p.InstallDate, "2006-01-02 15:04"))
	if p.OutOfDate > 0 {
		row("Out of Date", lipgloss.NewStyle().Foreground(t.Red).Render("flagged on "+date(p.OutOfDate, "2006-01-02")))
	}

	if len(p.Depends)+len(p.OptDepends)+len(p.MakeDepends)+len(p.CheckDepends)+len(p.RequiredBy)+
		len(p.Provides)+len(p.Conflicts)+len(p.Replaces) > 0 {
		section("Dependencies")
		listRow("Depends On", p.Depends, "  ")
		listRow("Optional", p.OptDepends, "\n")
		listRow("Make Deps", p.MakeDepends, "  ")
		listRow("Check Deps", p.CheckDepends, "  ")
		listRow("Required By", p.RequiredBy, "  ")
		listRow("Provides", p.Provides, "  ")
		listRow("Conflicts", p.Conflicts, "  ")
		listRow("Replaces", p.Replaces, "  ")
	}

	if p.IsAUR {
		sb.WriteString("\n" + lipgloss.NewStyle().Foreground(t.Gray).Render("press p to review the PKGBUILD before installing"))
	}

	return lipgloss.NewStyle().Width(width).Render(sb.String())
}

func renderPKGBUILD(p manager.Package, width int) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.RepoAUR).Bold(true).Render("PKGBUILD for " + p.Name))
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Render("(Press 'p' to go back)"))
	sb.WriteString("\n\n")

	if p.PKGBUILD == "" {
		sb.WriteString("Loading PKGBUILD or not available...")
	} else {
		for line := range strings.SplitSeq(p.PKGBUILD, "\n") {
			line = strings.TrimSuffix(line, "\r")
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Gray).Render(line))
				sb.WriteByte('\n')
			} else if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Blue).Render(parts[0]))
				sb.WriteByte('=')
				sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Text).Render(parts[1]))
				sb.WriteByte('\n')
			} else {
				sb.WriteString(lipgloss.NewStyle().Foreground(CurrentTheme.Text).Render(line))
				sb.WriteByte('\n')
			}
		}
	}
	return sb.String()
}

func fetchDetails(p manager.Package) tea.Cmd {
	return func() tea.Msg {
		err := manager.GetPackageDetails(&p)
		return detailsMsg{pkg: p, err: err}
	}
}

func fetchPKGBUILD(p manager.Package) tea.Cmd {
	return func() tea.Msg {
		build, err := manager.GetPKGBUILD(p.Name)
		if err != nil {
			build = fmt.Sprintf("Error fetching PKGBUILD: %v", err)
		}
		return pkgbuildMsg{name: p.Name, content: build}
	}
}
