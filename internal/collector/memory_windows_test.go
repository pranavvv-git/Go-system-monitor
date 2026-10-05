//go:build windows

package collector

import "testing"

func TestWindowsMemoryMetricsIncludesFreeBytes(t *testing.T) {
	metrics := windowsMemoryMetrics(memoryStatusEx{
		totalPhys:     100,
		availPhys:     25,
		totalPageFile: 200,
		availPageFile: 75,
	})

	values := map[string]float64{}
	for _, metric := range metrics {
		values[metric.Name] = metric.Value
	}

	if values["gosysmon_memory_free_bytes"] != 25 {
		t.Fatalf("expected free bytes to match available physical memory, got %v", values["gosysmon_memory_free_bytes"])
	}
	if values["gosysmon_memory_used_ratio"] != 0.75 {
		t.Fatalf("expected used ratio 0.75, got %v", values["gosysmon_memory_used_ratio"])
	}
}
