//go:build linux

package collector

import (
	"strings"
	"testing"
)

func TestParseNetDev(t *testing.T) {
	input := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 100 1 0 0 0 0 0 0 200 2 0 0 0 0 0 0
  eth0: 300 3 0 0 0 0 0 0 400 4 0 0 0 0 0 0
`
	metrics, err := parseNetDev(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 4 {
		t.Fatalf("expected 4 metrics, got %d", len(metrics))
	}
	if metrics[0].Name != "gosysmon_network_rx_bytes_total" || metrics[0].Value != 100 {
		t.Fatalf("unexpected first metric: %#v", metrics[0])
	}
}
