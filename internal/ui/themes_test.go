package ui

import "testing"

func TestApplyTheme(t *testing.T) {
	ApplyTheme("dracula")
	if got := runningStyle.GetForeground(); got != Themes["dracula"].Running {
		t.Fatalf("dracula running = %v, want %v", got, Themes["dracula"].Running)
	}
	if got := borderActive; got != Themes["dracula"].Accent {
		t.Fatalf("dracula accent = %v, want %v", got, Themes["dracula"].Accent)
	}

	ApplyTheme("nonsense")
	if got := runningStyle.GetForeground(); got != Themes["default"].Running {
		t.Fatalf("unknown theme should fall back to default, got %v", got)
	}

	ApplyTheme("default")
}
