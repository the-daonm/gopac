package ui

import "github.com/charmbracelet/lipgloss"

type ThemeColors struct {
	Base      lipgloss.Color
	Text      lipgloss.Color
	Red       lipgloss.Color
	Green     lipgloss.Color
	Yellow    lipgloss.Color
	Orange    lipgloss.Color
	Blue      lipgloss.Color
	Purple    lipgloss.Color
	Cyan      lipgloss.Color
	Gray      lipgloss.Color
	Highlight lipgloss.Color
	Border    lipgloss.Color

	// Semantic Roles
	RepoOfficial lipgloss.Color
	RepoAUR      lipgloss.Color
	Header       lipgloss.Color
	Focus        lipgloss.Color
}

var (
	CurrentTheme ThemeColors

	// Global Styles
	AppStyle       lipgloss.Style
	ContainerStyle lipgloss.Style
	ListStyle      lipgloss.Style
	DescStyle      lipgloss.Style
	HeaderStyle    lipgloss.Style
	InputStyle     lipgloss.Style
	LabelStyle     lipgloss.Style
	ValueStyle     lipgloss.Style
	LinkStyle      lipgloss.Style
	FocusedStyle   lipgloss.Style
	BlurredStyle   lipgloss.Style

	// Pre-defined Themes
	Themes = map[string]ThemeColors{
		"gruvbox": {
			Base:         lipgloss.Color("#282828"),
			Text:         lipgloss.Color("#ebdbb2"),
			Red:          lipgloss.Color("#cc241d"),
			Green:        lipgloss.Color("#98971a"),
			Yellow:       lipgloss.Color("#d79921"),
			Orange:       lipgloss.Color("#d65d0e"),
			Blue:         lipgloss.Color("#458588"),
			Purple:       lipgloss.Color("#b16286"),
			Cyan:         lipgloss.Color("#689d6a"),
			Gray:         lipgloss.Color("#928374"),
			Highlight:    lipgloss.Color("#504945"),
			Border:       lipgloss.Color("#7c6f64"),
			RepoOfficial: lipgloss.Color("#98971a"), // Green
			RepoAUR:      lipgloss.Color("#d65d0e"), // Orange
			Header:       lipgloss.Color("#98971a"), // Green
			Focus:        lipgloss.Color("#d79921"), // Yellow
		},
		"onedark": {
			Base:         lipgloss.Color("#282c34"),
			Text:         lipgloss.Color("#abb2bf"),
			Red:          lipgloss.Color("#e06c75"),
			Green:        lipgloss.Color("#98c379"),
			Yellow:       lipgloss.Color("#e5c07b"),
			Orange:       lipgloss.Color("#d19a66"),
			Blue:         lipgloss.Color("#61afef"),
			Purple:       lipgloss.Color("#c678dd"),
			Cyan:         lipgloss.Color("#56b6c2"),
			Gray:         lipgloss.Color("#5c6370"),
			Highlight:    lipgloss.Color("#3e4451"),
			Border:       lipgloss.Color("#4b5263"),
			RepoOfficial: lipgloss.Color("#61afef"), // Blue
			RepoAUR:      lipgloss.Color("#c678dd"), // Purple
			Header:       lipgloss.Color("#61afef"), // Blue
			Focus:        lipgloss.Color("#98c379"), // Green
		},
		"dracula": {
			Base:         lipgloss.Color("#282a36"),
			Text:         lipgloss.Color("#f8f8f2"),
			Red:          lipgloss.Color("#ff5555"),
			Green:        lipgloss.Color("#50fa7b"),
			Yellow:       lipgloss.Color("#f1fa8c"),
			Orange:       lipgloss.Color("#ffb86c"),
			Blue:         lipgloss.Color("#8be9fd"),
			Purple:       lipgloss.Color("#bd93f9"),
			Cyan:         lipgloss.Color("#8be9fd"),
			Gray:         lipgloss.Color("#6272a4"),
			Highlight:    lipgloss.Color("#44475a"),
			Border:       lipgloss.Color("#6272a4"),
			RepoOfficial: lipgloss.Color("#bd93f9"), // Purple
			RepoAUR:      lipgloss.Color("#ff79c6"), // Pink (Better contrast than Orange)
			Header:       lipgloss.Color("#bd93f9"), // Purple
			Focus:        lipgloss.Color("#50fa7b"), // Green
		},
		"nord": {
			Base:         lipgloss.Color("#2e3440"),
			Text:         lipgloss.Color("#d8dee9"),
			Red:          lipgloss.Color("#bf616a"),
			Green:        lipgloss.Color("#a3be8c"),
			Yellow:       lipgloss.Color("#ebcb8b"),
			Orange:       lipgloss.Color("#d08770"),
			Blue:         lipgloss.Color("#81a1c1"),
			Purple:       lipgloss.Color("#b48ead"),
			Cyan:         lipgloss.Color("#88c0d0"),
			Gray:         lipgloss.Color("#4c566a"),
			Highlight:    lipgloss.Color("#3b4252"),
			Border:       lipgloss.Color("#434c5e"),
			RepoOfficial: lipgloss.Color("#81a1c1"), // Frost Blue
			RepoAUR:      lipgloss.Color("#d08770"), // Aurora Orange
			Header:       lipgloss.Color("#5e81ac"), // Frost Dark Blue
			Focus:        lipgloss.Color("#88c0d0"), // Frost Cyan
		},
		"catppuccin": {
			Base:         lipgloss.Color("#1e1e2e"),
			Text:         lipgloss.Color("#cdd6f4"),
			Red:          lipgloss.Color("#f38ba8"),
			Green:        lipgloss.Color("#a6e3a1"),
			Yellow:       lipgloss.Color("#f9e2af"),
			Orange:       lipgloss.Color("#fab387"),
			Blue:         lipgloss.Color("#89b4fa"),
			Purple:       lipgloss.Color("#cba6f7"),
			Cyan:         lipgloss.Color("#94e2d5"),
			Gray:         lipgloss.Color("#6c7086"),
			Highlight:    lipgloss.Color("#313244"),
			Border:       lipgloss.Color("#45475a"),
			RepoOfficial: lipgloss.Color("#89b4fa"), // Blue
			RepoAUR:      lipgloss.Color("#cba6f7"), // Mauve (Purple)
			Header:       lipgloss.Color("#b4befe"), // Lavender
			Focus:        lipgloss.Color("#f5c2e7"), // Pink
		},
	}
)

func ApplyTheme(themeName string) {
	t, ok := Themes[themeName]
	if !ok {
		t = Themes["gruvbox"] // Default
	}
	CurrentTheme = t

	// Update Styles
	AppStyle = lipgloss.NewStyle().Margin(0)

	ContainerStyle = lipgloss.NewStyle().
		Padding(0)

	ListStyle = lipgloss.NewStyle().Padding(0, 1)
	DescStyle = lipgloss.NewStyle().Padding(0, 2)

	HeaderStyle = lipgloss.NewStyle().
		Foreground(t.Base).
		Background(t.Header).
		Bold(true).
		Padding(0, 2)

	InputStyle = lipgloss.NewStyle().
		Foreground(t.Focus).
		Padding(0, 2)

	LabelStyle = lipgloss.NewStyle().Foreground(t.Gray).Width(12).Bold(true)
	ValueStyle = lipgloss.NewStyle().Foreground(t.Text)
	LinkStyle = lipgloss.NewStyle().Foreground(t.Blue).Underline(true)

	FocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Focus).
		Padding(0, 2)

	BlurredStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(0, 2)
}

func GetRepoColor(isAUR bool) lipgloss.Color {
	if isAUR {
		return CurrentTheme.RepoAUR
	}
	return CurrentTheme.RepoOfficial
}
