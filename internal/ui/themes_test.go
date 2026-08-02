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
	if got := keyStyle.GetForeground(); got != Themes["dracula"].Key {
		t.Fatalf("dracula key = %v, want %v", got, Themes["dracula"].Key)
	}
	if got := groupStyle.GetForeground(); got != Themes["dracula"].Group {
		t.Fatalf("dracula group = %v, want %v", got, Themes["dracula"].Group)
	}
	if got := warningStyle.GetForeground(); got != Themes["dracula"].Warning {
		t.Fatalf("dracula warning = %v, want %v", got, Themes["dracula"].Warning)
	}

	ApplyTheme("nonsense")
	if got := runningStyle.GetForeground(); got != Themes["default"].Running {
		t.Fatalf("unknown theme should fall back to default, got %v", got)
	}

	ApplyTheme("default")
}
