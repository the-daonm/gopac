package manager

import (
	"os"
	"os/exec"
)

func InstallOrRemove(pkgName string, isAUR bool, remove bool) *exec.Cmd {
	var cmd *exec.Cmd

	if remove {
		cmd = exec.Command("sudo", "pacman", "-Rns", pkgName)
	} else {
		if isAUR {
			cmd = exec.Command("paru", "-S", pkgName)
		} else {
			cmd = exec.Command("sudo", "pacman", "-S", pkgName)
		}
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
