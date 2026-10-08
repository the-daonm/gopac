package ui

import (
	"errors"
	"testing"

	"gopac/internal/manager"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
