package collector

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type CPUCollector struct {
	StatPath string
}

func (c CPUCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if c.StatPath == "" {
		metrics, err := platformCPUMetrics()
		if metrics != nil || err != nil {
			return appendLogicalCores(metrics), err
		}
	}

	path := c.StatPath
	if path == "" {
		path = "/proc/stat"
	}

	if runtime.GOOS == "linux" && c.StatPath == "" {
		first, err := readCPUSnapshot(path)
		if err != nil {
			return []Metric{logicalCoresMetric()}, fmt.Errorf("collect cpu from %s: %w", path, err)
		}
		timer := time.NewTimer(500 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
		second, err := readCPUSnapshot(path)
		if err != nil {
			return []Metric{logicalCoresMetric()}, fmt.Errorf("collect cpu from %s: %w", path, err)
		}
		return appendLogicalCores(cpuMetrics(second, &first)), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return []Metric{logicalCoresMetric()}, fmt.Errorf("collect cpu from %s: %w", path, err)
	}
	defer file.Close()

	return parseCPUStat(file)
}

type cpuSnapshot struct {
	modes map[string]float64
	total float64
	idle  float64
}

func parseCPUStat(r io.Reader) ([]Metric, error) {
	snapshot, err := parseCPUSnapshot(r)
	if err != nil {
		return nil, err
	}

	return appendLogicalCores(cpuMetrics(snapshot, nil)), nil
}

func readCPUSnapshot(path string) (cpuSnapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		return cpuSnapshot{}, err
	}
	defer file.Close()
	return parseCPUSnapshot(file)
}

func parseCPUSnapshot(r io.Reader) (cpuSnapshot, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return cpuSnapshot{}, err
		}
		return cpuSnapshot{}, fmt.Errorf("missing aggregate cpu line")
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuSnapshot{}, fmt.Errorf("unexpected aggregate cpu line")
	}

	modes := []string{"user", "nice", "system", "idle", "iowait", "irq", "softirq", "steal", "guest", "guest_nice"}
	snapshot := cpuSnapshot{modes: map[string]float64{}}

	for i := 1; i < len(fields) && i <= len(modes); i++ {
		ticks, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return cpuSnapshot{}, fmt.Errorf("parse cpu %s ticks: %w", modes[i-1], err)
		}
		snapshot.modes[modes[i-1]] = ticks / 100
		snapshot.total += ticks
		if modes[i-1] == "idle" || modes[i-1] == "iowait" {
			snapshot.idle += ticks
		}
	}

	return snapshot, scanner.Err()
}

func cpuMetrics(current cpuSnapshot, previous *cpuSnapshot) []Metric {
	metrics := make([]Metric, 0, len(current.modes)+1)
	modes := []string{"user", "nice", "system", "idle", "iowait", "irq", "softirq", "steal", "guest", "guest_nice"}
	for _, mode := range modes {
		seconds, ok := current.modes[mode]
		if !ok {
			continue
		}
		metrics = append(metrics, Metric{
			Name:   "gosysmon_cpu_seconds_total",
			Help:   "Seconds the CPUs spent in each mode since boot.",
			Type:   Counter,
			Labels: map[string]string{"mode": mode},
			Value:  seconds,
		})
	}

	if previous != nil {
		totalDelta := current.total - previous.total
		idleDelta := current.idle - previous.idle
		if totalDelta > 0 {
			metrics = append(metrics, Metric{
				Name:  "gosysmon_cpu_used_ratio",
				Help:  "Current non-idle CPU time ratio sampled over a short interval.",
				Type:  Gauge,
				Value: (totalDelta - idleDelta) / totalDelta,
			})
		}
	} else if current.total > 0 {
		metrics = append(metrics, Metric{
			Name:  "gosysmon_cpu_used_ratio",
			Help:  "Approximate non-idle CPU time ratio since boot.",
			Type:  Gauge,
			Value: (current.total - current.idle) / current.total,
		})
	}

	return metrics
}

func logicalCoresMetric() Metric {
	return Metric{
		Name:  "gosysmon_cpu_logical_cores",
		Help:  "Number of logical CPU cores visible to the process.",
		Type:  Gauge,
		Value: float64(runtime.NumCPU()),
	}
}

func appendLogicalCores(metrics []Metric) []Metric {
	return append(metrics, logicalCoresMetric())
}
