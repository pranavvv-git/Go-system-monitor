//go:build windows

package collector

import (
	"fmt"
	"time"
	"unsafe"
)

var (
	procGetSystemTimes = kernel32.NewProc("GetSystemTimes")
)

type fileTime struct {
	lowDateTime  uint32
	highDateTime uint32
}

func platformCPUMetrics() ([]Metric, error) {
	first, err := windowsCPUSnapshot()
	if err != nil {
		return nil, err
	}
	time.Sleep(500 * time.Millisecond)
	second, err := windowsCPUSnapshot()
	if err != nil {
		return nil, err
	}

	idleDelta := second.idle - first.idle
	kernelDelta := second.kernel - first.kernel
	userDelta := second.user - first.user

	metrics := windowsCPUMetrics(second)
	if ratio, ok := windowsCPUUsedRatio(idleDelta, kernelDelta, userDelta); ok {
		metrics = append(metrics, Metric{
			Name:  "gosysmon_cpu_used_ratio",
			Help:  "Current non-idle CPU time ratio sampled over a short interval.",
			Type:  Gauge,
			Value: ratio,
		})
	}

	return metrics, nil
}

type windowsCPUSample struct {
	idle   float64
	kernel float64
	user   float64
}

func windowsCPUSnapshot() (windowsCPUSample, error) {
	var idle, kernel, user fileTime
	ret, _, err := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if ret == 0 {
		return windowsCPUSample{}, fmt.Errorf("GetSystemTimes: %w", err)
	}

	return windowsCPUSample{
		idle:   fileTimeSeconds(idle),
		kernel: fileTimeSeconds(kernel),
		user:   fileTimeSeconds(user),
	}, nil
}

func windowsCPUMetrics(sample windowsCPUSample) []Metric {
	system := sample.kernel - sample.idle
	if system < 0 {
		system = 0
	}
	return []Metric{
		{
			Name:   "gosysmon_cpu_seconds_total",
			Help:   "Seconds the CPUs spent in each mode since boot.",
			Type:   Counter,
			Labels: map[string]string{"mode": "idle"},
			Value:  sample.idle,
		},
		{
			Name:   "gosysmon_cpu_seconds_total",
			Help:   "Seconds the CPUs spent in each mode since boot.",
			Type:   Counter,
			Labels: map[string]string{"mode": "system"},
			Value:  system,
		},
		{
			Name:   "gosysmon_cpu_seconds_total",
			Help:   "Seconds the CPUs spent in each mode since boot.",
			Type:   Counter,
			Labels: map[string]string{"mode": "user"},
			Value:  sample.user,
		},
	}
}

func windowsCPUUsedRatio(idleDelta, kernelDelta, userDelta float64) (float64, bool) {
	// On Windows, kernel time includes idle time.
	totalDelta := kernelDelta + userDelta
	if totalDelta <= 0 {
		return 0, false
	}

	busyDelta := (kernelDelta - idleDelta) + userDelta
	if busyDelta < 0 {
		busyDelta = 0
	}
	ratio := busyDelta / totalDelta
	if ratio > 1 {
		ratio = 1
	}
	return ratio, true
}

func fileTimeSeconds(value fileTime) float64 {
	ticks := uint64(value.highDateTime)<<32 | uint64(value.lowDateTime)
	return float64(ticks) / 10000000
}
