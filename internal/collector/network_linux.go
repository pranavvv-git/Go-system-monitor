//go:build linux

package collector

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type NetworkCollector struct {
	DevPath string
}

func (c NetworkCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path := c.DevPath
	if path == "" {
		path = "/proc/net/dev"
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("collect network from %s: %w", path, err)
	}
	defer file.Close()
	return parseNetDev(file)
}

func parseNetDev(r io.Reader) ([]Metric, error) {
	scanner := bufio.NewScanner(r)
	var metrics []Metric
	line := 0
	for scanner.Scan() {
		line++
		if line <= 2 {
			continue
		}

		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}

		rx, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return nil, fmt.Errorf("parse rx bytes for %s: %w", name, err)
		}
		tx, err := strconv.ParseFloat(fields[8], 64)
		if err != nil {
			return nil, fmt.Errorf("parse tx bytes for %s: %w", name, err)
		}
		metrics = append(metrics, networkMetrics(name, rx, tx)...)
	}
	return metrics, scanner.Err()
}
