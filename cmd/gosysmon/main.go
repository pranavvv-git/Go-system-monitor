package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aadityya4real/Go-system-monitor/internal/collector"
	"github.com/aadityya4real/Go-system-monitor/internal/prometheus"
	"github.com/aadityya4real/Go-system-monitor/internal/server"
)

type pathList []string

func (p *pathList) String() string {
	return strings.Join(*p, ",")
}

func (p *pathList) Set(value string) error {
	if value == "" {
		return fmt.Errorf("path cannot be empty")
	}
	*p = append(*p, value)
	return nil
}

func main() {
	var paths pathList
	watch := flag.Bool("watch", false, "keep collecting metrics until interrupted")
	serve := flag.Bool("serve", false, "start the web dashboard and HTTP metrics server")
	addr := flag.String("addr", ":9090", "HTTP address for --serve")
	interval := flag.Duration("interval", 5*time.Second, "watch collection interval")
	quiet := flag.Bool("quiet", false, "suppress collector warnings on stderr")
	flag.Var(&paths, "path", "filesystem path to collect disk metrics for; repeatable")
	flag.Parse()

	if *interval <= 0 {
		fmt.Fprintln(os.Stderr, "interval must be positive")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	collectors := []collector.Collector{
		collector.CPUCollector{},
		collector.MemoryCollector{},
		collector.DiskCollector{Paths: paths},
		collector.NetworkCollector{},
		collector.UptimeCollector{},
	}

	if *serve {
		fmt.Fprintf(os.Stdout, "Go System Monitor dashboard: %s\n", server.URL(*addr))
		err := server.New(collectors, slog.Default()).ListenAndServe(ctx, *addr)
		if err != nil && err != context.Canceled {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if !*watch {
		if err := collectOnce(ctx, collectors, *quiet); err != nil {
			os.Exit(1)
		}
		return
	}

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		if err := collectOnce(ctx, collectors, *quiet); err != nil {
			os.Exit(1)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintln(os.Stdout)
		}
	}
}

func collectOnce(ctx context.Context, collectors []collector.Collector, quiet bool) error {
	metrics, err := collector.CollectAll(ctx, collectors...)
	if err != nil && !quiet {
		fmt.Fprintln(os.Stderr, "warning:", err)
	}

	if len(metrics) == 0 {
		if err == nil {
			err = fmt.Errorf("no metrics collected")
		}
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return prometheus.Render(os.Stdout, metrics)
}
