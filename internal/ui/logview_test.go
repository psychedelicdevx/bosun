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

func TestRenderedLogLinesColorSeverityAndStatus(t *testing.T) {
	ApplyTheme("default")
	m := New(nil)
	m.logLines = []string{
		`level=warn msg="retrying upstream" status=429`,
		`level=error msg="upstream failed" status=503`,
		`GET /missing 404 2ms id=502`,
	}

	rendered := m.renderedLogLines()
	if got := toneForText(m.logLines[0], "warn"); got != logToneWarning {
		t.Fatalf("warn tone = %v, want warning", got)
	}
	if got := toneForText(m.logLines[0], "429"); got != logToneWarning {
		t.Fatalf("429 tone = %v, want warning", got)
	}
	if got := toneForText(m.logLines[1], "error"); got != logToneError {
		t.Fatalf("error tone = %v, want error", got)
	}
	if got := toneForText(m.logLines[1], "503"); got != logToneError {
		t.Fatalf("503 tone = %v, want error", got)
	}
	for i, line := range rendered {
		if got := ansi.Strip(line); got != m.logLines[i] {
			t.Fatalf("rendered line %d changed text: %q", i, got)
		}
	}
	if got := toneForText(m.logLines[2], "404"); got != logToneWarning {
		t.Fatalf("404 tone = %v, want warning", got)
	}
	if got := toneForText(m.logLines[2], "502"); got != logToneNormal {
		t.Fatalf("unrelated numeric id tone = %v, want normal", got)
	}
}

func toneForText(line, text string) logTone {
	start := strings.Index(line, text)
	if start < 0 {
		return logToneNormal
	}
	return toneAt(semanticLogSpans(line), start, start+len(text))
}

func TestFilterHighlightOverridesSeverity(t *testing.T) {
	m := New(nil)
	m.logFilter = "warn"
	m.logLines = []string{`level=warn msg="retrying"`}

	rendered := m.renderedLogLines()[0]
	if !strings.Contains(rendered, borderActiveStyle.Render("warn")) {
		t.Fatalf("filter match should override warning color: %q", rendered)
	}
}
