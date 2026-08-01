package ui

import (
	"fmt"
	"os"
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

func (m *Model) refreshLogView() {
	atBottom := m.vp.AtBottom()
	m.vp.SetContent(strings.Join(m.shownLogLines(), "\n"))
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
