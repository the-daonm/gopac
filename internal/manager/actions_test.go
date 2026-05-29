package manager

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDetectAURHelper(t *testing.T) {
	// Save original PATH
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	// Create a temp dir for our "binaries"
	tmpDir := t.TempDir()

	// Override PATH to only look in tmpDir + basic system paths if needed,
	// but strictly tmpDir ensures we don't accidentally find real tools.
	// However, exec.LookPath might behave differently on some OSs if PATH is just one dir.
	// On Linux it's fine.
	os.Setenv("PATH", tmpDir)

	// Helper to reset state
	reset := func() {
		aurHelper = ""
	}

	// 1. Test with no helpers (should fallback to pacman)
	reset()
	if h := detectAURHelper(); h != "pacman" {
		t.Errorf("Scenario 1 (No helpers): Expected fallback 'pacman', got '%s'", h)
	}

	// 2. Test with 'yay' existing
	yayPath := filepath.Join(tmpDir, "yay")
	if err := os.WriteFile(yayPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	reset()
	if h := detectAURHelper(); h != "yay" {
		t.Errorf("Scenario 2 (yay exists): Expected 'yay', got '%s'", h)
	}

	// 3. Test priority: 'paru' should be preferred over 'yay' if both exist
	paruPath := filepath.Join(tmpDir, "paru")
	if err := os.WriteFile(paruPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	reset()
	if h := detectAURHelper(); h != "paru" {
		t.Errorf("Scenario 3 (paru & yay exist): Expected 'paru', got '%s'", h)
	}

	// 4. Test with ENV var overriding everything
	// Create custom helper
	customHelper := "my-helper"
	customPath := filepath.Join(tmpDir, customHelper)
	if err := os.WriteFile(customPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	os.Setenv("AUR_HELPER", customHelper)
	reset()
	if h := detectAURHelper(); h != customHelper {
		t.Errorf("Scenario 4 (ENV var set): Expected '%s', got '%s'", customHelper, h)
	}
	os.Unsetenv("AUR_HELPER")
}

func TestInstallOrRemove(t *testing.T) {
	// 1. Test AUR install with 'aura'
	SetAURHelper("aura")
	cmd := InstallOrRemove("some-package", true, false)
	// Expect -A
	foundA := slices.Contains(cmd.Args, "-A")
	if !foundA {
		t.Errorf("Scenario 1 (aura): Expected '-A' flag, got %v", cmd.Args)
	}

	// 2. Test AUR install with 'yay'
	SetAURHelper("yay")
	cmd = InstallOrRemove("some-package", true, false)
	// Expect -S
	foundS := slices.Contains(cmd.Args, "-S")
	if !foundS {
		t.Errorf("Scenario 2 (yay): Expected '-S' flag, got %v", cmd.Args)
	}

	// 3. Test Official install (should ignore helper)
	SetAURHelper("aura")                                // Even if helper is aura
	cmd = InstallOrRemove("some-package", false, false) // isAUR = false
	// Expect pacman -S
	foundPacman := slices.Contains(cmd.Args, "pacman")
	foundS = slices.Contains(cmd.Args, "-S")
	if !foundPacman || !foundS {
		t.Errorf("Scenario 3 (official): Expected 'pacman' and '-S', got %v", cmd.Args)
	}

	// Reset helper
	SetAURHelper("")
}

func TestBulkActionCmd(t *testing.T) {
	SetAURHelper("yay")
	defer SetAURHelper("")

	cmd := BulkActionCmd([]string{"git", "curl"}, []string{"yay-git"}, []string{"vim", "nano"})
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	expectedArgs := []string{"sh", "-c", "sudo pacman -Rns -- vim nano && sudo pacman -S -- git curl && yay -S -- yay-git"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Fatalf("Expected %d arguments, got %d. Args: %v", len(expectedArgs), len(cmd.Args), cmd.Args)
	}

	for i, arg := range cmd.Args {
		if arg != expectedArgs[i] {
			t.Errorf("Expected arg %d to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}
