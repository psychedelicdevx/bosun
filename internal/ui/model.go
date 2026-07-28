package ui

import (
	"context"
	"fmt"
	"strings"

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

	status      string
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

func (m Model) fetch() tea.Msg {
	list, err := m.client.List(context.Background())
	if err != nil {
		return errMsg{err}
	}
	return containersMsg(list)
}

func (m Model) Init() tea.Cmd {
	return m.fetch
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.vp.Width = msg.Width
		m.vp.Height = max(1, msg.Height-3)
		return m, nil
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
	case logLineMsg:
		if msg.gen != m.logGen {
			return m, nil
		}
		m.logLines = append(m.logLines, msg.line)
		m.vp.SetContent(strings.Join(m.logLines, "\n"))
		m.vp.GotoBottom()
		return m, waitForLog(m.logGen, m.logChan)
	case logClosedMsg:
		return m, nil
	case actionDoneMsg:
		m.status = actionStatus(msg)
		return m, m.fetch
	case tea.KeyMsg:
		if m.mode == modeLogs {
			return m.updateLogs(msg)
		}
		return m.updateList(msg)
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirming {
		switch msg.String() {
		case "y", "enter":
			m.confirming = false
			m.status = "removing " + m.pendingName + "..."
			return m, m.doAction("remove", m.pendingID, m.pendingName)
		default:
			m.confirming = false
			m.status = "cancelled"
			return m, nil
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
	case "s", "x", "r":
		if len(m.containers) == 0 {
			return m, nil
		}
		ct := m.containers[m.cursor]
		verb := map[string]string{"s": "start", "x": "stop", "r": "restart"}[msg.String()]
		m.status = verb + "ing " + ct.Name + "..."
		return m, m.doAction(verb, ct.ID, ct.Name)
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
	if m.mode == modeLogs {
		return m.logsView()
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
			style = erroredStyle
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
	b.WriteString(hintStyle.Render("↑/↓ move · enter logs · s start · x stop · r restart · d remove · q quit") + "\n")
	return b.String()
}
