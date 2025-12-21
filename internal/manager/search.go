package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"sync"
)

type Package struct {
	Name        string
	Version     string
	Description string
	IsAUR       bool
	IsInstalled bool
	Votes       int

	URL          string
	Maintainer   string
	LastModified int64
}

func Search(query string) ([]Package, error) {
	var results []Package
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Search Local Repos (pacman)
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd := exec.Command("pacman", "-Ss", query)
		out, _ := cmd.Output()
		localPkgs := parsePacmanOutput(string(out), query)

		mu.Lock()
		results = append(results, localPkgs...)
		mu.Unlock()
	}()

	// Search AUR (RPC API)
	wg.Add(1)
	go func() {
		defer wg.Done()
		aurPkgs, err := searchAUR(query)
		if err == nil {
			mu.Lock()
			results = append(results, aurPkgs...)
			mu.Unlock()
		}
	}()

	wg.Wait()

	if len(results) > 0 {
		checkInstalledStatus(results)
	}

	sortPackages(results, query)
	return results, nil
}

func sortPackages(pkgs []Package, query string) {
	query = strings.ToLower(query)
	sort.SliceStable(pkgs, func(i, j int) bool {
		p1 := pkgs[i]
		p2 := pkgs[j]
		if p1.Name == query && p2.Name != query {
			return true
		}
		if p2.Name == query && p1.Name != query {
			return false
		}
		if !p1.IsAUR && p2.IsAUR {
			return true
		}
		if p1.IsAUR && !p2.IsAUR {
			return false
		}
		if p1.IsAUR && p2.IsAUR {
			if p1.Votes != p2.Votes {
				return p1.Votes > p2.Votes
			}
		}
		if len(p1.Name) != len(p2.Name) {
			return len(p1.Name) < len(p2.Name)
		}
		return p1.Name < p2.Name
	})
}

func searchAUR(query string) ([]Package, error) {
	url := fmt.Sprintf("https://aur.archlinux.org/rpc/?v=5&type=search&by=name&arg=%s", query)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type aurResult struct {
		Name         string `json:"Name"`
		Version      string `json:"Version"`
		Description  string `json:"Description"`
		NumVotes     int    `json:"NumVotes"`
		URL          string `json:"URL"`
		Maintainer   string `json:"Maintainer"`
		LastModified int64  `json:"LastModified"`
	}
	type response struct {
		Results []aurResult `json:"results"`
	}

	var data response
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var pkgs []Package
	for _, r := range data.Results {
		pkgs = append(pkgs, Package{
			Name:         r.Name,
			Version:      r.Version,
			Description:  r.Description,
			IsAUR:        true,
			Votes:        r.NumVotes,
			URL:          r.URL,
			Maintainer:   r.Maintainer,
			LastModified: r.LastModified,
		})
	}
	return pkgs, nil
}

func parsePacmanOutput(raw string, query string) []Package {
	var pkgs []Package
	lines := strings.Split(raw, "\n")
	lowQuery := strings.ToLower(query)

	for i := 0; i < len(lines)-1; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.Contains(line, "/") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				nameSplit := strings.Split(parts[0], "/")
				if len(nameSplit) < 2 {
					continue
				}

				name := nameSplit[1]
				if !strings.Contains(strings.ToLower(name), lowQuery) {
					if i+1 < len(lines) {
						i++
					}
					continue
				}

				ver := parts[1]
				desc := ""
				if i+1 < len(lines) {
					desc = strings.TrimSpace(lines[i+1])
					i++
				}

				pkgs = append(pkgs, Package{
					Name:        name,
					Version:     ver,
					Description: desc,
					IsAUR:       false,
					Maintainer:  "Arch Linux",
				})
			}
		}
	}
	return pkgs
}

func checkInstalledStatus(pkgs []Package) {
	out, err := exec.Command("pacman", "-Qq").Output()
	if err != nil {
		return
	}
	installedMap := make(map[string]bool)
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" {
			installedMap[line] = true
		}
	}
	for i := range pkgs {
		if installedMap[pkgs[i].Name] {
			pkgs[i].IsInstalled = true
		}
	}
}
