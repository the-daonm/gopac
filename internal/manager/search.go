package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
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

	Detailed       bool
	Architecture   string
	Licenses       []string
	Groups         []string
	Provides       []string
	Depends        []string
	OptDepends     []string
	RequiredBy     []string
	Conflicts      []string
	Replaces       []string
	Packager       string
	BuildDate      int64
	InstallDate    int64
	InstallReason  string
	ValidatedBy    string
	DownloadSize   string
	InstalledSize  string
	Popularity     float64
	FirstSubmitted int64
	Keywords       []string
	MakeDepends    []string
	CheckDepends   []string
	PKGBUILD       string
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
		cmd.Env = append(cmd.Env, "LC_ALL=C")
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

func GetPackageDetails(p *Package) error {
	if p.IsAUR {
		return getAURDetails(p)
	}

	flag := "-Si"
	if p.IsInstalled {
		flag = "-Qi"
	}
	return getPacmanDetails(p, flag)
}

func GetPKGBUILD(pkgName string) (string, error) {
	url := fmt.Sprintf("https://aur.archlinux.org/cgit/aur.git/plain/PKGBUILD?h=%s", pkgName)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch PKGBUILD: %s", resp.Status)
	}

	var sb strings.Builder
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	return sb.String(), nil
}

func getAURDetails(p *Package) error {
	url := fmt.Sprintf("https://aur.archlinux.org/rpc/?v=5&type=info&arg[]=%s", p.Name)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type aurInfo struct {
		Name           string   `json:"Name"`
		Keywords       []string `json:"Keywords"`
		License        []string `json:"License"`
		Depends        []string `json:"Depends"`
		MakeDepends    []string `json:"MakeDepends"`
		OptDepends     []string `json:"OptDepends"`
		CheckDepends   []string `json:"CheckDepends"`
		Conflicts      []string `json:"Conflicts"`
		Provides       []string `json:"Provides"`
		Replaces       []string `json:"Replaces"`
		Groups         []string `json:"Groups"`
		Popularity     float64  `json:"Popularity"`
		FirstSubmitted int64    `json:"FirstSubmitted"`
		LastModified   int64    `json:"LastModified"`
		Maintainer     string   `json:"Maintainer"`
		URL            string   `json:"URL"`
		Description    string   `json:"Description"`
		Version        string   `json:"Version"`
	}
	type response struct {
		Results []aurInfo `json:"results"`
	}

	var data response
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	if len(data.Results) == 0 {
		return fmt.Errorf("no info found for %s", p.Name)
	}

	info := data.Results[0]
	p.Keywords = info.Keywords
	p.Licenses = info.License
	p.Depends = info.Depends
	p.MakeDepends = info.MakeDepends
	p.CheckDepends = info.CheckDepends
	p.OptDepends = info.OptDepends
	p.Conflicts = info.Conflicts
	p.Provides = info.Provides
	p.Replaces = info.Replaces
	p.Groups = info.Groups
	p.Popularity = info.Popularity
	p.FirstSubmitted = info.FirstSubmitted
	p.LastModified = info.LastModified
	p.Maintainer = info.Maintainer
	p.URL = info.URL
	p.Description = info.Description
	p.Version = info.Version

	p.Detailed = true
	return nil
}

func getPacmanDetails(p *Package, flag string) error {
	cmd := exec.Command("pacman", flag, p.Name)
	cmd.Env = append(cmd.Env, "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if val == "None" || val == "" {
			continue
		}

		switch key {
		case "Architecture":
			p.Architecture = val
		case "URL":
			p.URL = val
		case "Licenses":
			p.Licenses = strings.Fields(val)
		case "Groups":
			p.Groups = strings.Fields(val)
		case "Provides":
			p.Provides = strings.Fields(val)
		case "Depends On":
			p.Depends = strings.Fields(val)
		case "Optional Deps":
			p.OptDepends = append(p.OptDepends, val)
		case "Required By":
			p.RequiredBy = strings.Fields(val)
		case "Conflicts With":
			p.Conflicts = strings.Fields(val)
		case "Replaces":
			p.Replaces = strings.Fields(val)
		case "Download Size":
			p.DownloadSize = val
		case "Installed Size":
			p.InstalledSize = val
		case "Packager":
			p.Packager = val
		case "Build Date":
			formats := []string{
				"Tue 10 May 2005 11:30:05 PM +07",
				"Tue 10 May 2005 11:30:05 PM MST",
				"Tue May 10 11:30:05 2005",
				time.RFC1123,
			}
			for _, f := range formats {
				if t, err := time.Parse(f, val); err == nil {
					p.BuildDate = t.Unix()
					break
				}
			}
		case "Install Date":
			formats := []string{
				"Tue 10 May 2005 11:30:05 PM +07",
				"Tue 10 May 2005 11:30:05 PM MST",
				"Tue May 10 11:30:05 2005",
				time.RFC1123,
			}
			for _, f := range formats {
				if t, err := time.Parse(f, val); err == nil {
					p.InstallDate = t.Unix()
					break
				}
			}
		case "Install Reason":
			p.InstallReason = val
		case "Validated By":
			p.ValidatedBy = val
		}
	}
	p.Detailed = true
	return nil
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
