package ui

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type execDoneMsg struct {
	name string
	err  error
}

func execShell(id, name string) tea.Cmd {
	c := exec.Command("docker", "exec", "-it", id, "sh")
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return execDoneMsg{name, err}
	})
}
