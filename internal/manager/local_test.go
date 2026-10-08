package manager

import "testing"

func TestParsePacmanInfoBlock(t *testing.T) {
	block := `Name            : bash
Version         : 5.3.20-1
Description     : The GNU Bourne Again shell
Depends On      : readline  glibc
                  ncurses
Optional Deps   : bash-completion: for tab completion
Install Reason  : Explicitly installed
`
	var p Package
	parsePacmanInfo(block, &p)

	if p.Name != "bash" || p.Version != "5.3.20-1" || p.Description != "The GNU Bourne Again shell" {
		t.Errorf("unexpected header fields: %+v", p)
	}
	if len(p.Depends) != 3 || p.Depends[2] != "ncurses" {
		t.Errorf("unexpected depends: %v", p.Depends)
	}
	if p.InstallReason != "Explicitly installed" || !p.Detailed {
		t.Errorf("unexpected install reason/detailed: %+v", p)
	}
}
