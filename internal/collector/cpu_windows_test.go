//go:build windows

package collector

import "testing"

func TestFileTimeSeconds(t *testing.T) {
	got := fileTimeSeconds(fileTime{lowDateTime: 10000000})
	if got != 1 {
		t.Fatalf("expected 1 second, got %v", got)
	}
}

func TestWindowsCPUMetrics(t *testing.T) {
	metrics := windowsCPUMetrics(windowsCPUSample{idle: 1, kernel: 3, user: 3})
	if len(metrics) != 3 {
		t.Fatalf("expected 3 metrics, got %d", len(metrics))
	}
}

func TestWindowsCPUUsedRatio(t *testing.T) {
	got, ok := windowsCPUUsedRatio(2, 5, 3)
	if !ok {
		t.Fatal("expected ratio")
	}
	if got != 0.75 {
		t.Fatalf("expected 0.75, got %v", got)
	}
}

func TestWindowsCPUUsedRatioClampsInvalidDeltas(t *testing.T) {
	got, ok := windowsCPUUsedRatio(10, 2, 1)
	if !ok {
		t.Fatal("expected ratio")
	}
	if got != 0 {
		t.Fatalf("expected negative busy time to clamp to 0, got %v", got)
	}

	got, ok = windowsCPUUsedRatio(0, 1, 0)
	if !ok {
		t.Fatal("expected ratio")
	}
	if got != 1 {
		t.Fatalf("expected ratio to clamp to 1, got %v", got)
	}

	if _, ok := windowsCPUUsedRatio(0, 0, 0); ok {
		t.Fatal("expected zero elapsed time to be ignored")
	}
}
