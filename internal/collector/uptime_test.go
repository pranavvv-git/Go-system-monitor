//go:build linux

package collector

import "testing"

func TestParseUptime(t *testing.T) {
	got, err := parseUptime([]byte("123.45 678.90\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got != 123.45 {
		t.Fatalf("expected 123.45, got %v", got)
	}
}
