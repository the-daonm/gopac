package main

import (
	"flag"
	"fmt"
	"gopac/internal/config"
	"gopac/internal/manager"
	"gopac/internal/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = "dev"
)

func main() {
	var (
		helperStr string
		themeStr  string
		showVer   bool
	)

	// Define flags
	flag.StringVar(&helperStr, "helper", "", "Specify AUR helper to use (e.g. paru, yay)")
	flag.StringVar(&helperStr, "H", "", "Specify AUR helper (shorthand)")

	flag.StringVar(&themeStr, "theme", "", "Specify UI theme (gruvbox, onedark, dracula, nord, catppuccin)")
	flag.StringVar(&themeStr, "t", "", "Specify UI theme (shorthand)")

	flag.BoolVar(&showVer, "version", false, "Show version information")
	flag.BoolVar(&showVer, "v", false, "Show version information (shorthand)")

	// Custom Usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "A warm, beautiful TUI for Arch Linux package management.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.VisitAll(func(f *flag.Flag) {
			if len(f.Name) > 1 {
				shorthand := ""
				switch f.Name {
				case "helper":
					shorthand = "-H, "
				case "theme":
					shorthand = "-t, "
				case "version":
					shorthand = "-v, "
				}
				fmt.Fprintf(os.Stderr, "  %s--%-10s %s\n", shorthand, f.Name, f.Usage)
			}
		})
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintln(os.Stderr, "  gopac")
		fmt.Fprintln(os.Stderr, "  gopac -t dracula")
		fmt.Fprintln(os.Stderr, "  gopac --helper yay")
	}

	flag.Parse()

	if showVer {
		fmt.Printf("gopac %s\n", version)
		os.Exit(0)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		// Just warn, don't exit. Config might not exist.
	}

	// Determine Theme
	// Flag > Config > Default
	selectedTheme := ""
	if cfg != nil {
		selectedTheme = cfg.Theme
	}
	if themeStr != "" {
		selectedTheme = themeStr
	}
	ui.ApplyTheme(selectedTheme)

	// Determine Helper
	// Flag > Config > Auto-detect (handled in manager)
	if helperStr != "" {
		manager.SetAURHelper(helperStr)
	} else if cfg != nil && cfg.AURHelper != "" {
		manager.SetAURHelper(cfg.AURHelper)
	}

	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
