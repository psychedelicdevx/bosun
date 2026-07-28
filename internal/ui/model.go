package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

var (
	runningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	stoppedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	erroredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	headerStyle  = lipgloss.NewStyle().Bold(true)
	hintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type Model struct {
	client     *docker.Client
	containers []docker.Container
	loaded     bool
	err        error
}

func New(client *docker.Client) Model {
	return Model{client: client}
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
	case containersMsg:
		m.containers = msg
		m.loaded = true
	case errMsg:
		m.err = msg.err
		m.loaded = true
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
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
	for _, ct := range m.containers {
		style := stoppedStyle
		switch ct.State {
		case "running":
			style = runningStyle
		case "exited", "dead":
			style = erroredStyle
		}
		b.WriteString(fmt.Sprintf("%s  %-24s  %s\n", style.Render("●"), ct.Name, ct.Status))
	}
	b.WriteString("\n" + hintStyle.Render("q quit") + "\n")
	return b.String()
}
