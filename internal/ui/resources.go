package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

type volumesMsg []docker.Volume
type networksMsg []docker.Network

type resourceDoneMsg struct {
	verb string
	name string
	err  error
}

func (m Model) fetchVolumes() tea.Msg {
	list, err := m.client.Volumes(context.Background())
	if err != nil {
		return volumesMsg(nil)
	}
	return volumesMsg(list)
}

func (m Model) fetchNetworks() tea.Msg {
	list, err := m.client.Networks(context.Background())
	if err != nil {
		return networksMsg(nil)
	}
	return networksMsg(list)
}

func (m Model) removeVolume(name string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		err := client.RemoveVolume(context.Background(), name)
		return resourceDoneMsg{verb: "removed volume", name: name, err: err}
	}
}

func (m Model) removeNetwork(id, name string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		err := client.RemoveNetwork(context.Background(), id)
		return resourceDoneMsg{verb: "removed network", name: name, err: err}
	}
}

func (m Model) updateVolumes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		if m.volCursor < len(m.volumes)-1 {
			m.volCursor++
		}
	case "up", "k":
		if m.volCursor > 0 {
			m.volCursor--
		}
	case "d":
		if m.volCursor >= len(m.volumes) {
			return m, nil
		}
		v := m.volumes[m.volCursor]
		m.confirming = true
		m.pendingKind = "volume"
		m.pendingName = v.Name
	}
	return m, nil
}

func (m Model) updateNetworks(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		if m.netCursor < len(m.networks)-1 {
			m.netCursor++
		}
	case "up", "k":
		if m.netCursor > 0 {
			m.netCursor--
		}
	case "d":
		if m.netCursor >= len(m.networks) {
			return m, nil
		}
		n := m.networks[m.netCursor]
		m.confirming = true
		m.pendingKind = "network"
		m.pendingID = n.ID
		m.pendingName = n.Name
	}
	return m, nil
}

func (m Model) volumeListBody(inner int) string {
	if len(m.volumes) == 0 {
		return hintStyle.Render("no volumes")
	}
	var b strings.Builder
	for i, v := range m.volumes {
		line := v.Name
		if i == m.volCursor {
			if pad := inner - len([]rune(line)); pad > 0 {
				line += strings.Repeat(" ", pad)
			}
			b.WriteString(selRowStyle.Render(line))
		} else {
			b.WriteString(fit(line, inner))
		}
		if i < len(m.volumes)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m Model) networkListBody(inner int) string {
	if len(m.networks) == 0 {
		return hintStyle.Render("no networks")
	}
	var b strings.Builder
	for i, n := range m.networks {
		line := n.Name
		if i == m.netCursor {
			if pad := inner - len([]rune(line)); pad > 0 {
				line += strings.Repeat(" ", pad)
			}
			b.WriteString(selRowStyle.Render(line))
		} else {
			b.WriteString(fit(line, inner))
		}
		if i < len(m.networks)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m Model) volumeDetail() string {
	if m.volCursor >= len(m.volumes) {
		return hintStyle.Render("no volume selected")
	}
	v := m.volumes[m.volCursor]
	return fmt.Sprintf(
		"%s\n\n%s  %s\n%s  %s",
		headerStyle.Render(v.Name),
		labelStyle.Render("driver    "), v.Driver,
		labelStyle.Render("mountpoint"), v.Mountpoint,
	)
}

func (m Model) networkDetail() string {
	if m.netCursor >= len(m.networks) {
		return hintStyle.Render("no network selected")
	}
	n := m.networks[m.netCursor]
	subnet := n.Subnet
	if subnet == "" {
		subnet = "-"
	}
	gateway := n.Gateway
	if gateway == "" {
		gateway = "-"
	}
	return fmt.Sprintf(
		"%s\n\n%s  %s\n%s  %s\n%s  %s\n%s  %s\n%s  %s",
		headerStyle.Render(n.Name),
		labelStyle.Render("id     "), n.ID,
		labelStyle.Render("driver "), n.Driver,
		labelStyle.Render("scope  "), n.Scope,
		labelStyle.Render("subnet "), subnet,
		labelStyle.Render("gateway"), gateway,
	)
}
