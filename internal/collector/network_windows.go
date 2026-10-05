//go:build windows

package collector

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	errorInsufficientBuffer = 122
)

var (
	iphlpapi       = syscall.NewLazyDLL("iphlpapi.dll")
	procGetIfTable = iphlpapi.NewProc("GetIfTable")
)

type NetworkCollector struct {
	DevPath string
}

type mibIfRow struct {
	Name            [256]uint16
	Index           uint32
	Type            uint32
	MTU             uint32
	Speed           uint32
	PhysAddrLen     uint32
	PhysAddr        [8]byte
	AdminStatus     uint32
	OperStatus      uint32
	LastChange      uint32
	InOctets        uint32
	InUcastPkts     uint32
	InNUcastPkts    uint32
	InDiscards      uint32
	InErrors        uint32
	InUnknownProtos uint32
	OutOctets       uint32
	OutUcastPkts    uint32
	OutNUcastPkts   uint32
	OutDiscards     uint32
	OutErrors       uint32
	OutQLen         uint32
	DescrLen        uint32
	Descr           [256]byte
}

func (c NetworkCollector) Collect(ctx context.Context) ([]Metric, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var size uint32
	ret, _, _ := procGetIfTable.Call(0, uintptr(unsafe.Pointer(&size)), 0)
	if ret != errorInsufficientBuffer && size == 0 {
		return nil, fmt.Errorf("GetIfTable size query failed: %d", ret)
	}

	buf := make([]byte, size)
	ret, _, err := procGetIfTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 1)
	if ret != 0 {
		return nil, fmt.Errorf("GetIfTable: %w", err)
	}

	count := *(*uint32)(unsafe.Pointer(&buf[0]))
	rowSize := unsafe.Sizeof(mibIfRow{})
	metrics := make([]Metric, 0, int(count)*2)
	for i := uint32(0); i < count; i++ {
		row := (*mibIfRow)(unsafe.Pointer(uintptr(unsafe.Pointer(&buf[4])) + uintptr(i)*rowSize))
		name := rowDescription(row)
		if name == "" {
			name = fmt.Sprintf("if%d", row.Index)
		}
		metrics = append(metrics, networkMetrics(name, float64(row.InOctets), float64(row.OutOctets))...)
	}
	return metrics, nil
}

func rowDescription(row *mibIfRow) string {
	n := int(row.DescrLen)
	if n > len(row.Descr) {
		n = len(row.Descr)
	}
	return strings.Trim(string(row.Descr[:n]), "\x00 \t\r\n")
}
