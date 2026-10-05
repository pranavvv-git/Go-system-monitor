//go:build windows

package collector

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
)

type memoryStatusEx struct {
	length               uint32
	memoryLoad           uint32
	totalPhys            uint64
	availPhys            uint64
	totalPageFile        uint64
	availPageFile        uint64
	totalVirtual         uint64
	availVirtual         uint64
	availExtendedVirtual uint64
}

func platformMemoryMetrics() ([]Metric, error) {
	status := memoryStatusEx{length: uint32(unsafe.Sizeof(memoryStatusEx{}))}
	ret, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if ret == 0 {
		return nil, fmt.Errorf("GlobalMemoryStatusEx: %w", err)
	}

	return windowsMemoryMetrics(status), nil
}

func windowsMemoryMetrics(status memoryStatusEx) []Metric {
	metrics := []Metric{
		{
			Name:  "gosysmon_memory_total_bytes",
			Help:  memoryHelp("gosysmon_memory_total_bytes"),
			Type:  Gauge,
			Value: float64(status.totalPhys),
		},
		{
			Name:  "gosysmon_memory_available_bytes",
			Help:  memoryHelp("gosysmon_memory_available_bytes"),
			Type:  Gauge,
			Value: float64(status.availPhys),
		},
		{
			Name:  "gosysmon_memory_free_bytes",
			Help:  memoryHelp("gosysmon_memory_free_bytes"),
			Type:  Gauge,
			Value: float64(status.availPhys),
		},
		{
			Name:  "gosysmon_swap_total_bytes",
			Help:  memoryHelp("gosysmon_swap_total_bytes"),
			Type:  Gauge,
			Value: float64(status.totalPageFile),
		},
		{
			Name:  "gosysmon_swap_free_bytes",
			Help:  memoryHelp("gosysmon_swap_free_bytes"),
			Type:  Gauge,
			Value: float64(status.availPageFile),
		},
	}

	if status.totalPhys > 0 {
		metrics = append(metrics, Metric{
			Name:  "gosysmon_memory_used_ratio",
			Help:  "Memory usage ratio based on available physical memory.",
			Type:  Gauge,
			Value: float64(status.totalPhys-status.availPhys) / float64(status.totalPhys),
		})
	}

	return metrics
}
