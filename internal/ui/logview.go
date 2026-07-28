package ui

import (
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
