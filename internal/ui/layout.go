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

func (m Model) listTitle() string {
	if m.filter != "" || m.filtering {
		return fmt.Sprintf("Containers (%d/%d)", len(m.visible()), len(m.containers))
	}
	return "Containers"
}

func (m Model) listBody() string {
	if !m.loaded {
		return "loading..."
	}
	rs := m.rows()
	if len(rs) == 0 {
		if m.filter != "" || m.filtering {
			return hintStyle.Render("no match")
		}
		return hintStyle.Render("no containers")
	}
	var b strings.Builder
	for i, r := range rs {
		marker := "  "
		if i == m.cursor {
			marker = borderActiveStyle.Render("› ")
		}
		if r.header {
			chevron := "▼"
			if m.collapsed[r.project] {
				chevron = "▶"
			}
			label := chevron + " " + r.project
			if i == m.cursor {
				label = selectedStyle.Render(label)
			} else {
				label = headerStyle.Render(label)
			}
			b.WriteString(marker + label)
		} else {
			name := r.ct.Name
			if i == m.cursor {
				name = selectedStyle.Render(name)
			}
			b.WriteString(marker + "  " + stateStyle(r.ct).Render("●") + " " + name)
		}
		if i < len(rs)-1 {
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
	if ct, ok := m.selected(); ok {
		name = ct.Name
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
		headerStyle.Render(project), labelStyle.Render("(project)"),
		labelStyle.Render("containers"), total,
		labelStyle.Render("running   "), running,
	)
}

func (m Model) bottomBar() string {
	if m.confirming {
		return erroredStyle.Render(" remove " + m.pendingName + "? (y/n)")
	}
	if m.filtering {
		return " " + labelStyle.Render("filter ") + m.filter + borderActiveStyle.Render("▏")
	}
	if m.filter != "" {
		return " " + labelStyle.Render("filter ") + m.filter + hintStyle.Render("   esc clears")
	}
	if m.status != "" {
		return hintStyle.Render(" " + m.status)
	}
	return hintStyle.Render(" / filter · tab focus · enter logs · S stats · e shell · s/x/r/d · ? help · q quit")
}
