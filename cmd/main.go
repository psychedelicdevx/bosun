package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/psychedelicdevx/bosun/internal/docker"
	"github.com/psychedelicdevx/bosun/internal/ui"
)

func main() {
	client, err := docker.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bosun: cannot connect to docker: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "bosun: %v\n", err)
		os.Exit(1)
	}
}
