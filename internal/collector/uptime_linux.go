//go:build linux

package collector

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type UptimeCollector struct {
	UptimePath string
}

func (c UptimeCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path := c.UptimePath
	if path == "" {
		path = "/proc/uptime"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read uptime from %s: %w", path, err)
	}
	seconds, err := parseUptime(data)
	if err != nil {
		return nil, err
	}
	return uptimeMetric(seconds), nil
}

func parseUptime(data []byte) (float64, error) {
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, fmt.Errorf("missing uptime value")
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse uptime: %w", err)
	}
	return seconds, nil
}
