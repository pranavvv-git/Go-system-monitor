package prometheus

import (
	"strings"
	"testing"

	"github.com/aadityya4real/Go-system-monitor/internal/collector"
)

func TestRenderPrometheusText(t *testing.T) {
	var out strings.Builder
	err := Render(&out, []collector.Metric{{
		Name:   "gosysmon_disk_size_bytes",
		Help:   "Total filesystem size in bytes.",
		Type:   collector.Gauge,
		Labels: map[string]string{"path": `/var/lib/"app"`},
		Value:  42,
	}})
	if err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, "# HELP gosysmon_disk_size_bytes Total filesystem size in bytes.") {
		t.Fatalf("missing HELP line:\n%s", got)
	}
	if !strings.Contains(got, `gosysmon_disk_size_bytes{path="/var/lib/\"app\""} 42`) {
		t.Fatalf("missing escaped metric line:\n%s", got)
	}
}

func TestRenderEscapesLabels(t *testing.T) {
	var out strings.Builder
	err := Render(&out, []collector.Metric{{
		Name:   "sample_metric",
		Help:   "Sample\nmetric.",
		Type:   collector.Gauge,
		Labels: map[string]string{"path": "W:\\\nquoted\""},
		Value:  1,
	}})
	if err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, "# HELP sample_metric Sample metric.") {
		t.Fatalf("help text was not sanitized:\n%s", got)
	}
	if !strings.Contains(got, `sample_metric{path="W:\\\nquoted\""} 1`) {
		t.Fatalf("label was not escaped:\n%s", got)
	}
}
