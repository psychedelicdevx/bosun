package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Accent  lipgloss.Color
	Key     lipgloss.Color
	Group   lipgloss.Color
	Running lipgloss.Color
	Warning lipgloss.Color
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
	"default":    {Accent: "#A6E3A1", Key: "#89B4FA", Group: "#F9E2AF", Running: "#A6E3A1", Warning: "#F9E2AF", Stopped: "#6C7086", Errored: "#F38BA8", Muted: "#6C7086", Label: "#9399B2", SelBg: "#313244", SelFg: "#CDD6F4"},
	"catppuccin": {Accent: "183", Key: "111", Group: "223", Running: "151", Warning: "223", Stopped: "240", Errored: "211", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
	"dracula":    {Accent: "141", Key: "117", Group: "228", Running: "84", Warning: "228", Stopped: "240", Errored: "212", Muted: "240", Label: "245", SelBg: "238", SelFg: "253"},
	"tokyonight": {Accent: "111", Key: "117", Group: "222", Running: "149", Warning: "222", Stopped: "240", Errored: "210", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
	"nord":       {Accent: "109", Key: "110", Group: "222", Running: "108", Warning: "222", Stopped: "240", Errored: "167", Muted: "240", Label: "245", SelBg: "238", SelFg: "253"},
	"gruvbox":    {Accent: "214", Key: "109", Group: "214", Running: "142", Warning: "214", Stopped: "245", Errored: "167", Muted: "243", Label: "246", SelBg: "237", SelFg: "223"},
	"rosepine":   {Accent: "116", Key: "110", Group: "180", Running: "108", Warning: "180", Stopped: "240", Errored: "174", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
	"solarized":  {Accent: "37", Key: "33", Group: "136", Running: "100", Warning: "136", Stopped: "240", Errored: "160", Muted: "241", Label: "244", SelBg: "236", SelFg: "254"},
	"monokai":    {Accent: "197", Key: "81", Group: "186", Running: "148", Warning: "186", Stopped: "240", Errored: "208", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
	"mono":       {Accent: "250", Key: "250", Group: "250", Running: "250", Warning: "250", Stopped: "241", Errored: "244", Muted: "240", Label: "245", SelBg: "237", SelFg: "253"},
}

var (
	runningStyle      lipgloss.Style
	warningStyle      lipgloss.Style
	stoppedStyle      lipgloss.Style
	erroredStyle      lipgloss.Style
	headerStyle       lipgloss.Style
	groupStyle        lipgloss.Style
	keyStyle          lipgloss.Style
	hintStyle         lipgloss.Style
	labelStyle        lipgloss.Style
	selRowStyle       lipgloss.Style
	borderActiveStyle lipgloss.Style

	borderDim    lipgloss.Color
	borderActive lipgloss.Color
	labelDim     lipgloss.Color
	selectionBg  lipgloss.Color
)

func ApplyTheme(name string) {
	t, ok := Themes[name]
	if !ok {
		name = "default"
		t = Themes["default"]
	}
	CurrentTheme = name
	runningStyle = lipgloss.NewStyle().Foreground(t.Running)
	warningStyle = lipgloss.NewStyle().Foreground(t.Warning).Bold(true)
	stoppedStyle = lipgloss.NewStyle().Foreground(t.Stopped)
	erroredStyle = lipgloss.NewStyle().Foreground(t.Errored)
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(t.SelFg)
	groupStyle = lipgloss.NewStyle().Bold(true).Foreground(t.Group)
	keyStyle = lipgloss.NewStyle().Bold(true).Foreground(t.Key)
	hintStyle = lipgloss.NewStyle().Foreground(t.Muted)
	labelStyle = lipgloss.NewStyle().Foreground(t.Label)
	selRowStyle = lipgloss.NewStyle().Bold(true).Foreground(t.SelFg).Background(t.SelBg)
	borderActiveStyle = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)

	borderDim = t.Muted
	borderActive = t.Accent
	labelDim = t.Label
	selectionBg = t.SelBg
}

func init() {
	ApplyTheme("default")
}
