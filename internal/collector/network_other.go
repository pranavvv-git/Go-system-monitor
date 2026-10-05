//go:build !linux && !windows

package collector

import (
	"context"
	"fmt"
	"runtime"
)

type NetworkCollector struct {
	DevPath string
}

func (c NetworkCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return nil, fmt.Errorf("network collection is not implemented for %s", runtime.GOOS)
}
