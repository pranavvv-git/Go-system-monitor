//go:build !linux && !windows

package collector

import (
	"context"
	"fmt"
	"runtime"
)

type DiskCollector struct {
	Paths []string
}

func (c DiskCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return nil, fmt.Errorf("disk collection is implemented for linux; current platform is %s", runtime.GOOS)
}
