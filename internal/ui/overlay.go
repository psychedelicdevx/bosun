package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func overlay(bg, fg string, x, y int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	for i, fgLine := range fgLines {
		row := y + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		bgLine := bgLines[row]
		w := ansi.StringWidth(bgLine)
		if w < x {
			bgLine += strings.Repeat(" ", x-w)
		}
		fgWidth := ansi.StringWidth(fgLine)
		left := ansi.Truncate(bgLine, x, "")
		right := ansi.TruncateLeft(bgLine, x+fgWidth, "")
		bgLines[row] = left + "\x1b[0m" + fgLine + "\x1b[0m" + right
	}
	return strings.Join(bgLines, "\n")
}
