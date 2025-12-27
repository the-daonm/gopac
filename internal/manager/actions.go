package manager

import (
	"os"
	"os/exec"
)

var aurHelper string

func SetAURHelper(name string) {
	aurHelper = name
}

func detectAURHelper() string {
	if aurHelper != "" {
		return aurHelper
	}

	if env := os.Getenv("AUR_HELPER"); env != "" {
		if _, err := exec.LookPath(env); err == nil {
			aurHelper = env
			return env
		}
	}

	helpers := []string{"paru", "yay", "pikaur", "aura", "trizen"}
	for _, h := range helpers {
		if _, err := exec.LookPath(h); err == nil {
			aurHelper = h
			return h
		}
	}

	return "paru"
}

func InstallOrRemove(pkgName string, isAUR bool, remove bool) *exec.Cmd {
	var cmd *exec.Cmd

	if remove {
		cmd = exec.Command("sudo", "pacman", "-Rns", pkgName)
	} else {
		if isAUR {
			helper := detectAURHelper()
			args := []string{"-S", pkgName}
			if helper == "aura" {
				args = []string{"-A", pkgName}
			}
			cmd = exec.Command(helper, args...)
		} else {
			cmd = exec.Command("sudo", "pacman", "-S", pkgName)
		}
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
