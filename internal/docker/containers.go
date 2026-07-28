package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
)

type Container struct {
	ID     string
	Name   string
	Image  string
	State  string
	Status string
}

func (c *Client) List(ctx context.Context) ([]Container, error) {
	list, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	out := make([]Container, 0, len(list))
	for _, ct := range list {
		name := ""
		if len(ct.Names) > 0 {
			name = strings.TrimPrefix(ct.Names[0], "/")
		}
		id := ct.ID
		if len(id) > 12 {
			id = id[:12]
		}
		out = append(out, Container{
			ID:     id,
			Name:   name,
			Image:  ct.Image,
			State:  ct.State,
			Status: ct.Status,
		})
	}
	return out, nil
}
