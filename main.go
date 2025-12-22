package main

import (
	"flag"
	"fmt"
	"os"
	"gopac/internal/manager"
	"gopac/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	helper := flag.String("helper", "", "Specify AUR helper to use (e.g. paru, yay)")
	flag.StringVar(helper, "H", "", "Specify AUR helper to use (e.g. paru, yay)")
	flag.Parse()

	if *helper != "" {
		manager.SetAURHelper(*helper)
	}

	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
