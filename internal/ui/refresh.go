package ui

import (
	"context"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

const refreshTimeout = 5 * time.Second

type refreshMsg struct {
	containers   []docker.Container
	containerErr error
	images       []docker.Image
	imagesErr    error
	volumes      []docker.Volume
	volumesErr   error
	networks     []docker.Network
	networksErr  error
}

func (m Model) refreshAll() tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), refreshTimeout)
	defer cancel()

	var result refreshMsg
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		result.containers, result.containerErr = m.client.List(ctx)
	}()
	go func() {
		defer wg.Done()
		result.images, result.imagesErr = m.client.Images(ctx)
	}()
	go func() {
		defer wg.Done()
		result.volumes, result.volumesErr = m.client.Volumes(ctx)
	}()
	go func() {
		defer wg.Done()
		result.networks, result.networksErr = m.client.Networks(ctx)
	}()
	wg.Wait()
	return result
}

func (m *Model) beginRefresh() tea.Cmd {
	if m.refreshing {
		return nil
	}
	m.refreshing = true
	return m.refreshAll
}

func (m *Model) requestRefresh() tea.Cmd {
	if m.refreshing {
		m.refreshPending = true
		return nil
	}
	return m.beginRefresh()
}

func (m *Model) applyRefresh(msg refreshMsg) {
	m.refreshing = false

	if msg.containerErr != nil {
		m.containerErr = msg.containerErr
		if !m.loaded {
			m.err = msg.containerErr
		}
	} else {
		m.containers = msg.containers
		m.loaded = true
		m.err = nil
		m.containerErr = nil
		m.clampCursor()
	}

	m.imagesLoaded = true
	if msg.imagesErr != nil {
		m.imagesErr = msg.imagesErr
	} else {
		m.images = msg.images
		m.imagesErr = nil
		m.imgCursor = clampIndex(m.imgCursor, len(m.images))
	}

	m.volumesLoaded = true
	if msg.volumesErr != nil {
		m.volumesErr = msg.volumesErr
	} else {
		m.volumes = msg.volumes
		m.volumesErr = nil
		m.volCursor = clampIndex(m.volCursor, len(m.volumes))
	}

	m.networksLoaded = true
	if msg.networksErr != nil {
		m.networksErr = msg.networksErr
	} else {
		m.networks = msg.networks
		m.networksErr = nil
		m.netCursor = clampIndex(m.netCursor, len(m.networks))
	}
}

func clampIndex(cursor, length int) int {
	if length == 0 {
		return 0
	}
	return min(max(0, cursor), length-1)
}
