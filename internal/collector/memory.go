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

type MemoryCollector struct {
	MeminfoPath string
}

func (c MemoryCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if c.MeminfoPath == "" {
		if metrics, err := platformMemoryMetrics(); metrics != nil || err == nil {
			return metrics, err
		}
	}

	path := c.MeminfoPath
	if path == "" {
		path = "/proc/meminfo"
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("collect memory from %s: %w", path, err)
	}
	defer file.Close()

	return parseMeminfo(file)
}

func parseMeminfo(r io.Reader) ([]Metric, error) {
	values := map[string]float64{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return nil, fmt.Errorf("parse meminfo %s: %w", key, err)
		}
		values[key] = value * 1024
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	mapping := map[string]string{
		"MemTotal":     "gosysmon_memory_total_bytes",
		"MemAvailable": "gosysmon_memory_available_bytes",
		"MemFree":      "gosysmon_memory_free_bytes",
		"SwapTotal":    "gosysmon_swap_total_bytes",
		"SwapFree":     "gosysmon_swap_free_bytes",
	}

	metrics := make([]Metric, 0, len(mapping)+1)
	for key, name := range mapping {
		value, ok := values[key]
		if !ok {
			continue
		}
		metrics = append(metrics, Metric{
			Name:  name,
			Help:  memoryHelp(name),
			Type:  Gauge,
			Value: value,
		})
	}

	total, hasTotal := values["MemTotal"]
	available, hasAvailable := values["MemAvailable"]
	if hasTotal && hasAvailable && total > 0 {
		metrics = append(metrics, Metric{
			Name:  "gosysmon_memory_used_ratio",
			Help:  "Memory usage ratio based on MemAvailable.",
			Type:  Gauge,
			Value: (total - available) / total,
		})
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("no supported meminfo fields found")
	}
	return metrics, nil
}

func memoryHelp(name string) string {
	switch name {
	case "gosysmon_memory_total_bytes":
		return "Total physical memory in bytes."
	case "gosysmon_memory_available_bytes":
		return "Estimated physical memory available for new workloads in bytes."
	case "gosysmon_memory_free_bytes":
		return "Free physical memory in bytes."
	case "gosysmon_swap_total_bytes":
		return "Total swap memory in bytes."
	case "gosysmon_swap_free_bytes":
		return "Free swap memory in bytes."
	default:
		return "Memory metric in bytes."
	}
}
