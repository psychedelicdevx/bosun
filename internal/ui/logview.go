package ui

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aymanbagabas/go-osc52/v2"
	tea "github.com/charmbracelet/bubbletea"
)

type logLineMsg struct {
	gen  int
	line string
}

type logClosedMsg struct{ gen int }

func waitForLog(gen int, ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return logClosedMsg{gen}
		}
		return logLineMsg{gen, line}
	}
}

func (m *Model) stopLogs() {
	if m.logCancel != nil {
		m.logCancel()
		m.logCancel = nil
	}
}

func (m Model) shownLogLines() []string {
	if m.logFilter == "" {
		return m.logLines
	}
	q := strings.ToLower(m.logFilter)
	out := make([]string, 0, len(m.logLines))
	for _, l := range m.logLines {
		if strings.Contains(strings.ToLower(l), q) {
			out = append(out, l)
		}
	}
	return out
}

func (m Model) renderedLogLines() []string {
	lines := m.shownLogLines()
	out := make([]string, len(lines))
	if m.logFilter == "" {
		for i, line := range lines {
			out[i] = labelStyle.Render(line)
		}
		return out
	}

	matcher := regexp.MustCompile("(?i:" + regexp.QuoteMeta(m.logFilter) + ")")
	for i, line := range lines {
		out[i] = renderLogLine(line, matcher.FindAllStringIndex(line, -1))
	}
	return out
}

func renderLogLine(line string, matches [][]int) string {
	var b strings.Builder
	last := 0
	for _, match := range matches {
		b.WriteString(labelStyle.Render(line[last:match[0]]))
		b.WriteString(borderActiveStyle.Render(line[match[0]:match[1]]))
		last = match[1]
	}
	b.WriteString(labelStyle.Render(line[last:]))
	return b.String()
}

func (m *Model) refreshLogView() {
	atBottom := m.vp.AtBottom()
	m.vp.SetContent(strings.Join(m.renderedLogLines(), "\n"))
	if atBottom {
		m.vp.GotoBottom()
	}
}

func (m Model) updateLogFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.logFiltering = false
	case tea.KeyEsc:
		m.logFiltering = false
		m.logFilter = ""
	case tea.KeyBackspace:
		if len(m.logFilter) > 0 {
			m.logFilter = m.logFilter[:len(m.logFilter)-1]
		}
	case tea.KeySpace:
		m.logFilter += " "
	case tea.KeyRunes:
		m.logFilter += string(msg.Runes)
	}
	m.refreshLogView()
	return m, nil
}

func (m Model) copyLogs() (tea.Model, tea.Cmd) {
	lines := m.shownLogLines()
	content := strings.Join(lines, "\n")
	n := len(lines)
	copyCmd := func() tea.Msg {
		_, _ = osc52.New(content).WriteTo(os.Stdout)
		return nil
	}
	return m, tea.Batch(copyCmd, m.setStatus(fmt.Sprintf("copied %d log lines to clipboard", n)))
}
