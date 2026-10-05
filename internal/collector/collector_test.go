package collector

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCollectAllRunsCollectorsConcurrently(t *testing.T) {
	first := CollectorFunc(func(context.Context) ([]Metric, error) {
		return []Metric{{Name: "z_metric", Type: Gauge, Value: 1}}, nil
	})
	second := CollectorFunc(func(context.Context) ([]Metric, error) {
		return []Metric{{Name: "a_metric", Type: Gauge, Value: 2}}, errors.New("partial failure")
	})

	metrics, err := CollectAll(context.Background(), first, second)
	if err == nil {
		t.Fatal("expected joined error")
	}
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	if metrics[0].Name != "a_metric" || metrics[1].Name != "z_metric" {
		t.Fatalf("metrics were not sorted by name: %#v", metrics)
	}
}

func TestParseCPUStat(t *testing.T) {
	metrics, err := parseCPUStat(stringsReader("cpu  100 20 30 850 10 0 0 0 0 0\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) < 3 {
		t.Fatalf("expected cpu metrics, got %d", len(metrics))
	}
}

func TestParseMeminfo(t *testing.T) {
	input := "MemTotal:       1000 kB\nMemAvailable:    250 kB\nMemFree:         100 kB\nSwapTotal:         0 kB\nSwapFree:          0 kB\n"
	metrics, err := parseMeminfo(stringsReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 6 {
		t.Fatalf("expected 6 metrics, got %d", len(metrics))
	}
}

func stringsReader(s string) *strings.Reader {
	return strings.NewReader(s)
}
