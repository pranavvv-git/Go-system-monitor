package server

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/aadityya4real/Go-system-monitor/internal/collector"
	"github.com/aadityya4real/Go-system-monitor/internal/prometheus"
)

//go:embed static
var staticFiles embed.FS

type Server struct {
	collectors []collector.Collector
	logger     *slog.Logger
}

type StatsResponse struct {
	CollectedAt time.Time          `json:"collected_at"`
	DurationMS  int64              `json:"duration_ms"`
	Metrics     []collector.Metric `json:"metrics"`
	Warning     string             `json:"warning,omitempty"`
}

func New(collectors []collector.Collector, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{collectors: collectors, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/stats", s.handleStats)
	mux.HandleFunc("GET /metrics", s.handleMetrics)

	assets, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /", http.FileServer(http.FS(assets)))
	return withHeaders(mux)
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("dashboard listening", "addr", addr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return ctx.Err()
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	metrics, err := collector.CollectAll(r.Context(), s.collectors...)
	response := StatsResponse{
		CollectedAt: start.UTC(),
		DurationMS:  time.Since(start).Milliseconds(),
		Metrics:     metrics,
	}
	if err != nil {
		response.Warning = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	if len(metrics) == 0 && err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		s.logger.Error("encode stats response", "error", encodeErr)
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := collector.CollectAll(r.Context(), s.collectors...)
	if len(metrics) == 0 && err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	var buf bytes.Buffer
	if renderErr := prometheus.Render(&buf, metrics); renderErr != nil {
		http.Error(w, renderErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	if err != nil {
		w.Header().Set("X-GoSysMon-Warning", err.Error())
	}
	_, _ = w.Write(buf.Bytes())
}

func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func URL(addr string) string {
	host := addr
	if host == "" {
		host = ":9090"
	}
	if host[0] == ':' {
		host = "localhost" + host
	}
	return fmt.Sprintf("http://%s", host)
}
