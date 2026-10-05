package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aadityya4real/Go-system-monitor/internal/collector"
)

func TestStatsEndpointReturnsJSON(t *testing.T) {
	app := New([]collector.Collector{
		collector.CollectorFunc(func(context.Context) ([]collector.Metric, error) {
			return []collector.Metric{{
				Name:  "gosysmon_cpu_logical_cores",
				Help:  "Number of logical CPU cores visible to the process.",
				Type:  collector.Gauge,
				Value: 8,
			}}, nil
		}),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response StatsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if len(response.Metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(response.Metrics))
	}
}

func TestMetricsEndpointReturnsPrometheusText(t *testing.T) {
	app := New([]collector.Collector{
		collector.CollectorFunc(func(context.Context) ([]collector.Metric, error) {
			return []collector.Metric{{
				Name:  "gosysmon_cpu_logical_cores",
				Help:  "Number of logical CPU cores visible to the process.",
				Type:  collector.Gauge,
				Value: 8,
			}}, nil
		}),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "gosysmon_cpu_logical_cores 8") {
		t.Fatalf("missing metric in response:\n%s", rec.Body.String())
	}
}

func TestDashboardIsServed(t *testing.T) {
	app := New(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Go System Monitor") {
		t.Fatalf("missing dashboard markup")
	}
}

func TestHealthEndpoint(t *testing.T) {
	app := New(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("missing ok health response")
	}
}

func TestStatsEndpointPartialFailure(t *testing.T) {
	app := New([]collector.Collector{
		collector.CollectorFunc(func(context.Context) ([]collector.Metric, error) {
			return []collector.Metric{{Name: "sample_metric", Type: collector.Gauge, Value: 1}}, errors.New("partial")
		}),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for partial data, got %d", rec.Code)
	}
	var response StatsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Warning == "" {
		t.Fatalf("expected warning in response")
	}
}

func TestMetricsEndpointAllCollectorsFail(t *testing.T) {
	app := New([]collector.Collector{
		collector.CollectorFunc(func(context.Context) ([]collector.Metric, error) {
			return nil, errors.New("failed")
		}),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestGetOnlyRoutesRejectPost(t *testing.T) {
	app := New(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
