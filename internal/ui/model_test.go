package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func send(m Model, msgs ...tea.Msg) Model {
	for _, msg := range msgs {
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m
}

func TestCursorStaysInBounds(t *testing.T) {
	m := New(nil)
	m = send(m, containersMsg{{Name: "a"}, {Name: "b"}})

	m = send(m, key("up"))
	if m.cursor != 0 {
		t.Fatalf("up at top: want 0, got %d", m.cursor)
	}

	m = send(m, key("down"), key("down"), key("j"))
	if m.cursor != 1 {
		t.Fatalf("down past end: want 1, got %d", m.cursor)
	}

	m = send(m, containersMsg{{Name: "only"}})
	if m.cursor != 0 {
		t.Fatalf("reclamp after shrink: want 0, got %d", m.cursor)
	}
}

func TestStaleLogLinesIgnored(t *testing.T) {
	m := New(nil)
	m.right = viewLogs
	m.logGen = 2

	m = send(m, logLineMsg{gen: 1, line: "old stream"})
	if len(m.logLines) != 0 {
		t.Fatalf("stale gen should be ignored, got %v", m.logLines)
	}

	m = send(m, logLineMsg{gen: 2, line: "current"})
	if len(m.logLines) != 1 || m.logLines[0] != "current" {
		t.Fatalf("current gen should append, got %v", m.logLines)
	}
}

func TestRemoveNeedsConfirm(t *testing.T) {
	m := New(nil)
	m = send(m, containersMsg{{Name: "web", ID: "abc"}})

	m = send(m, key("d"))
	if !m.confirming || m.pendingName != "web" {
		t.Fatalf("d should arm confirm for web, got confirming=%v name=%q", m.confirming, m.pendingName)
	}

	m = send(m, key("n"))
	if m.confirming || m.status != "cancelled" {
		t.Fatalf("n should cancel, got confirming=%v status=%q", m.confirming, m.status)
	}
}
