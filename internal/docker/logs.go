package docker

import (
	"bufio"
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

func (c *Client) Logs(ctx context.Context, id string) (<-chan string, error) {
	info, err := c.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	rc, err := c.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "200",
	})
	if err != nil {
		return nil, err
	}

	lines := make(chan string)
	go func() {
		defer rc.Close()
		defer close(lines)

		var r io.Reader = rc
		if info.Config == nil || !info.Config.Tty {
			pr, pw := io.Pipe()
			go func() {
				_, _ = stdcopy.StdCopy(pw, pw, rc)
				pw.Close()
			}()
			r = pr
		}

		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			select {
			case <-ctx.Done():
				return
			case lines <- sc.Text():
			}
		}
	}()

	return lines, nil
}
