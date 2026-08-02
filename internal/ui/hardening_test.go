package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

type refreshTestEngine struct {
	entered chan struct{}
	release chan struct{}
}

func (e *refreshTestEngine) probe(ctx context.Context) error {
	e.entered <- struct{}{}
	select {
	case <-e.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *refreshTestEngine) List(ctx context.Context) ([]docker.Container, error) {
	return []docker.Container{{Name: "web"}}, e.probe(ctx)
}
func (e *refreshTestEngine) Images(ctx context.Context) ([]docker.Image, error) {
	return []docker.Image{{Repo: "image"}}, e.probe(ctx)
}
func (e *refreshTestEngine) Volumes(ctx context.Context) ([]docker.Volume, error) {
	return []docker.Volume{{Name: "data"}}, e.probe(ctx)
}
func (e *refreshTestEngine) Networks(ctx context.Context) ([]docker.Network, error) {
	return []docker.Network{{Name: "bridge"}}, e.probe(ctx)
}
func (*refreshTestEngine) Logs(context.Context, string) (<-chan string, error) { return nil, nil }
func (*refreshTestEngine) Stats(context.Context, string) (<-chan docker.Stats, error) {
	return nil, nil
}
func (*refreshTestEngine) Start(context.Context, string) error         { return nil }
func (*refreshTestEngine) Stop(context.Context, string) error          { return nil }
func (*refreshTestEngine) Restart(context.Context, string) error       { return nil }
func (*refreshTestEngine) Remove(context.Context, string) error        { return nil }
func (*refreshTestEngine) RemoveImage(context.Context, string) error   { return nil }
func (*refreshTestEngine) PruneImages(context.Context) (uint64, error) { return 0, nil }
func (*refreshTestEngine) RemoveVolume(context.Context, string) error  { return nil }
func (*refreshTestEngine) RemoveNetwork(context.Context, string) error { return nil }

func TestListWindowFollowsCursor(t *testing.T) {
	containers := make([]docker.Container, 10)
	for i := range containers {
		containers[i] = docker.Container{Name: fmt.Sprintf("container-%d", i)}
	}
	m := New(nil)
	m.containers = containers
	m.loaded = true
	m.cursor = 9

	body := ansi.Strip(m.listBody(30, 3))
	if !strings.Contains(body, "container-9") {
		t.Fatalf("selected row is not visible: %q", body)
	}
	if strings.Contains(body, "container-0") {
		t.Fatalf("window should scroll away from the first row: %q", body)
	}
}

func TestWindowRange(t *testing.T) {
	tests := []struct {
		total, cursor, height int
		start, end            int
	}{
		{20, 0, 5, 0, 5},
		{20, 6, 5, 2, 7},
		{20, 19, 5, 15, 20},
		{3, 2, 10, 0, 3},
	}
	for _, tt := range tests {
		start, end := windowRange(tt.total, tt.cursor, tt.height)
		if start != tt.start || end != tt.end {
			t.Fatalf("windowRange(%d, %d, %d) = (%d, %d), want (%d, %d)", tt.total, tt.cursor, tt.height, start, end, tt.start, tt.end)
		}
	}
}

func TestRefreshRetainsStaleDataAndRecovers(t *testing.T) {
	m := New(nil)
	m.loaded = true
	m.err = errors.New("old failure")
	m.images = []docker.Image{{Repo: "cached"}}

	m.applyRefresh(refreshMsg{
		containers: []docker.Container{{Name: "web"}},
		imagesErr:  errors.New("image refresh failed"),
		volumes:    []docker.Volume{{Name: "data"}},
		networks:   []docker.Network{{Name: "bridge"}},
	})

	if m.err != nil || m.containerErr != nil || len(m.containers) != 1 {
		t.Fatalf("container recovery failed: err=%v containerErr=%v containers=%v", m.err, m.containerErr, m.containers)
	}
	if len(m.images) != 1 || m.images[0].Repo != "cached" || m.imagesErr == nil {
		t.Fatalf("failed image refresh should retain cached data: images=%v err=%v", m.images, m.imagesErr)
	}
	if !strings.Contains(ansi.Strip(m.imageListBody(30, 5)), "showing cached data") {
		t.Fatal("stale image data should be labeled")
	}
}

func TestResourceLoadingAndFailureAreDistinct(t *testing.T) {
	m := New(nil)
	if got := ansi.Strip(m.volumeListBody(30, 5)); got != "loading..." {
		t.Fatalf("initial volume state = %q, want loading", got)
	}
	m.volumesLoaded = true
	m.volumesErr = errors.New("daemon unavailable")
	got := ansi.Strip(m.volumeListBody(30, 5))
	if !strings.Contains(got, "refresh failed") || !strings.Contains(got, "daemon unavailable") {
		t.Fatalf("failed volume state is not explicit: %q", got)
	}
}

func TestRefreshGuardPreventsOverlap(t *testing.T) {
	m := New(nil)
	if cmd := m.beginRefresh(); cmd != nil {
		t.Fatal("refresh already in flight should not start another command")
	}
	m.refreshing = false
	if cmd := m.beginRefresh(); cmd == nil || !m.refreshing {
		t.Fatal("idle model should start exactly one refresh")
	}
}

func TestRefreshAllRunsBoundedRequestsTogether(t *testing.T) {
	engine := &refreshTestEngine{
		entered: make(chan struct{}, 4),
		release: make(chan struct{}),
	}
	m := New(engine)
	done := make(chan refreshMsg, 1)
	go func() {
		done <- m.refreshAll().(refreshMsg)
	}()

	for i := 0; i < 4; i++ {
		select {
		case <-engine.entered:
		case <-time.After(time.Second):
			t.Fatal("refresh requests did not start concurrently")
		}
	}
	close(engine.release)

	select {
	case result := <-done:
		if result.containerErr != nil || result.imagesErr != nil || result.volumesErr != nil || result.networksErr != nil {
			t.Fatalf("refresh returned errors: %+v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("refresh did not complete after requests were released")
	}
}

func TestRequestedRefreshRunsAfterInflightRefresh(t *testing.T) {
	m := New(nil)
	if cmd := m.requestRefresh(); cmd != nil || !m.refreshPending {
		t.Fatalf("refresh request should queue behind inflight work: cmd=%v pending=%v", cmd, m.refreshPending)
	}
	m.applyRefresh(refreshMsg{})
	m.refreshPending = false
	if cmd := m.beginRefresh(); cmd == nil {
		t.Fatal("queued refresh should be able to start after completion")
	}
}

func TestFocusOnlyMovesToInteractiveRightViews(t *testing.T) {
	m := New(nil)
	m = send(m, key("tab"))
	if m.focusRight {
		t.Fatal("details panel should not claim keyboard focus")
	}
	m.right = viewLogs
	m = send(m, key("tab"))
	if !m.focusRight {
		t.Fatal("logs panel should accept keyboard focus")
	}
}

func TestMouseMomentumStaysWithLogPaneAfterFocusSwitch(t *testing.T) {
	m := New(nil)
	m.width = 120
	m.height = 40
	m.loaded = true
	m.containers = []docker.Container{
		{Name: "one"}, {Name: "two"}, {Name: "three"}, {Name: "four"}, {Name: "five"},
	}
	m.cursor = 3
	m.right = viewLogs
	m.focusRight = false
	m.vp.Width = 40
	m.vp.Height = 3
	m.vp.SetContent(strings.Repeat("line\n", 20))
	m.vp.GotoBottom()
	beforeOffset := m.vp.YOffset

	m = send(m, tea.MouseMsg{
		X:      100,
		Y:      10,
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	})

	if m.cursor != 3 {
		t.Fatalf("right-pane wheel moved container cursor: got %d, want 3", m.cursor)
	}
	if m.vp.YOffset >= beforeOffset {
		t.Fatalf("right-pane wheel did not scroll logs: before=%d after=%d", beforeOffset, m.vp.YOffset)
	}
}

func TestMouseWheelOnlyMovesListWhenPointerIsOverLeftPane(t *testing.T) {
	m := New(nil)
	m.width = 120
	m.height = 40
	m.loaded = true
	m.containers = []docker.Container{{Name: "one"}, {Name: "two"}, {Name: "three"}}

	m = send(m, tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	if m.cursor != 1 {
		t.Fatalf("left-pane wheel cursor = %d, want 1", m.cursor)
	}

	m.right = viewDetails
	m = send(m, tea.MouseMsg{
		X:      100,
		Y:      10,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	if m.cursor != 1 {
		t.Fatalf("right-pane momentum leaked into list: got %d, want 1", m.cursor)
	}
}

func TestMouseClicksTabsRowsAndInteractivePane(t *testing.T) {
	m := New(nil)
	m.width = 120
	m.height = 40
	m.loaded = true
	m.containers = []docker.Container{{Name: "one"}, {Name: "two"}, {Name: "three"}}
	m.images = []docker.Image{{Repo: "first"}, {Repo: "second"}}
	m.imagesLoaded = true

	// Images begins after the leading space, Containers, and the five-cell separator.
	m = send(m, tea.MouseMsg{X: 17, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if m.tab != tabImages {
		t.Fatalf("clicking Images tab selected %v", m.tab)
	}

	// At this size the panel begins at y=3 and its first body row is y=4.
	m = send(m, tea.MouseMsg{X: 10, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if m.imgCursor != 1 {
		t.Fatalf("clicking second image row selected %d", m.imgCursor)
	}

	m.tab = tabContainers
	m.right = viewLogs
	m.focusRight = false
	m = send(m, tea.MouseMsg{X: 100, Y: 10, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if !m.focusRight {
		t.Fatal("clicking logs pane should focus it")
	}

	m = send(m, tea.MouseMsg{X: 10, Y: 6, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if m.cursor != 2 || m.focusRight || m.right != viewDetails {
		t.Fatalf("clicking third container row did not select it cleanly: cursor=%d focusRight=%v right=%v", m.cursor, m.focusRight, m.right)
	}
}

func TestStateGlyphsDoNotRelyOnColor(t *testing.T) {
	if got := stateGlyph(docker.Container{State: "running"}); got != "●" {
		t.Fatalf("running glyph = %q", got)
	}
	if got := stateGlyph(docker.Container{State: "exited", Status: "Exited (0)"}); got != "○" {
		t.Fatalf("clean stop glyph = %q", got)
	}
	if got := stateGlyph(docker.Container{State: "exited", Status: "Exited (1)"}); got != "×" {
		t.Fatalf("failed glyph = %q", got)
	}
}

func TestClosedStreamsExposeReconnectState(t *testing.T) {
	m := New(nil)
	m.right = viewLogs
	m.focusRight = true
	m.logGen = 4
	m = send(m, logClosedMsg{gen: 4})
	if !m.logEnded || !strings.Contains(m.status, "reconnects") {
		t.Fatalf("closed logs should expose reconnect state: ended=%v status=%q", m.logEnded, m.status)
	}

	m.right = viewStats
	m.statsGen = 2
	m = send(m, statsClosedMsg{gen: 2})
	if !m.statsEnded || !strings.Contains(m.status, "reconnects") {
		t.Fatalf("closed stats should expose reconnect state: ended=%v status=%q", m.statsEnded, m.status)
	}
}

func TestUnicodeBackspaceRemovesWholeRune(t *testing.T) {
	m := New(nil)
	m.filtering = true
	m.filter = "café"
	m = send(m, tea.KeyMsg{Type: tea.KeyBackspace})
	if m.filter != "caf" {
		t.Fatalf("container filter after backspace = %q", m.filter)
	}

	m.filtering = false
	m.logFiltering = true
	m.logFilter = "İş"
	m = send(m, tea.KeyMsg{Type: tea.KeyBackspace})
	if m.logFilter != "İ" {
		t.Fatalf("log filter after backspace = %q", m.logFilter)
	}
}

func TestPruneRequiresConfirmation(t *testing.T) {
	m := New(nil)
	m.tab = tabImages
	m = send(m, key("p"))
	if !m.confirming || m.pendingKind != "prune" {
		t.Fatalf("prune should require confirmation: confirming=%v kind=%q", m.confirming, m.pendingKind)
	}
}

func TestWaitForLogBatchesAvailableLines(t *testing.T) {
	ch := make(chan string, 3)
	ch <- "one"
	ch <- "two"
	ch <- "three"
	close(ch)

	msg := waitForLog(7, ch)()
	batch, ok := msg.(logLinesMsg)
	if !ok || batch.gen != 7 || !batch.closed || len(batch.lines) != 3 {
		t.Fatalf("unexpected log batch: %#v", msg)
	}
}
