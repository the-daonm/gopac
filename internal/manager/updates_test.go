package manager

import (
	"os/exec"
	"testing"
)

func TestParseUpdateLines(t *testing.T) {
	out := "clang 22.1.8-1 -> 23.1.1-1\ndocker 1:29.8.1-1 -> 1:29.8.2-1 [ignored]\n\ngarbage line\n"
	pkgs := parseUpdateLines(out)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 updates, got %d: %+v", len(pkgs), pkgs)
	}
	if p := pkgs[1]; p.Name != "docker" || p.OldVersion != "1:29.8.1-1" || p.Version != "1:29.8.2-1" || !p.IsInstalled {
		t.Errorf("unexpected update: %+v", p)
	}
}

func TestVercmp(t *testing.T) {
	if _, err := exec.LookPath("vercmp"); err != nil {
		t.Skip("vercmp not available")
	}
	if vercmp("1.10-1", "1.9-1") <= 0 {
		t.Errorf("expected 1.10 > 1.9")
	}
	if vercmp("1:1.0-1", "2.0-1") <= 0 {
		t.Errorf("expected epoch to win")
	}
	if vercmp("1.0-1", "1.0-1") != 0 {
		t.Errorf("expected equal versions")
	}
}
