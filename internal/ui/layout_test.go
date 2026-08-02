package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestInsetsAreAdaptive(t *testing.T) {
	large := Model{width: 100, height: 30}
	if horizontal, vertical := large.insets(); horizontal != 1 || vertical != 1 {
		t.Fatalf("large terminal insets = (%d, %d), want (1, 1)", horizontal, vertical)
	}
	if width, height := large.contentSize(); width != 98 || height != 28 {
		t.Fatalf("large terminal content = (%d, %d), want (98, 28)", width, height)
	}

	small := Model{width: 60, height: 14}
	if horizontal, vertical := small.insets(); horizontal != 0 || vertical != 0 {
		t.Fatalf("small terminal insets = (%d, %d), want (0, 0)", horizontal, vertical)
	}
}

func TestSpacingIsAdaptive(t *testing.T) {
	large := Model{width: 100, height: 30}
	if got := large.panelGap(); got != 1 {
		t.Fatalf("large terminal panel gap = %d, want 1", got)
	}
	if got := large.tabPanelGap(); got != 1 {
		t.Fatalf("large terminal tab gap = %d, want 1", got)
	}
	if got := panelContentWidth(40); got != 36 {
		t.Fatalf("40-column panel content width = %d, want 36", got)
	}

	small := Model{width: 60, height: 10}
	if got := small.panelGap(); got != 0 {
		t.Fatalf("small terminal panel gap = %d, want 0", got)
	}
	if got := small.tabPanelGap(); got != 0 {
		t.Fatalf("small terminal tab gap = %d, want 0", got)
	}
}

func TestBottomBarFitsContentWidth(t *testing.T) {
	m := Model{width: 60, height: 20, tab: tabContainers}
	if got, width := lipgloss.Width(m.fittedBottomBar()), 60; got > width {
		t.Fatalf("bottom bar width = %d, want <= %d", got, width)
	}
}
