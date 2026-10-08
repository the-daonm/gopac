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
