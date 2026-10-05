//go:build linux

package collector

import (
	"context"
	"fmt"
	"syscall"
)

type DiskCollector struct {
	Paths []string
}

func (c DiskCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	paths := c.Paths
	if len(paths) == 0 {
		paths = []string{"/"}
	}

	var metrics []Metric
	for _, path := range paths {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(path, &stat); err != nil {
			return metrics, fmt.Errorf("statfs %s: %w", path, err)
		}

		size := float64(stat.Blocks) * float64(stat.Bsize)
		free := float64(stat.Bfree) * float64(stat.Bsize)
		available := float64(stat.Bavail) * float64(stat.Bsize)

		labels := map[string]string{"path": path}
		metrics = append(metrics,
			Metric{Name: "gosysmon_disk_size_bytes", Help: "Total filesystem size in bytes.", Type: Gauge, Labels: labels, Value: size},
			Metric{Name: "gosysmon_disk_free_bytes", Help: "Free filesystem bytes, including reserved blocks.", Type: Gauge, Labels: labels, Value: free},
			Metric{Name: "gosysmon_disk_available_bytes", Help: "Free filesystem bytes available to unprivileged users.", Type: Gauge, Labels: labels, Value: available},
		)

		if size > 0 {
			metrics = append(metrics, Metric{Name: "gosysmon_disk_used_ratio", Help: "Filesystem usage ratio.", Type: Gauge, Labels: labels, Value: (size - free) / size})
		}
	}

	return metrics, nil
}
