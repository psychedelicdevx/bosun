package ui

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

type rightView int

const (
	viewDetails rightView = iota
	viewLogs
	viewStats
)

const (
	logBufferMax   = 5000
	logBufferSlack = 1000
)

type Model struct {
	client         Engine
	containers     []docker.Container
	cursor         int
	loaded         bool
	err            error
	containerErr   error
	refreshing     bool
	refreshPending bool

	right      rightView
	focusRight bool
	helpOpen   bool

	vp           viewport.Model
	logLines     []string
	logChan      <-chan string
	logCancel    context.CancelFunc
	logGen       int
	logFilter    string
	logFiltering bool
	logEnded     bool

	stats       docker.Stats
	haveStats   bool
	statsChan   <-chan docker.Stats
	statsCancel context.CancelFunc
	statsGen    int
	statsEnded  bool

	tab            tab
	images         []docker.Image
	imgCursor      int
	imagesLoaded   bool
	imagesErr      error
	volumes        []docker.Volume
	volCursor      int
	volumesLoaded  bool
	volumesErr     error
	networks       []docker.Network
	netCursor      int
	networksLoaded bool
	networksErr    error

	status      string
	statusGen   int
	confirming  bool
	pendingKind string
	pendingID   string
	pendingName string
	busy        bool

	filter    string
	filtering bool

	collapsed map[string]bool
	themeIdx  int

	width  int
	height int
}

func New(client Engine) Model {
	return Model{
		client:     client,
		vp:         viewport.New(80, 20),
		collapsed:  map[string]bool{},
		themeIdx:   themeIndex(CurrentTheme),
		refreshing: true,
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

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.refreshAll, tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		_, rightW, panelH := m.dims()
		m.vp.Width = panelContentWidth(rightW)
		m.vp.Height = max(1, panelH-2)
		return m, nil
	case tickMsg:
		return m, tea.Batch(m.beginRefresh(), tick())
	case refreshMsg:
		m.applyRefresh(msg)
		if m.refreshPending {
			m.refreshPending = false
			return m, m.beginRefresh()
		}
		return m, nil
	case imagesMsg:
		m.images = msg
		m.imagesLoaded = true
		m.imagesErr = nil
		if m.imgCursor > len(m.images)-1 {
			m.imgCursor = max(0, len(m.images)-1)
		}
		return m, nil
	case volumesMsg:
		m.volumes = msg
		m.volumesLoaded = true
		m.volumesErr = nil
		if m.volCursor > len(m.volumes)-1 {
			m.volCursor = max(0, len(m.volumes)-1)
		}
		return m, nil
	case networksMsg:
		m.networks = msg
		m.networksLoaded = true
		m.networksErr = nil
		if m.netCursor > len(m.networks)-1 {
			m.netCursor = max(0, len(m.networks)-1)
		}
		return m, nil
	case resourceDoneMsg:
		m.busy = false
		if msg.err != nil {
			return m, m.setStatus(msg.verb + " " + msg.name + " failed: " + msg.err.Error())
		}
		return m, tea.Batch(m.setStatus(msg.verb+" "+msg.name), m.requestRefresh())
	case imageDoneMsg:
		m.busy = false
		if msg.err != nil {
			return m, m.setStatus(msg.verb + " failed: " + msg.err.Error())
		}
		status := msg.verb + " " + msg.name
		if msg.verb == "pruned" {
			status = "pruned dangling images, freed " + humanBytes(msg.reclaimed)
		}
		return m, tea.Batch(m.setStatus(status), m.requestRefresh())
	case containersMsg:
		m.containers = msg
		m.loaded = true
		m.err = nil
		m.containerErr = nil
		m.clampCursor()
		return m, nil
	case errMsg:
		m.containerErr = msg.err
		if !m.loaded {
			m.err = msg.err
		}
		return m, nil
	case statusClearMsg:
		if msg.gen == m.statusGen && !m.busy {
			m.status = ""
		}
		return m, nil
	case logLineMsg:
		if msg.gen != m.logGen {
			return m, nil
		}
		m.appendLogLines([]string{msg.line})
		return m, waitForLog(m.logGen, m.logChan)
	case logLinesMsg:
		if msg.gen != m.logGen {
			return m, nil
		}
		m.appendLogLines(msg.lines)
		if msg.closed {
			m.logCancel = nil
			m.logEnded = true
			return m, m.setStatus("log stream ended; enter reconnects")
		}
		return m, waitForLog(m.logGen, m.logChan)
	case logClosedMsg:
		if msg.gen != m.logGen || m.right != viewLogs {
			return m, nil
		}
		m.logCancel = nil
		m.logEnded = true
		return m, m.setStatus("log stream ended; enter reconnects")
	case statsMsg:
		if msg.gen != m.statsGen {
			return m, nil
		}
		m.stats = msg.s
		m.haveStats = true
		return m, waitForStats(m.statsGen, m.statsChan)
	case statsClosedMsg:
		if msg.gen != m.statsGen || m.right != viewStats {
			return m, nil
		}
		m.statsCancel = nil
		m.statsEnded = true
		return m, m.setStatus("stats stream ended; enter reconnects")
	case actionDoneMsg:
		m.busy = false
		return m, tea.Batch(m.setStatus(actionStatus(msg)), m.requestRefresh())
	case execDoneMsg:
		if msg.err != nil {
			return m, m.setStatus("exec " + msg.name + " failed: " + msg.err.Error())
		}
		return m, tea.Batch(m.setStatus("left shell: "+msg.name), m.requestRefresh())
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.helpOpen {
		switch msg.String() {
		case "q", "ctrl+c":
			m.stopLogs()
			m.stopStats()
			return m, tea.Quit
		default:
			m.helpOpen = false
			return m, nil
		}
	}
	if m.confirming {
		switch msg.String() {
		case "y", "enter":
			m.confirming = false
			kind, id, name := m.pendingKind, m.pendingID, m.pendingName
			m.pendingKind = ""
			m.busy = true
			switch kind {
			case "prune":
				return m, tea.Batch(m.setStatus("pruning dangling images..."), m.pruneImages())
			case "image":
				return m, tea.Batch(m.setStatus("removing image "+name+"..."), m.removeImage(id, name))
			case "volume":
				return m, tea.Batch(m.setStatus("removing volume "+name+"..."), m.removeVolume(name))
			case "network":
				return m, tea.Batch(m.setStatus("removing network "+name+"..."), m.removeNetwork(id, name))
			default:
				return m, tea.Batch(m.setStatus("removing "+name+"..."), m.doAction("remove", id, name))
			}
		default:
			m.confirming = false
			m.pendingKind = ""
			return m, m.setStatus("cancelled")
		}
	}

	if m.filtering {
		return m.updateFilter(msg)
	}

	if m.logFiltering {
		return m.updateLogFilter(msg)
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
		if m.tab == tabContainers && m.right != viewDetails {
			m.focusRight = !m.focusRight
		}
		return m, nil
	case "1":
		m.tab = tabContainers
		return m, nil
	case "2":
		if m.tab != tabImages {
			m.toDetails()
			m.tab = tabImages
		}
		return m, nil
	case "3":
		if m.tab != tabVolumes {
			m.toDetails()
			m.tab = tabVolumes
		}
		return m, nil
	case "4":
		if m.tab != tabNetworks {
			m.toDetails()
			m.tab = tabNetworks
		}
		return m, nil
	case "T":
		m.themeIdx = (m.themeIdx + 1) % len(ThemeNames)
		ApplyTheme(ThemeNames[m.themeIdx])
		return m, m.setStatus("theme: " + ThemeNames[m.themeIdx])
	case "/":
		if m.focusRight && m.right == viewLogs {
			m.logFiltering = true
			return m, nil
		}
		if m.tab != tabContainers {
			return m, nil
		}
		m.toDetails()
		m.filtering = true
		m.cursor = 0
		return m, nil
	case "esc":
		if m.right == viewLogs && (m.logFilter != "" || m.logFiltering) {
			m.logFiltering = false
			m.logFilter = ""
			m.refreshLogView()
			return m, nil
		}
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
		if msg.String() == "y" {
			return m.copyLogs()
		}
		if msg.String() == "enter" && m.logEnded {
			return m.openLogs()
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	if m.focusRight && m.right == viewStats {
		if msg.String() == "enter" && m.statsEnded {
			return m.openStats()
		}
		return m, nil
	}

	if m.tab == tabImages {
		return m.updateImages(msg)
	}
	if m.tab == tabVolumes {
		return m.updateVolumes(msg)
	}
	if m.tab == tabNetworks {
		return m.updateNetworks(msg)
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
		if m.busy {
			return m, nil
		}
		verb := map[string]string{"s": "start", "x": "stop", "r": "restart"}[msg.String()]
		if r, ok := m.currentRow(); ok && r.header {
			ids := m.projectContainerIDs(r.project)
			if len(ids) == 0 {
				return m, nil
			}
			m.busy = true
			return m, tea.Batch(m.setStatus(gerund[verb]+" "+r.project+" stack..."), m.stackAction(verb, r.project, ids))
		}
		ct, ok := m.selected()
		if !ok {
			return m, nil
		}
		m.busy = true
		return m, tea.Batch(m.setStatus(gerund[verb]+" "+ct.Name+"..."), m.doAction(verb, ct.ID, ct.Name))
	case "d":
		if m.busy {
			return m, nil
		}
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
		m.filter = removeLastRune(m.filter)
		m.cursor = 0
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
	m.stopLogs()
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
	m.logFilter = ""
	m.logFiltering = false
	m.logEnded = false
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
	m.stopStats()
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
	m.statsEnded = false
	m.right = viewStats
	m.focusRight = true
	return m, waitForStats(m.statsGen, ch)
}

func (m *Model) appendLogLines(lines []string) {
	if len(lines) == 0 {
		return
	}
	atBottom := m.vp.AtBottom()
	m.logLines = append(m.logLines, lines...)
	if len(m.logLines) > logBufferMax+logBufferSlack {
		m.logLines = append([]string(nil), m.logLines[len(m.logLines)-logBufferMax:]...)
	}
	m.vp.SetContent(strings.Join(m.renderedLogLines(), "\n"))
	if atBottom {
		m.vp.GotoBottom()
	}
}

func removeLastRune(s string) string {
	if s == "" {
		return s
	}
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading bosun..."
	}
	if m.err != nil {
		width, height := m.contentSize()
		view := panel("Error", "cannot reach container engine:\n\n"+m.err.Error(), width, height-1, true) +
			"\n" + hintStyle.Render(" q quit")
		return m.insetView(view)
	}
	leftW, rightW, panelH := m.dims()

	leftActive := !m.focusRight && !m.helpOpen
	rightActive := m.focusRight && !m.helpOpen
	bodyHeight := max(1, panelH-2)
	var left, right string
	switch m.tab {
	case tabImages:
		left = panel("Images", m.imageListBody(panelContentWidth(leftW), bodyHeight), leftW, panelH, leftActive)
		right = panel("Image", m.imageDetail(), rightW, panelH, rightActive)
	case tabVolumes:
		left = panel("Volumes", m.volumeListBody(panelContentWidth(leftW), bodyHeight), leftW, panelH, leftActive)
		right = panel("Volume", m.volumeDetail(), rightW, panelH, rightActive)
	case tabNetworks:
		left = panel("Networks", m.networkListBody(panelContentWidth(leftW), bodyHeight), leftW, panelH, leftActive)
		right = panel("Network", m.networkDetail(), rightW, panelH, rightActive)
	default:
		left = panel(m.listTitle(), m.listBody(panelContentWidth(leftW), bodyHeight), leftW, panelH, leftActive)
		right = panel(m.rightTitle(), m.rightBody(), rightW, panelH, rightActive)
	}

	panelRow := lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", m.panelGap()), right)
	view := m.tabStrip() + strings.Repeat("\n", 1+m.tabPanelGap()) +
		panelRow + "\n" +
		m.fittedBottomBar()

	if m.helpOpen {
		box := m.helpBox()
		width, height := m.contentSize()
		x := max(0, (width-lipgloss.Width(box))/2)
		y := max(0, (height-lipgloss.Height(box))/2)
		view = overlay(view, box, x, y)
	}
	return m.insetView(view)
}
