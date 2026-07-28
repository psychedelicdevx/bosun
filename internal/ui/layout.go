package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

const (
	borderDim    = lipgloss.Color("240")
	borderActive = lipgloss.Color("42")
	labelDim     = lipgloss.Color("245")
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
	bodyH := max(1, height-2)

	label := fit(" "+title+" ", max(1, inner-1))
	fill := max(0, inner-1-lipgloss.Width(label))
	top := bs.Render("╭─") + ts.Render(label) + bs.Render(strings.Repeat("─", fill)) + bs.Render("╮")

	lines := strings.Split(body, "\n")
	rows := make([]string, bodyH)
	for i := 0; i < bodyH; i++ {
		content := ""
		if i < len(lines) {
			content = lines[i]
		}
		rows[i] = bs.Render("│") + fit(content, inner) + bs.Render("│")
	}

	bottom := bs.Render("╰" + strings.Repeat("─", inner) + "╯")
	return top + "\n" + strings.Join(rows, "\n") + "\n" + bottom
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

func (m Model) dims() (leftW, rightW, panelH int) {
	leftW = min(40, max(26, m.width/3))
	rightW = m.width - leftW
	panelH = max(3, m.height-1)
	return
}

func (m Model) listBody() string {
	if !m.loaded {
		return "loading..."
	}
	if len(m.containers) == 0 {
		return hintStyle.Render("no containers")
	}
	var b strings.Builder
	for i, ct := range m.containers {
		name := ct.Name
		cur := "  "
		if i == m.cursor {
			cur = borderActiveStyle.Render("› ")
			name = selectedStyle.Render(name)
		}
		b.WriteString(cur + stateStyle(ct).Render("●") + " " + name)
		if i < len(m.containers)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func detailsBody(ct docker.Container) string {
	return fmt.Sprintf(
		"%s %s\n\n%s  %s\n%s  %s\n%s  %s\n%s  %s",
		stateStyle(ct).Render("●"), headerStyle.Render(ct.Name),
		labelStyle.Render("id    "), ct.ID,
		labelStyle.Render("image "), ct.Image,
		labelStyle.Render("state "), ct.State,
		labelStyle.Render("status"), ct.Status,
	)
}

func (m Model) statsBody() string {
	if !m.haveStats {
		return hintStyle.Render("waiting for samples...")
	}
	return fmt.Sprintf(
		"%s  %6.2f%%\n%s  %6.2f%%\n\n%s  %s / %s",
		labelStyle.Render("cpu"), m.stats.CPUPercent,
		labelStyle.Render("mem"), m.stats.MemPercent,
		labelStyle.Render("   "), humanBytes(m.stats.MemUsage), humanBytes(m.stats.MemLimit),
	)
}

func (m Model) rightTitle() string {
	name := ""
	if m.cursor < len(m.containers) {
		name = m.containers[m.cursor].Name
	}
	switch m.right {
	case viewLogs:
		return "Logs: " + name
	case viewStats:
		return "Stats: " + name
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
		if m.cursor >= len(m.containers) {
			return hintStyle.Render("no container selected")
		}
		return detailsBody(m.containers[m.cursor])
	}
}

func (m Model) bottomBar() string {
	if m.confirming {
		return erroredStyle.Render(" remove " + m.pendingName + "? (y/n)")
	}
	if m.status != "" {
		return hintStyle.Render(" " + m.status)
	}
	return hintStyle.Render(" tab focus · enter logs · S stats · e shell · s/x/r/d actions · ? help · q quit")
}
