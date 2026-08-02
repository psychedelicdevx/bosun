package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

func panel(title, body string, width, height int, active bool) string {
	bc := borderDim
	lc := labelDim
	if active {
		bc = borderActive
		lc = borderActive
	}
	bs := lipgloss.NewStyle().Foreground(bc)
	ts := lipgloss.NewStyle().Foreground(lc).Bold(true)

	inner := max(1, width-2)
	padding := panelPadding(width)
	contentWidth := max(1, inner-2*padding)
	bodyH := max(1, height-2)

	label := lipgloss.NewStyle().MaxWidth(max(1, inner-1)).Render(" " + title + " ")
	fill := max(0, inner-1-lipgloss.Width(label))
	top := bs.Render("╭─") + ts.Render(label) + bs.Render(strings.Repeat("─", fill)) + bs.Render("╮")

	lines := strings.Split(body, "\n")
	rows := make([]string, bodyH)
	for i := 0; i < bodyH; i++ {
		content := ""
		if i < len(lines) {
			content = lines[i]
		}
		rows[i] = bs.Render("│") + strings.Repeat(" ", padding) + fit(content, contentWidth) + strings.Repeat(" ", padding) + bs.Render("│")
	}

	bottom := bs.Render("╰" + strings.Repeat("─", inner) + "╯")
	return top + "\n" + strings.Join(rows, "\n") + "\n" + bottom
}

func panelPadding(width int) int {
	if width >= 16 {
		return 1
	}
	return 0
}

func panelContentWidth(width int) int {
	return max(1, width-2-2*panelPadding(width))
}

func fit(s string, w int) string {
	s = lipgloss.NewStyle().MaxWidth(w).Render(s)
	if gap := w - lipgloss.Width(s); gap > 0 {
		s += strings.Repeat(" ", gap)
	}
	return s
}

func stateStyle(ct docker.Container) lipgloss.Style {
	switch ct.State {
	case "running":
		return runningStyle
	case "exited", "dead":
		if strings.HasPrefix(ct.Status, "Exited (0)") {
			return stoppedStyle
		}
		return erroredStyle
	}
	return stoppedStyle
}

func stateGlyph(ct docker.Container) string {
	switch ct.State {
	case "running":
		return "●"
	case "exited", "dead":
		if strings.HasPrefix(ct.Status, "Exited (0)") {
			return "○"
		}
		return "×"
	default:
		return "·"
	}
}

func (m Model) dims() (leftW, rightW, panelH int) {
	width, height := m.contentSize()
	gap := m.panelGap()
	leftW = min(40, max(26, width/3))
	rightW = max(3, width-leftW-gap)
	panelH = max(3, height-2-m.tabPanelGap())
	return
}

func (m Model) panelGap() int {
	width, _ := m.contentSize()
	if width >= 72 {
		return 1
	}
	return 0
}

func (m Model) tabPanelGap() int {
	_, height := m.contentSize()
	if height >= 12 {
		return 1
	}
	return 0
}

func (m Model) insets() (horizontal, vertical int) {
	if m.width >= 72 {
		horizontal = 1
	}
	if m.height >= 18 {
		vertical = 1
	}
	return
}

func (m Model) contentSize() (width, height int) {
	horizontal, vertical := m.insets()
	return max(1, m.width-2*horizontal), max(1, m.height-2*vertical)
}

func (m Model) insetView(view string) string {
	horizontal, vertical := m.insets()
	if horizontal > 0 {
		view = lipgloss.NewStyle().MarginLeft(horizontal).Render(view)
	}
	if vertical > 0 {
		padding := strings.Repeat("\n", vertical)
		view = padding + view + padding
	}
	return view
}

func (m Model) listTitle() string {
	if m.filter != "" || m.filtering {
		return fmt.Sprintf("Containers (%d/%d)", len(m.visible()), len(m.containers))
	}
	return "Containers"
}

func (m Model) listBody(inner, height int) string {
	if !m.loaded {
		return "loading..."
	}
	rs := m.rows()
	banner := resourceBanner(m.containerErr, len(rs) > 0)
	if len(rs) == 0 {
		if banner != "" {
			return banner
		}
		if m.filter != "" || m.filtering {
			return hintStyle.Render("no match")
		}
		return hintStyle.Render("no containers")
	}
	if banner != "" {
		if height <= 1 {
			return banner
		}
		height--
	}
	start, end := windowRange(len(rs), m.cursor, height)
	var b strings.Builder
	if banner != "" {
		b.WriteString(banner + "\n")
	}
	for i := start; i < end; i++ {
		r := rs[i]
		if i == m.cursor {
			b.WriteString(highlightRow(r, m.collapsed[r.project], inner))
		} else if r.header {
			chevron := "▼"
			if m.collapsed[r.project] {
				chevron = "▶"
			}
			b.WriteString(groupStyle.Render(chevron + " " + r.project))
		} else {
			b.WriteString("    " + stateStyle(r.ct).Render(stateGlyph(r.ct)) + " " + r.ct.Name)
		}
		if i < end-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func windowRange(total, cursor, height int) (start, end int) {
	if total <= 0 || height <= 0 {
		return 0, 0
	}
	height = min(height, total)
	cursor = clampIndex(cursor, total)
	start = max(0, cursor-height+1)
	end = min(total, start+height)
	return start, end
}

func resourceBanner(err error, hasData bool) string {
	if err == nil {
		return ""
	}
	if hasData {
		return warningStyle.Render("refresh failed; showing cached data")
	}
	return erroredStyle.Render("refresh failed") + "\n" + hintStyle.Render(err.Error())
}

func highlightRow(r row, collapsed bool, inner int) string {
	if r.header {
		chevron := "▼"
		if collapsed {
			chevron = "▶"
		}
		line := groupStyle.Background(selectionBg).Render(chevron + " " + r.project)
		return fillSelection(line, inner)
	}

	indent := selRowStyle.Render("    ")
	dot := stateStyle(r.ct).Background(selectionBg).Render(stateGlyph(r.ct))
	name := selRowStyle.Render(" " + r.ct.Name)
	return fillSelection(indent+dot+name, inner)
}

func fillSelection(line string, inner int) string {
	if gap := inner - lipgloss.Width(line); gap > 0 {
		line += lipgloss.NewStyle().Background(selectionBg).Render(strings.Repeat(" ", gap))
	}
	return line
}

func detailsBody(ct docker.Container) string {
	return fmt.Sprintf(
		"%s %s\n\n%s  %s\n%s  %s\n%s  %s\n%s  %s",
		stateStyle(ct).Render(stateGlyph(ct)), headerStyle.Render(ct.Name),
		labelStyle.Render("id    "), ct.ID,
		labelStyle.Render("image "), ct.Image,
		labelStyle.Render("state "), ct.State,
		labelStyle.Render("status"), ct.Status,
	)
}

func (m Model) statsBody() string {
	if !m.haveStats {
		if m.statsEnded {
			return warningStyle.Render("stream ended before the first sample") + "\n" + hintStyle.Render("press enter to reconnect")
		}
		return hintStyle.Render("waiting for samples...")
	}
	body := fmt.Sprintf(
		"%s  %6.2f%%\n%s  %6.2f%%\n\n%s  %s / %s",
		labelStyle.Render("cpu"), m.stats.CPUPercent,
		labelStyle.Render("mem"), m.stats.MemPercent,
		labelStyle.Render("   "), humanBytes(m.stats.MemUsage), humanBytes(m.stats.MemLimit),
	)
	if m.statsEnded {
		body += "\n\n" + warningStyle.Render("stream ended") + "  " + hintStyle.Render("enter reconnects")
	}
	return body
}

func (m Model) rightTitle() string {
	name := ""
	if ct, ok := m.selected(); ok {
		name = ct.Name
	}
	switch m.right {
	case viewLogs:
		title := "Logs: " + name
		if m.logEnded {
			title += " · ended"
		}
		return title
	case viewStats:
		title := "Stats: " + name
		if m.statsEnded {
			title += " · ended"
		}
		return title
	default:
		return "Details"
	}
}

func (m Model) rightBody() string {
	switch m.right {
	case viewLogs:
		return m.vp.View()
	case viewStats:
		return m.statsBody()
	default:
		if r, ok := m.currentRow(); ok && r.header {
			return m.projectSummary(r.project)
		}
		ct, ok := m.selected()
		if !ok {
			return hintStyle.Render("no container selected")
		}
		return detailsBody(ct)
	}
}

func (m Model) projectSummary(project string) string {
	total, running := 0, 0
	for _, c := range m.containers {
		p := c.Project
		if p == "" {
			p = standaloneKey
		}
		if p == project {
			total++
			if c.State == "running" {
				running++
			}
		}
	}
	return fmt.Sprintf(
		"%s %s\n\n%s  %d\n%s  %d",
		groupStyle.Render(project), labelStyle.Render("(project)"),
		labelStyle.Render("containers"), total,
		labelStyle.Render("running   "), running,
	)
}

func (m Model) bottomBar() string {
	if m.confirming {
		verb := "remove"
		if m.pendingKind == "prune" {
			verb = "prune"
		}
		return " " + erroredStyle.Render(verb+" "+m.pendingName+"?") + "  " + shortcuts(
			shortcut{"y", "confirm"}, shortcut{"n", "cancel"},
		)
	}
	if m.logFiltering {
		return " " + labelStyle.Render("log filter ") + m.logFilter + borderActiveStyle.Render("▏")
	}
	if m.filtering {
		return " " + labelStyle.Render("filter ") + m.filter + borderActiveStyle.Render("▏")
	}
	if m.status != "" {
		return hintStyle.Render(" " + m.status)
	}
	if m.focusRight && m.right == viewLogs {
		if m.logEnded {
			return " " + shortcuts(shortcut{"enter", "reconnect"}, shortcut{"y", "copy"}, shortcut{"esc", "back"}, shortcut{"?", "help"}, shortcut{"q", "quit"})
		}
		if m.logFilter != "" {
			return " " + labelStyle.Render("log filter ") + m.logFilter + "   " + shortcuts(
				shortcut{"/", "edit"}, shortcut{"esc", "clears"}, shortcut{"y", "copy"},
			)
		}
		return " " + shortcuts(
			shortcut{"/", "filter logs"}, shortcut{"y", "copy"}, shortcut{"esc", "back"}, shortcut{"?", "help"}, shortcut{"q", "quit"},
		)
	}
	if m.focusRight && m.right == viewStats {
		if m.statsEnded {
			return " " + shortcuts(shortcut{"enter", "reconnect"}, shortcut{"esc", "back"}, shortcut{"?", "help"}, shortcut{"q", "quit"})
		}
		return " " + shortcuts(shortcut{"esc", "back"}, shortcut{"?", "help"}, shortcut{"q", "quit"})
	}
	if m.filter != "" {
		return " " + labelStyle.Render("filter ") + m.filter + "   " + shortcuts(shortcut{"esc", "clears"})
	}
	switch m.tab {
	case tabImages:
		return " " + shortcuts(shortcut{"1-4", "view"}, shortcut{"d", "remove"}, shortcut{"p", "prune dangling"}, shortcut{"?", "help"}, shortcut{"q", "quit"})
	case tabVolumes, tabNetworks:
		return " " + shortcuts(shortcut{"1-4", "view"}, shortcut{"d", "remove"}, shortcut{"?", "help"}, shortcut{"q", "quit"})
	}
	return " " + shortcuts(
		shortcut{"1-4", "view"}, shortcut{"/", "filter"}, shortcut{"enter", "logs"}, shortcut{"S", "stats"}, shortcut{"e", "shell"}, shortcut{"s/x/r/d", "actions"}, shortcut{"?", "help"}, shortcut{"q", "quit"},
	)
}

func (m Model) fittedBottomBar() string {
	bar := m.bottomBar()
	width, _ := m.contentSize()
	if lipgloss.Width(bar) <= width {
		return bar
	}

	if !m.confirming && !m.filtering && !m.logFiltering && m.status == "" && m.tab == tabContainers && !(m.focusRight && m.right == viewLogs) {
		bar = " " + shortcuts(
			shortcut{"1-4", "view"}, shortcut{"/", "filter"}, shortcut{"enter", "logs"}, shortcut{"s/x/r/d", "act"}, shortcut{"?", "help"}, shortcut{"q", "quit"},
		)
		if lipgloss.Width(bar) <= width {
			return bar
		}
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(bar)
}

type shortcut struct {
	key   string
	label string
}

func shortcuts(items ...shortcut) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, keyStyle.Render(item.key)+hintStyle.Render(" "+item.label))
	}
	return strings.Join(parts, hintStyle.Render(" · "))
}
