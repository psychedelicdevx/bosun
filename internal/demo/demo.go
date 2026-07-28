package demo

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/psychedelicdevx/bosun/internal/docker"
)

type Engine struct {
	mu         sync.Mutex
	containers []docker.Container
}

func New() *Engine {
	return &Engine{containers: []docker.Container{
		{ID: "a1b2c3d4e5f6", Name: "web", Image: "nginx:1.27", State: "running", Status: "Up 3 days"},
		{ID: "b2c3d4e5f6a1", Name: "api", Image: "ghcr.io/acme/api:latest", State: "running", Status: "Up 3 days (healthy)"},
		{ID: "c3d4e5f6a1b2", Name: "postgres", Image: "postgres:16", State: "running", Status: "Up 3 days"},
		{ID: "d4e5f6a1b2c3", Name: "redis", Image: "redis:7", State: "running", Status: "Up 3 days"},
		{ID: "e5f6a1b2c3d4", Name: "worker", Image: "ghcr.io/acme/worker:latest", State: "running", Status: "Up 2 days"},
		{ID: "f6a1b2c3d4e5", Name: "mailpit", Image: "axllent/mailpit", State: "running", Status: "Up 3 days"},
		{ID: "0a1b2c3d4e5f", Name: "migrate", Image: "ghcr.io/acme/migrate:1.4", State: "exited", Status: "Exited (0) 3 days ago"},
		{ID: "1b2c3d4e5f60", Name: "seed", Image: "ghcr.io/acme/seed:1.4", State: "exited", Status: "Exited (1) 3 days ago"},
	}}
}

func (e *Engine) List(ctx context.Context) ([]docker.Container, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]docker.Container(nil), e.containers...), nil
}

var demoLog = []string{
	`GET  /healthz          200   1ms`,
	`GET  /api/orders       200  14ms`,
	`POST /api/orders       201  38ms`,
	`level=info msg="job picked up" queue=default id=8841`,
	`GET  /api/users/42     200   9ms`,
	`level=warn msg="retrying upstream" attempt=2`,
	`GET  /assets/app.css   304   0ms`,
	`level=info msg="job done" id=8841 dur=112ms`,
	`POST /api/webhooks     202   5ms`,
	`GET  /api/orders?page=2 200  17ms`,
}

func (e *Engine) Logs(ctx context.Context, id string) (<-chan string, error) {
	out := make(chan string)
	go func() {
		defer close(out)
		t := time.NewTicker(350 * time.Millisecond)
		defer t.Stop()
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
			line := fmt.Sprintf("%s  %s", stamp(i), demoLog[i%len(demoLog)])
			select {
			case <-ctx.Done():
				return
			case out <- line:
			}
		}
	}()
	return out, nil
}

func (e *Engine) Stats(ctx context.Context, id string) (<-chan docker.Stats, error) {
	out := make(chan docker.Stats)
	go func() {
		defer close(out)
		t := time.NewTicker(time.Second)
		defer t.Stop()
		const limit = 2 * 1024 * 1024 * 1024
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
			cpu := 3 + 6*math.Abs(math.Sin(float64(i)/3))
			mem := uint64(46*1024*1024) + uint64(i%7)*1024*1024
			s := docker.Stats{
				CPUPercent: cpu,
				MemUsage:   mem,
				MemLimit:   limit,
				MemPercent: float64(mem) / limit * 100,
			}
			select {
			case <-ctx.Done():
				return
			case out <- s:
			}
		}
	}()
	return out, nil
}

func (e *Engine) Start(ctx context.Context, id string) error {
	return e.setState(id, "running", "Up 1 second")
}

func (e *Engine) Stop(ctx context.Context, id string) error {
	return e.setState(id, "exited", "Exited (0) 1 second ago")
}

func (e *Engine) Restart(ctx context.Context, id string) error {
	return e.setState(id, "running", "Up 1 second")
}

func (e *Engine) Remove(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := e.containers[:0]
	for _, c := range e.containers {
		if c.ID != id {
			out = append(out, c)
		}
	}
	e.containers = out
	return nil
}

func (e *Engine) setState(id, state, status string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := range e.containers {
		if e.containers[i].ID == id {
			e.containers[i].State = state
			e.containers[i].Status = status
		}
	}
	return nil
}

func stamp(i int) string {
	return fmt.Sprintf("10:%02d:%02d", 20+i/60, i%60)
}
