package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const actionTimeout = 30 * time.Second

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

var gerund = map[string]string{
	"start":   "starting",
	"stop":    "stopping",
	"restart": "restarting",
}

func runContainerAction(client Engine, verb, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), actionTimeout)
	defer cancel()
	switch verb {
	case "start":
		return client.Start(ctx, id)
	case "stop":
		return client.Stop(ctx, id)
	case "restart":
		return client.Restart(ctx, id)
	case "remove":
		return client.Remove(ctx, id)
	}
	return nil
}

func (m Model) doAction(verb, id, name string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		return actionDoneMsg{verb, name, runContainerAction(client, verb, id)}
	}
}

func (m Model) stackAction(verb, project string, ids []string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		var err error
		for _, id := range ids {
			if e := runContainerAction(client, verb, id); e != nil && err == nil {
				err = e
			}
		}
		return actionDoneMsg{verb, project + " stack", err}
	}
}

func actionStatus(msg actionDoneMsg) string {
	if msg.err != nil {
		return fmt.Sprintf("%s %s failed: %v", msg.verb, msg.name, msg.err)
	}
	return fmt.Sprintf("%s %s", pastTense[msg.verb], msg.name)
}
