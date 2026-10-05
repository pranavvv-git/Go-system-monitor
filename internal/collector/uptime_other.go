//go:build !linux && !windows

package collector

import (
	"context"
	"fmt"
	"runtime"
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
	return nil, fmt.Errorf("uptime collection is not implemented for %s", runtime.GOOS)
}
