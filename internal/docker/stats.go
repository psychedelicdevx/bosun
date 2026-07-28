package docker

import (
	"context"
	"encoding/json"

	"github.com/docker/docker/api/types/container"
)

type Stats struct {
	CPUPercent float64
	MemUsage   uint64
	MemLimit   uint64
	MemPercent float64
}

func (c *Client) Stats(ctx context.Context, id string) (<-chan Stats, error) {
	resp, err := c.cli.ContainerStats(ctx, id, true)
	if err != nil {
		return nil, err
	}

	out := make(chan Stats)
	go func() {
		defer resp.Body.Close()
		defer close(out)

		dec := json.NewDecoder(resp.Body)
		for {
			var v container.StatsResponse
			if err := dec.Decode(&v); err != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case out <- compute(v):
			}
		}
	}()
	return out, nil
}

func compute(v container.StatsResponse) Stats {
	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(v.CPUStats.SystemUsage) - float64(v.PreCPUStats.SystemUsage)
	cpus := float64(v.CPUStats.OnlineCPUs)
	if cpus == 0 {
		cpus = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	}

	cpuPct := 0.0
	if sysDelta > 0 && cpuDelta > 0 {
		cpuPct = (cpuDelta / sysDelta) * cpus * 100
	}

	mem := v.MemoryStats.Usage
	if cache, ok := v.MemoryStats.Stats["inactive_file"]; ok && cache < mem {
		mem -= cache
	}
	memPct := 0.0
	if v.MemoryStats.Limit > 0 {
		memPct = float64(mem) / float64(v.MemoryStats.Limit) * 100
	}

	return Stats{CPUPercent: cpuPct, MemUsage: mem, MemLimit: v.MemoryStats.Limit, MemPercent: memPct}
}
