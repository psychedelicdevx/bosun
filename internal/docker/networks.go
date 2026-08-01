package docker

import (
	"context"

	"github.com/docker/docker/api/types/network"
)

type Network struct {
	ID      string
	Name    string
	Driver  string
	Scope   string
	Subnet  string
	Gateway string
}

func (c *Client) Networks(ctx context.Context) ([]Network, error) {
	list, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]Network, 0, len(list))
	for _, n := range list {
		subnet, gateway := "", ""
		if len(n.IPAM.Config) > 0 {
			subnet = n.IPAM.Config[0].Subnet
			gateway = n.IPAM.Config[0].Gateway
		}
		id := n.ID
		if len(id) > 12 {
			id = id[:12]
		}
		out = append(out, Network{
			ID:      id,
			Name:    n.Name,
			Driver:  n.Driver,
			Scope:   n.Scope,
			Subnet:  subnet,
			Gateway: gateway,
		})
	}
	return out, nil
}

func (c *Client) RemoveNetwork(ctx context.Context, id string) error {
	return c.cli.NetworkRemove(ctx, id)
}
