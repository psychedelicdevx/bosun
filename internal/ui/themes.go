package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Accent  lipgloss.Color
	Running lipgloss.Color
	Stopped lipgloss.Color
	Errored lipgloss.Color
	Muted   lipgloss.Color
	Label   lipgloss.Color
	SelBg   lipgloss.Color
	SelFg   lipgloss.Color
}

var ThemeNames = []string{"default", "catppuccin", "dracula", "tokyonight", "nord", "gruvbox", "rosepine", "solarized", "monokai", "mono"}

var CurrentTheme = "default"

func themeIndex(name string) int {
	for i, n := range ThemeNames {
		if n == name {
			return i
		}
	}
	return 0
}

var Themes = map[string]Theme{
	"default":    {Accent: "42", Running: "42", Stopped: "240", Errored: "196", Muted: "240", Label: "245", SelBg: "237", SelFg: "252"},
	"catppuccin": {Accent: "183", Running: "151", Stopped: "240", Errored: "211", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
	"dracula":    {Accent: "141", Running: "84", Stopped: "240", Errored: "212", Muted: "240", Label: "245", SelBg: "238", SelFg: "253"},
	"tokyonight": {Accent: "111", Running: "149", Stopped: "240", Errored: "210", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
	"nord":       {Accent: "109", Running: "108", Stopped: "240", Errored: "167", Muted: "240", Label: "245", SelBg: "238", SelFg: "253"},
	"gruvbox":    {Accent: "214", Running: "142", Stopped: "245", Errored: "167", Muted: "243", Label: "246", SelBg: "237", SelFg: "223"},
	"rosepine":   {Accent: "116", Running: "108", Stopped: "240", Errored: "174", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
	"solarized":  {Accent: "37", Running: "100", Stopped: "240", Errored: "160", Muted: "241", Label: "244", SelBg: "236", SelFg: "254"},
	"monokai":    {Accent: "197", Running: "148", Stopped: "240", Errored: "208", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
	"mono":       {Accent: "250", Running: "250", Stopped: "241", Errored: "244", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
}

var (
	runningStyle      lipgloss.Style
	stoppedStyle      lipgloss.Style
	erroredStyle      lipgloss.Style
	headerStyle       lipgloss.Style
	hintStyle         lipgloss.Style
	labelStyle        lipgloss.Style
	selRowStyle       lipgloss.Style
	borderActiveStyle lipgloss.Style

	borderDim    lipgloss.Color
	borderActive lipgloss.Color
	labelDim     lipgloss.Color
)

func ApplyTheme(name string) {
	t, ok := Themes[name]
	if !ok {
		name = "default"
		t = Themes["default"]
	}
	CurrentTheme = name
	runningStyle = lipgloss.NewStyle().Foreground(t.Running)
	stoppedStyle = lipgloss.NewStyle().Foreground(t.Stopped)
	erroredStyle = lipgloss.NewStyle().Foreground(t.Errored)
	headerStyle = lipgloss.NewStyle().Bold(true)
	hintStyle = lipgloss.NewStyle().Foreground(t.Muted)
	labelStyle = lipgloss.NewStyle().Foreground(t.Label)
	selRowStyle = lipgloss.NewStyle().Bold(true).Foreground(t.SelFg).Background(t.SelBg)
	borderActiveStyle = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)

	borderDim = t.Muted
	borderActive = t.Accent
	labelDim = t.Label
}

func init() {
	ApplyTheme("default")
}
