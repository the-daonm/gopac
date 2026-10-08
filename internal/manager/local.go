package manager

import (
	"os"
	"os/exec"
	"sort"
	"strings"
)

// ListInstalled returns every installed package with its local pacman info.
// Foreign packages (not in any sync repo) are reported as AUR packages and
// left undetailed so their AUR info is fetched on demand.
func ListInstalled() ([]Package, error) {
	cmd := exec.Command("pacman", "-Qi")
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	foreign := pacmanNameSet("-Qmq")

	var pkgs []Package
	for block := range strings.SplitSeq(string(out), "\n\n") {
		if strings.TrimSpace(block) == "" {
			continue
		}
		var p Package
		parsePacmanInfo(block, &p)
		if p.Name == "" {
			continue
		}
		p.IsInstalled = true
		if foreign[p.Name] {
			p.IsAUR = true
			p.Detailed = false
		} else {
			p.Maintainer = "Arch Linux"
		}
		pkgs = append(pkgs, p)
	}

	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
	return pkgs, nil
}

// pacmanNameSet runs a pacman query that prints one package name per line.
// pacman exits non-zero when nothing matches, so errors yield an empty set.
func pacmanNameSet(args ...string) map[string]bool {
	set := make(map[string]bool)
	out, _ := exec.Command("pacman", args...).Output()
	for line := range strings.SplitSeq(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			set[line] = true
		}
	}
	return set
}
