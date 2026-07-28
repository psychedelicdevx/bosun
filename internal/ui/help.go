package ui

import "strings"

func (m Model) helpView() string {
	rows := [][2]string{
		{"↑ / ↓", "move selection (also j / k)"},
		{"enter", "stream logs for selected container"},
		{"S", "live CPU / memory stats"},
		{"e", "exec a shell into a running container"},
		{"s", "start container"},
		{"x", "stop container"},
		{"r", "restart container"},
		{"d", "remove container (asks to confirm)"},
		{"esc", "back to the list"},
		{"?", "toggle this help"},
		{"q", "quit"},
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render("bosun — keybindings") + "\n\n")
	for _, r := range rows {
		b.WriteString("  " + headerStyle.Render(pad(r[0], 8)) + hintStyle.Render(r[1]) + "\n")
	}
	b.WriteString("\n" + hintStyle.Render("press any key to close") + "\n")
	return b.String()
}

func pad(s string, w int) string {
	for len([]rune(s)) < w {
		s += " "
	}
	return s
}
