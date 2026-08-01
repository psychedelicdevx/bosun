package ui

import (
	"context"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

type Engine interface {
	List(ctx context.Context) ([]docker.Container, error)
	Logs(ctx context.Context, id string) (<-chan string, error)
	Stats(ctx context.Context, id string) (<-chan docker.Stats, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Restart(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error

	Images(ctx context.Context) ([]docker.Image, error)
	RemoveImage(ctx context.Context, id string) error
	PruneImages(ctx context.Context) (uint64, error)

	Volumes(ctx context.Context) ([]docker.Volume, error)
	RemoveVolume(ctx context.Context, name string) error

	Networks(ctx context.Context) ([]docker.Network, error)
	RemoveNetwork(ctx context.Context, id string) error
}
