package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

type statsMsg struct {
	gen int
	s   docker.Stats
}

type statsClosedMsg struct{ gen int }

func waitForStats(gen int, ch <-chan docker.Stats) tea.Cmd {
	return func() tea.Msg {
		s, ok := <-ch
		if !ok {
			return statsClosedMsg{gen}
		}
		return statsMsg{gen, s}
	}
}

func (m *Model) stopStats() {
	if m.statsCancel != nil {
		m.statsCancel()
		m.statsCancel = nil
	}
}

func humanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
