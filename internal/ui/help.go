package ui

import "strings"

func (m Model) helpBox() string {
	rows := [][2]string{
		{"↑ / ↓", "move selection (also j / k)"},
		{"/", "filter the list by name"},
		{"space", "collapse or expand a compose group"},
		{"tab", "switch focus between panels"},
		{"enter", "stream logs for the selected container"},
		{"S", "live CPU / memory stats"},
		{"e", "shell into a running container"},
		{"s", "start container"},
		{"x", "stop container"},
		{"r", "restart container"},
		{"d", "remove container (asks to confirm)"},
		{"esc", "back to details"},
		{"T", "cycle color theme"},
		{"?", "close this help"},
		{"q", "quit"},
	}

	var b strings.Builder
	b.WriteString("\n")
	for _, r := range rows {
		b.WriteString("  " + headerStyle.Render(pad(r[0], 10)) + labelStyle.Render(r[1]) + "\n")
	}

	return panel("Keybindings", strings.TrimRight(b.String(), "\n"), 54, len(rows)+4, true)
}

func pad(s string, w int) string {
	for len([]rune(s)) < w {
		s += " "
	}
	return s
}
