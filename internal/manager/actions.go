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

	return "pacman"
}

// HasAURHelper reports whether an AUR helper is configured or installed.
func HasAURHelper() bool {
	return detectAURHelper() != "pacman"
}

func UpdateSystem() *exec.Cmd {
	helper := detectAURHelper()

	if helper != "" && helper != "pacman" {
		return interactiveShell(shellQuote(helper) + " -Syu")
	}
	return interactiveShell("sudo pacman -Syu")
}

func InstallOrRemove(pkgName string, isAUR bool, remove bool) *exec.Cmd {
	var cmd *exec.Cmd

	if remove {
		cmd = exec.Command("sudo", "pacman", "-Rns", "--", pkgName)
	} else {
		if isAUR {
			helper := detectAURHelper()
			args := []string{"-S", "--", pkgName}
			if helper == "aura" {
				args = []string{"-A", "--", pkgName}
			}
			cmd = exec.Command(helper, args...)
		} else {
			cmd = exec.Command("sudo", "pacman", "-S", "--", pkgName)
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
		commands = append(commands, "sudo pacman -Rns -- "+shellJoin(toRemove))
	}

	if len(toInstallOfficial) > 0 {
		commands = append(commands, "sudo pacman -S -- "+shellJoin(toInstallOfficial))
	}

	if len(toInstallAUR) > 0 {
		helper := detectAURHelper()
		flag := "-S"
		if helper == "aura" {
			flag = "-A"
		}
		commands = append(commands, shellQuote(helper)+" "+flag+" -- "+shellJoin(toInstallAUR))
	}

	if len(commands) == 0 {
		return nil
	}

	return interactiveShell(strings.Join(commands, " && "))
}

// pauseOnFailure keeps the output of a failed command on screen until the user
// presses Enter; otherwise the TUI redraws immediately and hides the error.
const pauseOnFailure = ` || { s=$?; printf '\n\033[1;31mgopac: command failed (exit %d). Press Enter to return...\033[0m' "$s"; read -r _; exit "$s"; }`

func interactiveShell(script string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", script+pauseOnFailure)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// shellQuote quotes s for safe use as a single word in a POSIX shell command.
// Plain words are left as-is to keep the command readable.
func shellQuote(s string) string {
	if s != "" && strings.IndexFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("@%+=:,./_-", r))
	}) == -1 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func shellJoin(words []string) string {
	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = shellQuote(w)
	}
	return strings.Join(quoted, " ")
}
