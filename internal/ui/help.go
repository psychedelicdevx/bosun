package ui

import "strings"

func (m Model) helpBox() string {
	rows := [][2]string{
		{"↑ / ↓", "move (j / k)"},
		{"1-4", "switch view"},
		{"/", "filter list"},
		{"space", "toggle group"},
		{"tab", "switch focus"},
		{"enter", "open logs"},
		{"/ logs", "filter logs"},
		{"y logs", "copy logs"},
		{"S", "live stats"},
		{"e", "shell"},
		{"s", "start / stack"},
		{"x", "stop / stack"},
		{"r", "restart / stack"},
		{"d", "remove (confirm)"},
		{"p", "prune images"},
		{"esc", "back / clear"},
		{"T", "cycle theme"},
		{"?", "close help"},
		{"q", "quit"},
	}

	width, height := m.contentSize()
	if height < len(rows)+4 {
		return twoColumnHelp(rows, min(64, width))
	}

	var b strings.Builder
	b.WriteString("\n")
	for _, r := range rows {
		b.WriteString("  " + keyStyle.Render(pad(r[0], 10)) + labelStyle.Render(r[1]) + "\n")
	}

	return panel("Keybindings", strings.TrimRight(b.String(), "\n"), min(46, width), len(rows)+4, true)
}

func twoColumnHelp(rows [][2]string, width int) string {
	inner := panelContentWidth(width)
	leftWidth := inner / 2
	rightWidth := inner - leftWidth
	half := (len(rows) + 1) / 2

	var b strings.Builder
	b.WriteString("\n")
	for i := 0; i < half; i++ {
		b.WriteString(fit(helpRow(rows[i]), leftWidth))
		if right := i + half; right < len(rows) {
			b.WriteString(fit(helpRow(rows[right]), rightWidth))
		} else {
			b.WriteString(strings.Repeat(" ", rightWidth))
		}
		if i < half-1 {
			b.WriteString("\n")
		}
	}
	return panel("Keybindings", b.String(), width, half+4, true)
}

func helpRow(row [2]string) string {
	return keyStyle.Render(pad(row[0], 8)) + labelStyle.Render(row[1])
}

func pad(s string, w int) string {
	for len([]rune(s)) < w {
		s += " "
	}
	return s
}
