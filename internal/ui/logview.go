package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	modeList mode = iota
	modeLogs
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

func (m Model) updateLogs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.stopLogs()
		return m, tea.Quit
	case "esc":
		m.stopLogs()
		m.mode = modeList
		return m, nil
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m Model) logsView() string {
	name := ""
	if m.cursor < len(m.containers) {
		name = m.containers[m.cursor].Name
	}
	return headerStyle.Render("LOGS: "+name) + "\n" +
		m.vp.View() + "\n" +
		hintStyle.Render("↑/↓ scroll · esc back · q quit")
}
