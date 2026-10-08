package manager

import "testing"

func TestCheckInstalledStatusClearsStaleFlag(t *testing.T) {
	cacheMu.Lock()
	orig := installedCache
	installedCache = map[string]bool{"vim": true}
	cacheMu.Unlock()
	defer func() {
		cacheMu.Lock()
		installedCache = orig
		cacheMu.Unlock()
	}()

	pkgs := []Package{
		{Name: "vim"},
		{Name: "nano", IsInstalled: true}, // removed since it was cached
	}
	checkInstalledStatus(pkgs)

	if !pkgs[0].IsInstalled {
		t.Errorf("expected vim to be installed")
	}
	if pkgs[1].IsInstalled {
		t.Errorf("expected nano to no longer be installed")
	}
}

func TestParsePacmanOutput(t *testing.T) {
	raw := `extra/gtk3 1:3.24.43-4 [installed]
    GObject-based multi-platform GUI toolkit
extra/gtkmm3 3.24.9-1
    C++ bindings for GTK 3
core/glib2 2.82.4-2
    Low level core library
`
	pkgs := parsePacmanOutput(raw, "GTK")
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d: %+v", len(pkgs), pkgs)
	}
	if pkgs[0].Name != "gtk3" || pkgs[0].Version != "1:3.24.43-4" || pkgs[0].Description != "GObject-based multi-platform GUI toolkit" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[1].Name != "gtkmm3" || pkgs[1].Description != "C++ bindings for GTK 3" {
		t.Errorf("unexpected second package: %+v", pkgs[1])
	}
}
