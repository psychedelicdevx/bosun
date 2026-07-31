package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
)

type Image struct {
	ID       string
	Repo     string
	Size     int64
	Created  int64
	Dangling bool
}

func (c *Client) Images(ctx context.Context) ([]Image, error) {
	list, err := c.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]Image, 0, len(list))
	for _, im := range list {
		repo := "<none>:<none>"
		if len(im.RepoTags) > 0 {
			repo = im.RepoTags[0]
		}
		id := strings.TrimPrefix(im.ID, "sha256:")
		if len(id) > 12 {
			id = id[:12]
		}
		out = append(out, Image{
			ID:       id,
			Repo:     repo,
			Size:     im.Size,
			Created:  im.Created,
			Dangling: len(im.RepoTags) == 0 || repo == "<none>:<none>",
		})
	}
	return out, nil
}

func (c *Client) RemoveImage(ctx context.Context, id string) error {
	_, err := c.cli.ImageRemove(ctx, id, image.RemoveOptions{})
	return err
}

func (c *Client) PruneImages(ctx context.Context) (uint64, error) {
	rep, err := c.cli.ImagesPrune(ctx, filters.NewArgs(filters.Arg("dangling", "true")))
	if err != nil {
		return 0, err
	}
	return rep.SpaceReclaimed, nil
}
