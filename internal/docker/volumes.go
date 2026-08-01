package docker

import (
	"context"

	"github.com/docker/docker/api/types/volume"
)

type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
}

func (c *Client) Volumes(ctx context.Context) ([]Volume, error) {
	resp, err := c.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]Volume, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		out = append(out, Volume{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
		})
	}
	return out, nil
}

func (c *Client) RemoveVolume(ctx context.Context, name string) error {
	return c.cli.VolumeRemove(ctx, name, false)
}
