package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/psychedelicdevx/bosun/internal/demo"
	"github.com/psychedelicdevx/bosun/internal/docker"
	"github.com/psychedelicdevx/bosun/internal/ui"
)

var version = "dev"

func main() {
	if hasFlag("--version") || hasFlag("-v") {
		fmt.Println("bosun", version)
		return
	}
	if hasFlag("--themes") {
		for _, n := range ui.ThemeNames {
			fmt.Println(n)
		}
		return
	}

	ui.ApplyTheme(flagValue("--theme", os.Getenv("BOSUN_THEME")))

	var engine ui.Engine
	if hasFlag("--demo") || os.Getenv("BOSUN_DEMO") != "" {
		engine = demo.New()
	} else {
		client, err := connectDocker()
		if err != nil {
			fmt.Fprintf(os.Stderr, "bosun: cannot connect to docker: %v\n", err)
			os.Exit(1)
		}
		engine = client
	}

	p := tea.NewProgram(ui.New(engine), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "bosun: %v\n", err)
		os.Exit(1)
	}
}

func connectDocker() (*docker.Client, error) {
	if h := flagValue("--host", ""); h != "" {
		return docker.NewWithHost(h)
	}
	if os.Getenv("DOCKER_HOST") != "" {
		return docker.New()
	}
	if h := docker.ContextHost(); h != "" {
		if strings.HasPrefix(h, "ssh://") {
			fmt.Fprintf(os.Stderr, "bosun: docker context points at %s; ssh endpoints are not supported yet, using the default socket. set DOCKER_HOST to a tcp:// endpoint for a remote daemon.\n", h)
			return docker.New()
		}
		return docker.NewWithHost(h)
	}
	return docker.New()
}

func hasFlag(f string) bool {
	for _, a := range os.Args[1:] {
		if a == f {
			return true
		}
	}
	return false
}

func flagValue(name, def string) string {
	args := os.Args[1:]
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, name+"=") {
			return strings.TrimPrefix(a, name+"=")
		}
	}
	return def
}
