package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderedLogLinesHighlightMatchesWithoutChangingText(t *testing.T) {
	ApplyTheme("default")
	m := New(nil)
	m.logFilter = "job"
	m.logLines = []string{`level=info msg="Job picked up; job done"`}

	rendered := m.renderedLogLines()
	if len(rendered) != 1 {
		t.Fatalf("rendered lines = %d, want 1", len(rendered))
	}
	if got := ansi.Strip(rendered[0]); got != m.logLines[0] {
		t.Fatalf("rendered text = %q, want %q", got, m.logLines[0])
	}
	if got := strings.Count(rendered[0], borderActiveStyle.Render("Job")) + strings.Count(rendered[0], borderActiveStyle.Render("job")); got != 2 {
		t.Fatalf("highlighted matches = %d, want 2", got)
	}
}

func TestRenderedLogLinesHandlesUnicodeFilter(t *testing.T) {
	m := New(nil)
	m.logFilter = "çalış"
	m.logLines = []string{"Çalışıyor"}

	if got := ansi.Strip(m.renderedLogLines()[0]); got != "Çalışıyor" {
		t.Fatalf("rendered unicode line = %q", got)
	}
}
