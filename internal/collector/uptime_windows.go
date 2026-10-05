//go:build windows

package collector

import "context"

var procGetTickCount64 = kernel32.NewProc("GetTickCount64")

type UptimeCollector struct {
	UptimePath string
}

func (c UptimeCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	ret, _, _ := procGetTickCount64.Call()
	seconds := float64(ret) / 1000
	return uptimeMetric(seconds), nil
}
