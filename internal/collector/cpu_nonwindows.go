//go:build !windows

package collector

func platformCPUMetrics() ([]Metric, error) {
	return nil, nil
}
