package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

type tab int

const (
	tabContainers tab = iota
	tabImages
)

type imagesMsg []docker.Image

type imageDoneMsg struct {
	verb      string
	name      string
	reclaimed uint64
	err       error
}

func (m Model) fetchImages() tea.Msg {
	list, err := m.client.Images(context.Background())
	if err != nil {
		return imagesMsg(nil)
	}
	return imagesMsg(list)
}

func (m Model) removeImage(id, name string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		err := client.RemoveImage(context.Background(), id)
		return imageDoneMsg{verb: "removed image", name: name, err: err}
	}
}

func (m Model) pruneImages() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		n, err := client.PruneImages(context.Background())
		return imageDoneMsg{verb: "pruned", reclaimed: n, err: err}
	}
}

func (m Model) updateImages(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		if m.imgCursor < len(m.images)-1 {
			m.imgCursor++
		}
	case "up", "k":
		if m.imgCursor > 0 {
			m.imgCursor--
		}
	case "d":
		if m.imgCursor >= len(m.images) {
			return m, nil
		}
		im := m.images[m.imgCursor]
		m.confirming = true
		m.pendingImage = true
		m.pendingID = im.ID
		m.pendingName = im.Repo
	case "p":
		return m, tea.Batch(m.setStatus("pruning dangling images..."), m.pruneImages())
	}
	return m, nil
}

func (m Model) tabStrip() string {
	name := func(s string, active bool) string {
		if active {
			return borderActiveStyle.Render(s)
		}
		return hintStyle.Render(s)
	}
	return " " + name("Containers", m.tab == tabContainers) +
		hintStyle.Render("  ·  ") + name("Images", m.tab == tabImages)
}

func (m Model) imageListBody(inner int) string {
	if len(m.images) == 0 {
		return hintStyle.Render("no images")
	}
	var b strings.Builder
	for i, im := range m.images {
		size := humanBytes(uint64(im.Size))
		repo := im.Repo
		gap := inner - lipgloss.Width(repo) - lipgloss.Width(size)
		if gap < 1 {
			repo = fit(repo, max(1, inner-lipgloss.Width(size)-1))
			gap = 1
		}
		switch {
		case i == m.imgCursor:
			b.WriteString(selRowStyle.Render(repo + strings.Repeat(" ", gap) + size))
		case im.Dangling:
			b.WriteString(stoppedStyle.Render(repo + strings.Repeat(" ", gap) + size))
		default:
			b.WriteString(repo + strings.Repeat(" ", gap) + hintStyle.Render(size))
		}
		if i < len(m.images)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m Model) imageDetail() string {
	if m.imgCursor >= len(m.images) {
		return hintStyle.Render("no image selected")
	}
	im := m.images[m.imgCursor]
	dangling := "no"
	if im.Dangling {
		dangling = "yes (safe to prune)"
	}
	return fmt.Sprintf(
		"%s\n\n%s  %s\n%s  %s\n%s  %s\n%s  %s",
		headerStyle.Render(im.Repo),
		labelStyle.Render("id      "), im.ID,
		labelStyle.Render("size    "), humanBytes(uint64(im.Size)),
		labelStyle.Render("created "), humanAge(im.Created),
		labelStyle.Render("dangling"), dangling,
	)
}

func humanAge(created int64) string {
	if created == 0 {
		return "unknown"
	}
	d := time.Since(time.Unix(created, 0))
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}
