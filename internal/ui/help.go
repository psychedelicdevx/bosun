package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) helpOverlay() string {
	rows := [][2]string{
		{"↑ / ↓", "move selection (also j / k)"},
		{"tab", "switch focus between panels"},
		{"enter", "stream logs for the selected container"},
		{"S", "live CPU / memory stats"},
		{"e", "shell into a running container"},
		{"s", "start container"},
		{"x", "stop container"},
		{"r", "restart container"},
		{"d", "remove container (asks to confirm)"},
		{"esc", "back to details"},
		{"?", "toggle this help"},
		{"q", "quit"},
	}

	var b strings.Builder
	for i, r := range rows {
		b.WriteString(headerStyle.Render(pad(r[0], 8)) + labelStyle.Render(r[1]))
		if i < len(rows)-1 {
			b.WriteString("\n")
		}
	}

	box := panel("Keybindings", b.String(), 44, len(rows)+2, true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func pad(s string, w int) string {
	for len([]rune(s)) < w {
		s += " "
	}
	return s
}
