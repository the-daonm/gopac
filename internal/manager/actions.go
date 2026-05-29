package manager

import (
	"os"
	"os/exec"
	"strings"
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

func UpdateSystem() *exec.Cmd {
	var cmd *exec.Cmd
	helper := detectAURHelper()

	if helper != "" && helper != "pacman" {
		cmd = exec.Command(helper, "-Syu")
	} else {
		cmd = exec.Command("sudo", "pacman", "-Syu")
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
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

func BulkActionCmd(toInstallOfficial []string, toInstallAUR []string, toRemove []string) *exec.Cmd {
	var commands []string

	if len(toRemove) > 0 {
		commands = append(commands, "sudo pacman -Rns "+strings.Join(toRemove, " "))
	}

	if len(toInstallOfficial) > 0 {
		commands = append(commands, "sudo pacman -S "+strings.Join(toInstallOfficial, " "))
	}

	if len(toInstallAUR) > 0 {
		helper := detectAURHelper()
		flag := "-S"
		if helper == "aura" {
			flag = "-A"
		}
		commands = append(commands, helper+" "+flag+" "+strings.Join(toInstallAUR, " "))
	}

	if len(commands) == 0 {
		return nil
	}

	// Join commands with " && " and run via sh -c
	fullCmd := strings.Join(commands, " && ")
	cmd := exec.Command("sh", "-c", fullCmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

