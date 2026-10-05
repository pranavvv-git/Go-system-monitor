//go:build linux

package collector

import (
	"context"
	"testing"
)

func TestDiskCollectorLinuxRoot(t *testing.T) {
	metrics, err := (DiskCollector{Paths: []string{"/"}}).Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) < 4 {
		t.Fatalf("expected disk metrics, got %d", len(metrics))
	}
}
