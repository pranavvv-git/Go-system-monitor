//go:build windows

package collector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	procGetDiskFreeSpaceExW = kernel32.NewProc("GetDiskFreeSpaceExW")
)

type DiskCollector struct {
	Paths []string
}

func (c DiskCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	paths := c.Paths
	if len(paths) == 0 {
		drives, err := windowsDriveRoots()
		if err != nil {
			return nil, err
		}
		paths = drives
	}

	var metrics []Metric
	for _, path := range paths {
		root, err := windowsVolumeRoot(path)
		if err != nil {
			return metrics, err
		}

		var available, total, free uint64
		rootPtr, err := syscall.UTF16PtrFromString(root)
		if err != nil {
			return metrics, fmt.Errorf("encode path %s: %w", root, err)
		}

		ret, _, callErr := procGetDiskFreeSpaceExW.Call(
			uintptr(unsafe.Pointer(rootPtr)),
			uintptr(unsafe.Pointer(&available)),
			uintptr(unsafe.Pointer(&total)),
			uintptr(unsafe.Pointer(&free)),
		)
		if ret == 0 {
			return metrics, fmt.Errorf("GetDiskFreeSpaceExW %s: %w", root, callErr)
		}

		labels := map[string]string{"path": root}
		metrics = append(metrics,
			Metric{Name: "gosysmon_disk_size_bytes", Help: "Total filesystem size in bytes.", Type: Gauge, Labels: labels, Value: float64(total)},
			Metric{Name: "gosysmon_disk_free_bytes", Help: "Free filesystem bytes, including reserved blocks.", Type: Gauge, Labels: labels, Value: float64(free)},
			Metric{Name: "gosysmon_disk_available_bytes", Help: "Free filesystem bytes available to unprivileged users.", Type: Gauge, Labels: labels, Value: float64(available)},
		)

		if total > 0 {
			metrics = append(metrics, Metric{Name: "gosysmon_disk_used_ratio", Help: "Filesystem usage ratio.", Type: Gauge, Labels: labels, Value: float64(total-free) / float64(total)})
		}
	}

	return metrics, nil
}

func windowsDriveRoots() ([]string, error) {
	var roots []string
	for letter := 'A'; letter <= 'Z'; letter++ {
		root := fmt.Sprintf("%c:\\", letter)
		if windowsDiskAvailable(root) {
			roots = append(roots, root)
		}
	}
	if len(roots) > 0 {
		return roots, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	root, err := windowsVolumeRoot(wd)
	if err != nil {
		return nil, err
	}
	return []string{root}, nil
}

func currentDriveRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return windowsVolumeRoot(wd)
}

func windowsDiskAvailable(root string) bool {
	rootPtr, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return false
	}
	var available, total, free uint64
	ret, _, _ := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(rootPtr)),
		uintptr(unsafe.Pointer(&available)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&free)),
	)
	return ret != 0 && total > 0
}

func windowsVolumeRoot(path string) (string, error) {
	if path == "/" || path == `\` {
		return currentDriveRoot()
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", path, err)
	}

	volume := filepath.VolumeName(abs)
	if volume == "" {
		return "", fmt.Errorf("path %s does not include a Windows volume", path)
	}

	if strings.HasPrefix(volume, `\\`) {
		return volume + `\`, nil
	}
	return volume + `\`, nil
}
