//go:build !windows

package collector

func platformMemoryMetrics() ([]Metric, error) {
	return nil, nil
}
