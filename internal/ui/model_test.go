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

func TestFilterMatchesAndSelects(t *testing.T) {
	m := New(nil)
	m = send(m, containersMsg{
		{Name: "web", ID: "1"}, {Name: "api", ID: "2"},
		{Name: "web-worker", ID: "3"}, {Name: "db", ID: "4"},
	})

	m = send(m, key("/"))
	if !m.filtering {
		t.Fatal("/ should enter filtering")
	}
	for _, r := range "web" {
		m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if got := len(m.visible()); got != 2 {
		t.Fatalf("filter web: want 2 visible, got %d", got)
	}

	sel, ok := m.selected()
	if !ok || sel.ID != "1" {
		t.Fatalf("cursor should land on first match web, got %+v ok=%v", sel, ok)
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.filtering || m.filter != "" || len(m.visible()) != 4 {
		t.Fatalf("esc should clear filter, got filtering=%v filter=%q visible=%d", m.filtering, m.filter, len(m.visible()))
	}
}

func TestLogBufferCapped(t *testing.T) {
	m := New(nil)
	for i := 0; i < logBufferMax+logBufferSlack+2000; i++ {
		m = send(m, logLineMsg{gen: 0, line: "line"})
	}
	if len(m.logLines) > logBufferMax+logBufferSlack {
		t.Fatalf("log buffer unbounded: got %d, want <= %d", len(m.logLines), logBufferMax+logBufferSlack)
	}
	if len(m.logLines) < logBufferMax {
		t.Fatalf("log buffer trimmed too aggressively: got %d", len(m.logLines))
	}
}

func TestLogFilter(t *testing.T) {
	m := New(nil)
	m.right = viewLogs
	m.focusRight = true
	m.logLines = []string{"info: started", "ERROR: boom", "info: ok", "error: again"}

	m = send(m, key("/"))
	if !m.logFiltering {
		t.Fatal("/ in log view should enter log filtering")
	}
	for _, r := range "error" {
		m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if got := len(m.shownLogLines()); got != 2 {
		t.Fatalf("filter error: want 2 shown, got %d (%v)", got, m.shownLogLines())
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.logFiltering {
		t.Fatal("enter should stop typing but keep filter")
	}
	if m.logFilter != "error" || len(m.shownLogLines()) != 2 {
		t.Fatalf("filter should persist after enter, got %q shown=%d", m.logFilter, len(m.shownLogLines()))
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.logFilter != "" || len(m.shownLogLines()) != 4 {
		t.Fatalf("esc should clear log filter, got %q shown=%d", m.logFilter, len(m.shownLogLines()))
	}
}

func TestImageRemoveNeedsConfirm(t *testing.T) {
	m := New(nil)
	m = send(m, imagesMsg{{Repo: "nginx:1.27", ID: "abc"}})

	m = send(m, key("2"))
	if m.tab != tabImages {
		t.Fatal("key 2 should switch to images view")
	}

	m = send(m, key("d"))
	if !m.confirming || m.pendingKind != "image" || m.pendingName != "nginx:1.27" {
		t.Fatalf("d should arm image confirm, got confirming=%v kind=%q name=%q", m.confirming, m.pendingKind, m.pendingName)
	}

	m = send(m, key("n"))
	if m.confirming || m.pendingKind != "" {
		t.Fatalf("n should cancel image confirm, got confirming=%v kind=%q", m.confirming, m.pendingKind)
	}
}

func TestVolumeRemoveNeedsConfirm(t *testing.T) {
	m := New(nil)
	m = send(m, volumesMsg{{Name: "pgdata"}, {Name: "redis"}})

	m = send(m, key("3"))
	if m.tab != tabVolumes {
		t.Fatal("key 3 should switch to volumes view")
	}

	m = send(m, key("d"))
	if !m.confirming || m.pendingKind != "volume" || m.pendingName != "pgdata" {
		t.Fatalf("d should arm volume confirm, got confirming=%v kind=%q name=%q", m.confirming, m.pendingKind, m.pendingName)
	}

	m = send(m, key("n"))
	if m.confirming || m.pendingKind != "" {
		t.Fatalf("n should cancel volume confirm, got confirming=%v kind=%q", m.confirming, m.pendingKind)
	}
}

func TestComposeGroupingCollapse(t *testing.T) {
	m := New(nil)
	m = send(m, containersMsg{
		{Name: "shop-web", ID: "1", Project: "shop"},
		{Name: "shop-db", ID: "2", Project: "shop"},
		{Name: "mailpit", ID: "3"},
	})

	// rows: header(shop), shop-web, shop-db, header(standalone), mailpit
	if got := len(m.rows()); got != 5 {
		t.Fatalf("want 5 rows, got %d", got)
	}
	if r, _ := m.currentRow(); !r.header || r.project != "shop" {
		t.Fatalf("cursor 0 should be shop header, got %+v", r)
	}

	// collapse shop (cursor on its header)
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	if got := len(m.rows()); got != 3 {
		t.Fatalf("collapsed shop: want 3 rows, got %d", got)
	}
}
