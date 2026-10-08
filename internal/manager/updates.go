package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// aurInfoBatch keeps AUR info requests well below the RPC URL length limit.
const aurInfoBatch = 100

// ListUpdates returns installed packages that have a newer version available,
// with Version set to the new version and OldVersion to the installed one.
// Repo updates are always returned; an AUR failure is reported as the error.
func ListUpdates(ctx context.Context) ([]Package, error) {
	pkgs, err := repoUpdates(ctx)
	if err != nil {
		return nil, err
	}

	aurPkgs, aurErr := aurUpdates(ctx)
	pkgs = append(pkgs, aurPkgs...)

	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
	return pkgs, aurErr
}

// repoUpdates prefers checkupdates (pacman-contrib), which syncs a temporary
// database copy without root. pacman -Qu only sees the last `pacman -Sy`.
func repoUpdates(ctx context.Context) ([]Package, error) {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("checkupdates"); err == nil {
		cmd = exec.CommandContext(ctx, "checkupdates")
	} else {
		cmd = exec.CommandContext(ctx, "pacman", "-Qu")
	}
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		// Both tools exit non-zero (1 or 2) when there is nothing to update.
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("checking repo updates: %w", err)
	}
	return parseUpdateLines(string(out)), nil
}

// parseUpdateLines parses "name old -> new" lines as printed by pacman -Qu
// and checkupdates.
func parseUpdateLines(out string) []Package {
	var pkgs []Package
	for line := range strings.SplitSeq(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 4 || f[2] != "->" {
			continue
		}
		pkgs = append(pkgs, Package{
			Name:        f[0],
			OldVersion:  f[1],
			Version:     f[3],
			IsInstalled: true,
			Maintainer:  "Arch Linux",
		})
	}
	return pkgs
}

func aurUpdates(ctx context.Context) ([]Package, error) {
	out, err := exec.CommandContext(ctx, "pacman", "-Qm").Output()
	if err != nil {
		return nil, nil // no foreign packages
	}
	installed := make(map[string]string)
	var names []string
	for line := range strings.SplitSeq(string(out), "\n") {
		if f := strings.Fields(line); len(f) == 2 {
			installed[f[0]] = f[1]
			names = append(names, f[0])
		}
	}

	var pkgs []Package
	for start := 0; start < len(names); start += aurInfoBatch {
		batch := names[start:min(start+aurInfoBatch, len(names))]
		infos, err := aurInfo(ctx, batch)
		if err != nil {
			return pkgs, err
		}
		for _, info := range infos {
			old := installed[info.Name]
			if vercmp(info.Version, old) > 0 {
				pkgs = append(pkgs, Package{
					Name:        info.Name,
					OldVersion:  old,
					Version:     info.Version,
					Description: info.Description,
					IsAUR:       true,
					IsInstalled: true,
					Maintainer:  info.Maintainer,
				})
			}
		}
	}
	return pkgs, nil
}

type aurInfoResult struct {
	Name        string `json:"Name"`
	Version     string `json:"Version"`
	Description string `json:"Description"`
	Maintainer  string `json:"Maintainer"`
}

func aurInfo(ctx context.Context, names []string) ([]aurInfoResult, error) {
	q := url.Values{"v": {"5"}, "type": {"info"}, "arg[]": names}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://aur.archlinux.org/rpc/?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AUR request failed: %s", resp.Status)
	}

	var data struct {
		Error   string          `json:"error"`
		Results []aurInfoResult `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.Error != "" {
		return nil, fmt.Errorf("AUR: %s", data.Error)
	}
	return data.Results, nil
}

// vercmp compares two pacman versions using pacman's own vercmp tool and
// returns <0, 0 or >0. Unknown results are treated as equal.
func vercmp(a, b string) int {
	out, err := exec.Command("vercmp", a, b).Output()
	if err != nil {
		return 0
	}
	var n int
	fmt.Sscan(strings.TrimSpace(string(out)), &n)
	return n
}
