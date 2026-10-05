package collector

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Metric struct {
	Name   string            `json:"name"`
	Help   string            `json:"help"`
	Type   MetricType        `json:"type"`
	Labels map[string]string `json:"labels,omitempty"`
	Value  float64           `json:"value"`
}

type Collector interface {
	Collect(context.Context) ([]Metric, error)
}

type CollectorFunc func(context.Context) ([]Metric, error)

func (f CollectorFunc) Collect(ctx context.Context) ([]Metric, error) {
	return f(ctx)
}

func CollectAll(ctx context.Context, collectors ...Collector) ([]Metric, error) {
	type result struct {
		metrics []Metric
		err     error
	}

	results := make(chan result, len(collectors))
	var wg sync.WaitGroup

	for _, c := range collectors {
		wg.Add(1)
		go func(c Collector) {
			defer wg.Done()
			metrics, err := c.Collect(ctx)
			results <- result{metrics: metrics, err: err}
		}(c)
	}

	wg.Wait()
	close(results)

	var metrics []Metric
	var errs []error
	for r := range results {
		metrics = append(metrics, r.metrics...)
		if r.err != nil {
			errs = append(errs, r.err)
		}
	}

	sort.SliceStable(metrics, func(i, j int) bool {
		left := metrics[i].Name + labelsKey(metrics[i].Labels)
		right := metrics[j].Name + labelsKey(metrics[j].Labels)
		return left < right
	})

	return metrics, errors.Join(errs...)
}

func labelsKey(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, key := range keys {
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(labels[key])
		b.WriteString(";")
	}
	return b.String()
}
