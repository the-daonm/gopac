// Package history persists search history between sessions.
package history

import (
	"os"
	"path/filepath"
	"strings"
)

// MaxEntries is the number of most recent searches that are kept.
const MaxEntries = 100

func path() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gopac", "history"), nil
}

// Load returns saved searches, oldest first. A missing file is not an error.
func Load() []string {
	p, err := path()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var entries []string
	for line := range strings.SplitSeq(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			entries = append(entries, line)
		}
	}
	return entries
}

// Save writes the most recent MaxEntries searches.
func Save(entries []string) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	if len(entries) > MaxEntries {
		entries = entries[len(entries)-MaxEntries:]
	}
	return os.WriteFile(p, []byte(strings.Join(entries, "\n")+"\n"), 0o644)
}
