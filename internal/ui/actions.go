package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type actionDoneMsg struct {
	verb string
	name string
	err  error
}

var pastTense = map[string]string{
	"start":   "started",
	"stop":    "stopped",
	"restart": "restarted",
	"remove":  "removed",
}

func (m Model) doAction(verb, id, name string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx := context.Background()
		var err error
		switch verb {
		case "start":
			err = client.Start(ctx, id)
		case "stop":
			err = client.Stop(ctx, id)
		case "restart":
			err = client.Restart(ctx, id)
		case "remove":
			err = client.Remove(ctx, id)
		}
		return actionDoneMsg{verb, name, err}
	}
}

func actionStatus(msg actionDoneMsg) string {
	if msg.err != nil {
		return fmt.Sprintf("%s %s failed: %v", msg.verb, msg.name, msg.err)
	}
	return fmt.Sprintf("%s %s", pastTense[msg.verb], msg.name)
}
