//go:build windows

package collector

import "testing"

func TestWindowsVolumeRoot(t *testing.T) {
	got, err := windowsVolumeRoot(`C:\Users`)
	if err != nil {
		t.Fatal(err)
	}
	if got != `C:\` {
		t.Fatalf("expected C:\\, got %s", got)
	}
}
