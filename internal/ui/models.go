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
	InstalledMapMsg map[string]bool
	TickMsg         time.Time
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

		m.panelHeight = m.height - 9

		if m.panelHeight < 5 {
			m.panelHeight = 5
		}

		innerW := m.width - 6
		if innerW < 10 {
			innerW = 10
		}

		m.listWidth = int(float64(innerW) * 0.35)
		m.descWidth = innerW - m.listWidth - 1

		m.list.SetSize(m.listWidth, m.panelHeight)
		m.viewport.Height = m.panelHeight
		m.viewport.Width = m.descWidth - 4

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+l":
			m.activeTab = (m.activeTab + 1) % len(tabs)
			m.updateListItems()
			return m, nil
		case "ctrl+h":
			m.activeTab = (m.activeTab - 1 + len(tabs)) % len(tabs)
			m.updateListItems()
			return m, nil

		case "tab":
			m.searching = !m.searching
			if m.searching {
				m.input.Focus()
				return m, textinput.Blink
			}
			m.input.Blur()
			return m, nil
		}

		if m.searching {
			switch msg.String() {
			case "enter":
				m.searching = false
				m.input.Blur()
				return m, performSearch(m.input.Value())
			}
			m.input, cmd = m.input.Update(msg)
			m.lastID++
			cmds = append(cmds, cmd, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg { return TickMsg(t) }))
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(Item); ok {
				c := manager.InstallOrRemove(i.Pkg.Name, i.Pkg.IsAUR, i.Pkg.IsInstalled)
				return m, tea.ExecProcess(c, func(err error) tea.Msg { return refreshInstalledStatus() })
			}
		}
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

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
	}

	if i, ok := m.list.SelectedItem().(Item); ok {
		m.viewport.SetContent(renderDescription(i.Pkg))
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

func renderDescription(p manager.Package) string {
	aurTag := lipgloss.NewStyle().Foreground(GruvOrange).Bold(true).Render("AUR")
	officialTag := lipgloss.NewStyle().Foreground(GruvGreen).Bold(true).Render("OFFICIAL")
	repoDisplay := officialTag
	titleColor := GruvGreen
	if p.IsAUR {
		repoDisplay = aurTag
		titleColor = GruvOrange
	}

	dateStr := "N/A"
	if p.LastModified > 0 {
		dateStr = time.Unix(p.LastModified, 0).Format("2006-01-02")
	}

	maintainer := p.Maintainer
	if maintainer == "" {
		maintainer = "None"
	}

	row := func(key, val string) string {
		return fmt.Sprintf("%s %s", LabelStyle.Render(key+":"), ValueStyle.Render(val))
	}

	header := lipgloss.NewStyle().
		Foreground(titleColor).
		Background(GruvBg).
		Bold(true).
		Padding(0, 1).
		Render(p.Name)

	details := []string{
		row("Repository", repoDisplay),
		row("Version", p.Version),
		row("Maintainer", maintainer),
		row("Votes", fmt.Sprintf("%d", p.Votes)),
		row("Updated", dateStr),
		row("URL", LinkStyle.Render(p.URL)),
	}

	descBlock := lipgloss.NewStyle().
		Foreground(GruvFg).
		PaddingTop(1).
		Render(p.Description)

	return fmt.Sprintf(
		"\n%s\n\n%s\n%s",
		header,
		strings.Join(details, "\n"),
		descBlock,
	)
}
