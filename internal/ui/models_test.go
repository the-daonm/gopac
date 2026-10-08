package ui

import (
	"errors"
	"strings"
	"testing"

	"gopac/internal/manager"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestDetailsMsgOnlyUpdatesMatchingSource(t *testing.T) {
	m := NewModel()
	m.allItems = []Item{
		{Pkg: manager.Package{Name: "foo", IsAUR: false}},
		{Pkg: manager.Package{Name: "foo", IsAUR: true, PKGBUILD: "pkgname=foo"}},
	}

	res, _ := m.Update(detailsMsg{pkg: manager.Package{Name: "foo", IsAUR: true, Version: "2.0", Detailed: true}})
	m = res.(Model)

	if m.allItems[0].Pkg.Detailed {
		t.Errorf("official package should not receive AUR details")
	}
	aur := m.allItems[1].Pkg
	if !aur.Detailed || aur.Version != "2.0" {
		t.Errorf("AUR package not updated: %+v", aur)
	}
	if aur.PKGBUILD != "pkgname=foo" {
		t.Errorf("PKGBUILD was clobbered by details fetch")
	}
}

func TestDetailsMsgErrorStopsLoading(t *testing.T) {
	m := NewModel()
	m.allItems = []Item{{Pkg: manager.Package{Name: "foo"}}}
	m.loadingDetailsFor = "foo"

	res, _ := m.Update(detailsMsg{pkg: manager.Package{Name: "foo"}, err: errors.New("boom")})
	m = res.(Model)

	if m.allItems[0].DetailErr != "boom" {
		t.Errorf("expected detail error to be recorded, got %q", m.allItems[0].DetailErr)
	}
	if m.loadingDetailsFor != "" {
		t.Errorf("expected loading state to be cleared")
	}
}

func TestViewFitsNarrowTerminal(t *testing.T) {
	ApplyTheme("")
	m := NewModel()
	res, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = res.(Model)
	m.focusSide = 0
	m.searching = false

	if h := lipgloss.Height(m.View()); h != 24 {
		t.Errorf("view height = %d, want 24", h)
	}
}

func TestTabAtMatchesRenderedHeader(t *testing.T) {
	ApplyTheme("")
	m := NewModel()
	res, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = res.(Model)

	header := ansi.Strip(strings.Split(m.View(), "\n")[0])
	tabsText := " " + strings.Join(tabs, "  ") + " "
	idx := strings.LastIndex(header, tabsText)
	if idx < 0 {
		t.Fatalf("tabs not found in header %q", header)
	}
	col := ansi.StringWidth(header[:idx])
	for i, name := range tabs {
		// Check both the first and last cell of each padded tab.
		for _, x := range []int{col, col + len(name) + 1} {
			if got := m.tabAt(x); got != i {
				t.Errorf("tabAt(%d) = %d, want %d (%s)", x, got, i, name)
			}
		}
		col += len(name) + 2
	}
	if got := m.tabAt(0); got != -1 {
		t.Errorf("tabAt(0) = %d, want -1", got)
	}
}

func TestUpdatesTabKeepsVersionsAfterDetails(t *testing.T) {
	m := NewModel()
	res, _ := m.Update(updatesMsg{pkgs: []manager.Package{
		{Name: "clang", OldVersion: "22.1-1", Version: "23.1-1", IsInstalled: true},
	}})
	m = res.(Model)
	m.activeTab = updatesTab
	m.updatesLoaded = true
	m.updateListItems()

	if n := len(m.list.Items()); n != 1 {
		t.Fatalf("expected 1 update in list, got %d", n)
	}

	// -Qi reports the installed version.
	res, _ = m.Update(detailsMsg{pkg: manager.Package{Name: "clang", Version: "22.1-1", Detailed: true}})
	m = res.(Model)
	if p := m.updates[0].Pkg; p.Version != "23.1-1" || p.OldVersion != "22.1-1" || !p.Detailed {
		t.Errorf("unexpected update item after details: %+v", p)
	}
}

func TestToggleQueueAll(t *testing.T) {
	m := NewModel()
	m.allItems = []Item{
		{Pkg: manager.Package{Name: "a", IsInstalled: true, IsOrphan: true}},
		{Pkg: manager.Package{Name: "b", IsInstalled: true, IsOrphan: true}},
		{Pkg: manager.Package{Name: "c", IsInstalled: true}},
	}
	m.activeTab = 5 // ORPHANS
	m.updateListItems()

	m.toggleQueueAll()
	if len(m.markedRemove) != 2 || len(m.markedInstall) != 0 {
		t.Fatalf("expected both orphans queued for removal, got %v", m.markedRemove)
	}
	m.toggleQueueAll()
	if len(m.markedRemove) != 0 {
		t.Errorf("expected second toggle to clear the queue, got %v", m.markedRemove)
	}
}

func TestClickSelectsListItem(t *testing.T) {
	ApplyTheme("")
	m := NewModel()
	res, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = res.(Model)
	m.allItems = []Item{
		{Pkg: manager.Package{Name: "alpha", Detailed: true}},
		{Pkg: manager.Package{Name: "bravo", Detailed: true}},
		{Pkg: manager.Package{Name: "charlie", Detailed: true}},
	}
	m.updateListItems()

	// Find the row where "charlie" is drawn and click it.
	row := -1
	for y, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		if strings.Contains(line, "charlie") {
			row = y
			break
		}
	}
	if row < 0 {
		t.Fatal("charlie not rendered")
	}
	res, _ = m.Update(tea.MouseMsg{X: 5, Y: row, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m = res.(Model)
	if got := m.list.Index(); got != 2 {
		t.Errorf("selected index = %d, want 2", got)
	}
}

func TestQDoesNotQuit(t *testing.T) {
	m := NewModel()
	m.searching = false
	m.focusSide = 0
	for _, mode := range []string{"list", "help", "confirm"} {
		m.showingHelp = mode == "help"
		m.showingConfirm = mode == "confirm"
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd != nil {
			if _, quit := cmd().(tea.QuitMsg); quit {
				t.Errorf("q quit the app in %s mode", mode)
			}
		}
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if _, quit := cmd().(tea.QuitMsg); !quit {
		t.Errorf("ctrl+c should quit")
	}
}
