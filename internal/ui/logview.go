package ui

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/aymanbagabas/go-osc52/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type logLineMsg struct {
	gen  int
	line string
}

type logLinesMsg struct {
	gen    int
	lines  []string
	closed bool
}

type logClosedMsg struct{ gen int }

var (
	logSeverityPattern     = regexp.MustCompile(`(?i)\b(?:warn(?:ing)?|error|fatal|panic)\b`)
	logStatusFieldPattern  = regexp.MustCompile(`(?i)\bstatus(?:_code)?[=: ]+([45][0-9]{2})\b`)
	logAccessStatusPattern = regexp.MustCompile(`\s([45][0-9]{2})\s+\d+(?:\.\d+)?(?:ms|s)\b`)
)

type logTone int

const (
	logToneNormal logTone = iota
	logToneWarning
	logToneError
	logToneFilter
)

type logSpan struct {
	start int
	end   int
	tone  logTone
}

func waitForLog(gen int, ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return logClosedMsg{gen}
		}
		lines := []string{line}
		for len(lines) < 256 {
			select {
			case next, open := <-ch:
				if !open {
					return logLinesMsg{gen: gen, lines: lines, closed: true}
				}
				lines = append(lines, next)
			default:
				return logLinesMsg{gen: gen, lines: lines}
			}
		}
		return logLinesMsg{gen: gen, lines: lines}
	}
}

func (m *Model) stopLogs() {
	if m.logCancel != nil {
		m.logCancel()
		m.logCancel = nil
		m.logGen++
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
			out[i] = renderLogLine(line, nil)
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
	spans := semanticLogSpans(line)
	for _, match := range matches {
		spans = append(spans, logSpan{start: match[0], end: match[1], tone: logToneFilter})
	}

	boundaries := []int{0, len(line)}
	for _, span := range spans {
		boundaries = append(boundaries, span.start, span.end)
	}
	sort.Ints(boundaries)

	var b strings.Builder
	for i := 0; i < len(boundaries)-1; i++ {
		start, end := boundaries[i], boundaries[i+1]
		if start == end {
			continue
		}
		b.WriteString(logToneStyle(toneAt(spans, start, end)).Render(line[start:end]))
	}
	return b.String()
}

func semanticLogSpans(line string) []logSpan {
	spans := make([]logSpan, 0, 2)
	for _, match := range logSeverityPattern.FindAllStringIndex(line, -1) {
		tone := logToneError
		if strings.HasPrefix(strings.ToLower(line[match[0]:match[1]]), "warn") {
			tone = logToneWarning
		}
		spans = append(spans, logSpan{start: match[0], end: match[1], tone: tone})
	}
	for _, pattern := range []*regexp.Regexp{logStatusFieldPattern, logAccessStatusPattern} {
		for _, match := range pattern.FindAllStringSubmatchIndex(line, -1) {
			start, end := match[2], match[3]
			tone := logToneWarning
			if line[start] == '5' {
				tone = logToneError
			}
			spans = append(spans, logSpan{start: start, end: end, tone: tone})
		}
	}
	return spans
}

func toneAt(spans []logSpan, start, end int) logTone {
	tone := logToneNormal
	for _, span := range spans {
		if start >= span.start && end <= span.end && span.tone > tone {
			tone = span.tone
		}
	}
	return tone
}

func logToneStyle(tone logTone) lipgloss.Style {
	switch tone {
	case logToneWarning:
		return warningStyle
	case logToneError:
		return erroredStyle.Bold(true)
	case logToneFilter:
		return borderActiveStyle
	default:
		return labelStyle
	}
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
		m.logFilter = removeLastRune(m.logFilter)
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
