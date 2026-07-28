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
		if m.cursor > len(m.containers)-1 {
			m.cursor = max(0, len(m.containers)-1)
		}
	case errMsg:
		m.err = msg.err
		m.loaded = true
	case tea.KeyMsg:
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
	b.WriteString("\n" + hintStyle.Render("↑/↓ move · q quit") + "\n")
	return b.String()
}
