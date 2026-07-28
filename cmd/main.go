package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/psychedelicdevx/bosun/internal/demo"
	"github.com/psychedelicdevx/bosun/internal/docker"
	"github.com/psychedelicdevx/bosun/internal/ui"
)

func main() {
	var engine ui.Engine
	if demoMode() {
		engine = demo.New()
	} else {
		client, err := docker.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "bosun: cannot connect to docker: %v\n", err)
			os.Exit(1)
		}
		engine = client
	}

	p := tea.NewProgram(ui.New(engine), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "bosun: %v\n", err)
		os.Exit(1)
	}
}

func demoMode() bool {
	if os.Getenv("BOSUN_DEMO") != "" {
		return true
	}
	for _, a := range os.Args[1:] {
		if a == "--demo" {
			return true
		}
	}
	return false
}
