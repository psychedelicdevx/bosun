package ui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

var (
	runningStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	stoppedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	erroredStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	headerStyle       = lipgloss.NewStyle().Bold(true)
	hintStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	labelStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	selRowStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("237"))
	borderActiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)

type rightView int

const (
	viewDetails rightView = iota
	viewLogs
	viewStats
)

type Model struct {
	client     Engine
	containers []docker.Container
	cursor     int
	loaded     bool
	err        error

	right      rightView
	focusRight bool
	helpOpen   bool

	vp        viewport.Model
	logLines  []string
	logChan   <-chan string
	logCancel context.CancelFunc
	logGen    int

	stats       docker.Stats
	haveStats   bool
	statsChan   <-chan docker.Stats
	statsCancel context.CancelFunc
	statsGen    int

	status      string
	statusGen   int
	confirming  bool
	pendingID   string
	pendingName string

	filter    string
	filtering bool

	collapsed map[string]bool

	width  int
	height int
}

func New(client Engine) Model {
	return Model{
		client:    client,
		vp:        viewport.New(80, 20),
		collapsed: map[string]bool{},
	}
}

type containersMsg []docker.Container

type errMsg struct{ err error }

type tickMsg struct{}

type statusClearMsg struct{ gen int }

func tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *Model) setStatus(s string) tea.Cmd {
	m.status = s
	m.statusGen++
	gen := m.statusGen
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return statusClearMsg{gen} })
}

func (m *Model) toDetails() {
	m.stopLogs()
	m.stopStats()
	m.right = viewDetails
	m.focusRight = false
}

func (m Model) visible() []docker.Container {
	if m.filter == "" {
		return m.containers
	}
	q := strings.ToLower(m.filter)
	var out []docker.Container
	for _, c := range m.containers {
		if strings.Contains(strings.ToLower(c.Name), q) {
			out = append(out, c)
		}
	}
	return out
}

func (m Model) selected() (docker.Container, bool) {
	r, ok := m.currentRow()
	if !ok || r.header {
		return docker.Container{}, false
	}
	return r.ct, true
}

func (m *Model) clampCursor() {
	n := len(m.rows())
	if m.cursor > n-1 {
		m.cursor = max(0, n-1)
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m Model) fetch() tea.Msg {
	list, err := m.client.List(context.Background())
	if err != nil {
		return errMsg{err}
	}
	return containersMsg(list)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetch, tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		_, rightW, panelH := m.dims()
		m.vp.Width = max(1, rightW-2)
		m.vp.Height = max(1, panelH-2)
		return m, nil
	case tickMsg:
		return m, tea.Batch(m.fetch, tick())
	case containersMsg:
		m.containers = msg
		m.loaded = true
		m.clampCursor()
		return m, nil
	case errMsg:
		m.err = msg.err
		m.loaded = true
		return m, nil
	case statusClearMsg:
		if msg.gen == m.statusGen {
			m.status = ""
		}
		return m, nil
	case logLineMsg:
		if msg.gen != m.logGen {
			return m, nil
		}
		atBottom := m.vp.AtBottom()
		m.logLines = append(m.logLines, msg.line)
		m.vp.SetContent(strings.Join(m.logLines, "\n"))
		if atBottom {
			m.vp.GotoBottom()
		}
		return m, waitForLog(m.logGen, m.logChan)
	case logClosedMsg:
		return m, nil
	case statsMsg:
		if msg.gen != m.statsGen {
			return m, nil
		}
		m.stats = msg.s
		m.haveStats = true
		return m, waitForStats(m.statsGen, m.statsChan)
	case statsClosedMsg:
		return m, nil
	case actionDoneMsg:
		return m, tea.Batch(m.setStatus(actionStatus(msg)), m.fetch)
	case execDoneMsg:
		if msg.err != nil {
			return m, m.setStatus("exec " + msg.name + " failed: " + msg.err.Error())
		}
		return m, tea.Batch(m.setStatus("left shell: "+msg.name), m.fetch)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.helpOpen {
		m.helpOpen = false
		return m, nil
	}
	if m.confirming {
		switch msg.String() {
		case "y", "enter":
			m.confirming = false
			return m, tea.Batch(m.setStatus("removing "+m.pendingName+"..."), m.doAction("remove", m.pendingID, m.pendingName))
		default:
			m.confirming = false
			return m, m.setStatus("cancelled")
		}
	}

	if m.filtering {
		return m.updateFilter(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.stopLogs()
		m.stopStats()
		return m, tea.Quit
	case "?":
		m.helpOpen = true
		return m, nil
	case "tab":
		m.focusRight = !m.focusRight
		return m, nil
	case "/":
		m.toDetails()
		m.filtering = true
		m.cursor = 0
		return m, nil
	case "esc":
		if m.right != viewDetails || m.focusRight {
			m.toDetails()
			return m, nil
		}
		if m.filter != "" {
			m.filter = ""
			m.cursor = 0
		}
		return m, nil
	}

	if m.focusRight && m.right == viewLogs {
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "down", "j":
		if m.cursor < len(m.rows())-1 {
			m.toDetails()
			m.cursor++
		}
	case "up", "k":
		if m.cursor > 0 {
			m.toDetails()
			m.cursor--
		}
	case " ":
		if r, ok := m.currentRow(); ok && r.header {
			m.collapsed[r.project] = !m.collapsed[r.project]
			m.clampCursor()
		}
	case "enter":
		if r, ok := m.currentRow(); ok && r.header {
			m.collapsed[r.project] = !m.collapsed[r.project]
			m.clampCursor()
			return m, nil
		}
		return m.openLogs()
	case "S":
		return m.openStats()
	case "e":
		ct, ok := m.selected()
		if !ok {
			return m, nil
		}
		if ct.State != "running" {
			return m, m.setStatus(ct.Name + " is not running")
		}
		return m, execShell(ct.ID, ct.Name)
	case "s", "x", "r":
		ct, ok := m.selected()
		if !ok {
			return m, nil
		}
		verb := map[string]string{"s": "start", "x": "stop", "r": "restart"}[msg.String()]
		return m, tea.Batch(m.setStatus(verb+"ing "+ct.Name+"..."), m.doAction(verb, ct.ID, ct.Name))
	case "d":
		ct, ok := m.selected()
		if !ok {
			return m, nil
		}
		m.confirming = true
		m.pendingID = ct.ID
		m.pendingName = ct.Name
	}
	return m, nil
}

func (m Model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.filtering = false
	case tea.KeyEsc:
		m.filtering = false
		m.filter = ""
		m.cursor = 0
	case tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.cursor = 0
		}
	case tea.KeySpace:
		m.filter += " "
		m.cursor = 0
	case tea.KeyRunes:
		m.filter += string(msg.Runes)
		m.cursor = 0
	}
	return m, nil
}

func (m Model) openLogs() (tea.Model, tea.Cmd) {
	sel, ok := m.selected()
	if !ok {
		return m, nil
	}
	m.stopStats()
	m.logGen++
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := m.client.Logs(ctx, sel.ID)
	if err != nil {
		cancel()
		return m, m.setStatus("logs failed: " + err.Error())
	}
	m.logCancel = cancel
	m.logChan = ch
	m.logLines = nil
	m.vp.SetContent("")
	m.vp.GotoTop()
	m.right = viewLogs
	m.focusRight = true
	return m, waitForLog(m.logGen, ch)
}

func (m Model) openStats() (tea.Model, tea.Cmd) {
	ct, ok := m.selected()
	if !ok {
		return m, nil
	}
	if ct.State != "running" {
		return m, m.setStatus(ct.Name + " is not running")
	}
	m.stopLogs()
	m.statsGen++
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := m.client.Stats(ctx, ct.ID)
	if err != nil {
		cancel()
		return m, m.setStatus("stats failed: " + err.Error())
	}
	m.statsCancel = cancel
	m.statsChan = ch
	m.haveStats = false
	m.right = viewStats
	m.focusRight = true
	return m, waitForStats(m.statsGen, ch)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading bosun..."
	}
	if m.err != nil {
		return panel("Error", "cannot reach docker:\n\n"+m.err.Error(), m.width, m.height-1, true) +
			"\n" + hintStyle.Render(" q quit")
	}
	leftW, rightW, panelH := m.dims()
	left := panel(m.listTitle(), m.listBody(leftW-2), leftW, panelH, !m.focusRight && !m.helpOpen)
	right := panel(m.rightTitle(), m.rightBody(), rightW, panelH, m.focusRight && !m.helpOpen)
	view := lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + m.bottomBar()

	if m.helpOpen {
		box := m.helpBox()
		x := max(0, (m.width-lipgloss.Width(box))/2)
		y := max(0, (m.height-lipgloss.Height(box))/2)
		view = overlay(view, box, x, y)
	}
	return view
}
