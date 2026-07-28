package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

var (
	runningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	stoppedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	erroredStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	headerStyle   = lipgloss.NewStyle().Bold(true)
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("236"))
)

type Model struct {
	client     *docker.Client
	containers []docker.Container
	cursor     int
	loaded     bool
	err        error

	mode      mode
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

	width  int
	height int
}

func New(client *docker.Client) Model {
	return Model{
		client: client,
		vp:     viewport.New(80, 20),
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
		m.vp.Width = msg.Width
		m.vp.Height = max(1, msg.Height-3)
		return m, nil
	case tickMsg:
		return m, tea.Batch(m.fetch, tick())
	case containersMsg:
		m.containers = msg
		m.loaded = true
		if m.cursor > len(m.containers)-1 {
			m.cursor = max(0, len(m.containers)-1)
		}
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
		switch m.mode {
		case modeLogs:
			return m.updateLogs(msg)
		case modeStats:
			return m.updateStats(msg)
		case modeHelp:
			m.mode = modeList
			return m, nil
		default:
			return m.updateList(msg)
		}
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "down", "j":
		if m.cursor < len(m.containers)-1 {
			m.cursor++
		}
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "?":
		m.mode = modeHelp
	case "enter":
		if len(m.containers) == 0 {
			return m, nil
		}
		m.logGen++
		ctx, cancel := context.WithCancel(context.Background())
		ch, err := m.client.Logs(ctx, m.containers[m.cursor].ID)
		if err != nil {
			cancel()
			m.logLines = []string{"error opening logs: " + err.Error()}
			m.vp.SetContent(m.logLines[0])
			m.mode = modeLogs
			return m, nil
		}
		m.logCancel = cancel
		m.logChan = ch
		m.logLines = nil
		m.vp.SetContent("")
		m.vp.GotoTop()
		m.mode = modeLogs
		return m, waitForLog(m.logGen, ch)
	case "S":
		if len(m.containers) == 0 {
			return m, nil
		}
		ct := m.containers[m.cursor]
		if ct.State != "running" {
			return m, m.setStatus(ct.Name + " is not running")
		}
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
		m.mode = modeStats
		return m, waitForStats(m.statsGen, ch)
	case "e":
		if len(m.containers) == 0 {
			return m, nil
		}
		ct := m.containers[m.cursor]
		if ct.State != "running" {
			return m, m.setStatus(ct.Name + " is not running")
		}
		return m, execShell(ct.ID, ct.Name)
	case "s", "x", "r":
		if len(m.containers) == 0 {
			return m, nil
		}
		ct := m.containers[m.cursor]
		verb := map[string]string{"s": "start", "x": "stop", "r": "restart"}[msg.String()]
		return m, tea.Batch(m.setStatus(verb+"ing "+ct.Name+"..."), m.doAction(verb, ct.ID, ct.Name))
	case "d":
		if len(m.containers) == 0 {
			return m, nil
		}
		ct := m.containers[m.cursor]
		m.confirming = true
		m.pendingID = ct.ID
		m.pendingName = ct.Name
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	switch m.mode {
	case modeLogs:
		return m.logsView()
	case modeStats:
		return m.statsView()
	case modeHelp:
		return m.helpView()
	}
	return m.listView()
}

func (m Model) listView() string {
	if m.err != nil {
		return fmt.Sprintf("cannot reach docker: %v\n\n%s\n", m.err, hintStyle.Render("q quit"))
	}
	if !m.loaded {
		return "loading containers...\n"
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render("CONTAINERS") + "\n\n")
	if len(m.containers) == 0 {
		b.WriteString(hintStyle.Render("no containers") + "\n")
	}
	for i, ct := range m.containers {
		style := stoppedStyle
		switch ct.State {
		case "running":
			style = runningStyle
		case "exited", "dead":
			if strings.HasPrefix(ct.Status, "Exited (0)") {
				style = stoppedStyle
			} else {
				style = erroredStyle
			}
		}
		row := fmt.Sprintf("%-24s  %s", ct.Name, ct.Status)
		if i == m.cursor {
			b.WriteString(fmt.Sprintf("%s %s\n", style.Render("●"), selectedStyle.Render(row)))
		} else {
			b.WriteString(fmt.Sprintf("%s %s\n", style.Render("●"), row))
		}
	}
	b.WriteString("\n")
	if m.confirming {
		b.WriteString(erroredStyle.Render(fmt.Sprintf("remove %s? (y/n)", m.pendingName)) + "\n")
	} else if m.status != "" {
		b.WriteString(hintStyle.Render(m.status) + "\n")
	}
	b.WriteString(hintStyle.Render("↑/↓ move · enter logs · S stats · e shell · s/x/r/d actions · ? help · q quit") + "\n")
	return b.String()
}
