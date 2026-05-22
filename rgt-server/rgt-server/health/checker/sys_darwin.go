//go:build darwin

package checker

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

// Darwin syscall numbers — stable across amd64 and arm64.
// Using raw numbers avoids depending on darwin-specific exported symbols
// in the syscall package that are unavailable when cross-compiling from Linux.
const (
	sysSysctl       = 202 // SYS___SYSCTL
	sysSysctlByName = 274 // SYS_SYSCTLBYNAME
)

// Darwin MIB constants from <sys/sysctl.h>
const (
	ctlKern    = 1
	ctlHW      = 6
	ctlVM      = 2
	kernCPTime = 52 // KERN_CP_TIME  → kern.cp_time ([]uint32 or []uint64, 5 elements)
	hwMemSize  = 24 // HW_MEMSIZE    → hw.memsize (uint64)
	hwPageSize = 7  // HW_PAGESIZE   → hw.pagesize (int32)
	vmLoadAvg  = 2  // VM_LOADAVG    → vm.loadavg (struct loadavg)
)

// sysctl reads a kernel value via integer MIB array.
func sysctl(mib []int32, buf unsafe.Pointer, size *uintptr) error {
	_, _, errno := syscall.Syscall6(
		sysSysctl,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		uintptr(buf),
		uintptr(unsafe.Pointer(size)),
		0, // newp  (NULL = read-only)
		0, // newlen
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// sysctlByName reads a kernel value via its string name (like sysctlbyname(3)).
func sysctlByName(name string, buf unsafe.Pointer, size *uintptr) error {
	// The kernel expects a null-terminated C string.
	nameb := append([]byte(name), 0)
	_, _, errno := syscall.Syscall6(
		sysSysctlByName,
		uintptr(unsafe.Pointer(&nameb[0])),
		uintptr(len(nameb)),
		uintptr(buf),
		uintptr(unsafe.Pointer(size)),
		0, // newp
		0, // newlen
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// readUint64 reads a sysctl whose value is 4 or 8 bytes wide.
func readUint64ByMIB(mib []int32) (uint64, error) {
	var buf [8]byte
	size := uintptr(len(buf))
	if err := sysctl(mib, unsafe.Pointer(&buf[0]), &size); err != nil {
		return 0, err
	}
	switch size {
	case 4:
		return uint64(binary.LittleEndian.Uint32(buf[:4])), nil
	case 8:
		return binary.LittleEndian.Uint64(buf[:8]), nil
	default:
		return 0, fmt.Errorf("unexpected sysctl response size %d", size)
	}
}

func readUint64ByName(name string) (uint64, error) {
	var buf [8]byte
	size := uintptr(len(buf))
	if err := sysctlByName(name, unsafe.Pointer(&buf[0]), &size); err != nil {
		return 0, fmt.Errorf("sysctlbyname(%q): %w", name, err)
	}
	switch size {
	case 4:
		return uint64(binary.LittleEndian.Uint32(buf[:4])), nil
	case 8:
		return binary.LittleEndian.Uint64(buf[:8]), nil
	default:
		return 0, fmt.Errorf("sysctlbyname(%q): unexpected size %d", name, size)
	}
}

// ─── CPU ─────────────────────────────────────────────────────────────────────

// readCPUSample reads kern.cp_time (MIB [1, 52]).
//
// kern.cp_time returns 5 CPU-state tick counters:
// [CP_USER, CP_NICE, CP_SYS, CP_INTR, CP_IDLE]
// Width is uint32 on x86_64 and uint64 on arm64.
func readCPUSample() (rawCPUSample, error) {
	const states = 5
	// Allocate space for 5 × uint64 (worst case).
	var buf [states * 8]byte
	size := uintptr(len(buf))

	mib := []int32{ctlKern, kernCPTime}
	if err := sysctl(mib, unsafe.Pointer(&buf[0]), &size); err != nil {
		return rawCPUSample{}, fmt.Errorf("sysctl kern.cp_time: %w", err)
	}

	var ticks [5]uint64
	switch size {
	case states * 4:
		for i := range ticks {
			ticks[i] = uint64(binary.LittleEndian.Uint32(buf[i*4 : i*4+4]))
		}
	case states * 8:
		for i := range ticks {
			ticks[i] = binary.LittleEndian.Uint64(buf[i*8 : i*8+8])
		}
	default:
		return rawCPUSample{}, fmt.Errorf("kern.cp_time: unexpected byte count %d", size)
	}

	var total uint64
	for _, t := range ticks {
		total += t
	}
	return rawCPUSample{total: total, idle: ticks[4]}, nil // ticks[4] = CP_IDLE
}

// ─── Memory ───────────────────────────────────────────────────────────────────

// readMemStats reads hw.memsize for total RAM and computes available memory
// from vm.page_free_count + vm.page_inactive_count (both reclaimable without
// swapping) multiplied by hw.pagesize.
func readMemStats() (rawMemStats, error) {
	totalBytes, err := readUint64ByMIB([]int32{ctlHW, hwMemSize})
	if err != nil {
		return rawMemStats{}, fmt.Errorf("sysctl hw.memsize: %w", err)
	}

	pageSize, err := readUint64ByMIB([]int32{ctlHW, hwPageSize})
	if err != nil {
		return rawMemStats{}, fmt.Errorf("sysctl hw.pagesize: %w", err)
	}

	freePages, err := readUint64ByName("vm.page_free_count")
	if err != nil {
		return rawMemStats{}, err
	}

	// vm.page_inactive_count: pages that can be reclaimed; not all kernels
	// expose this sysctl, so we treat absence as zero.
	inactivePages, _ := readUint64ByName("vm.page_inactive_count")

	availBytes := (freePages + inactivePages) * pageSize
	return rawMemStats{
		totalKB:     totalBytes / 1024,
		availableKB: availBytes / 1024,
	}, nil
}

// ─── Load average ─────────────────────────────────────────────────────────────

// readLoadAvg reads vm.loadavg (MIB [2, 2]).
//
// The kernel returns a struct loadavg:
//
//	{ ldavg[3] int32 ; fscale int32 }  (16 bytes)
//
// floating-point load = ldavg[i] / fscale.
func readLoadAvg() (load1, load5, load15 float64, err error) {
	var buf [16]byte
	size := uintptr(len(buf))

	mib := []int32{ctlVM, vmLoadAvg}
	if e := sysctl(mib, unsafe.Pointer(&buf[0]), &size); e != nil {
		return 0, 0, 0, fmt.Errorf("sysctl vm.loadavg: %w", e)
	}
	if size < 16 {
		return 0, 0, 0, fmt.Errorf("vm.loadavg: short read (%d bytes)", size)
	}

	fscale := float64(int32(binary.LittleEndian.Uint32(buf[12:])))
	if fscale == 0 {
		return 0, 0, 0, fmt.Errorf("vm.loadavg: fscale is zero")
	}
	load1 = float64(int32(binary.LittleEndian.Uint32(buf[0:]))) / fscale
	load5 = float64(int32(binary.LittleEndian.Uint32(buf[4:]))) / fscale
	load15 = float64(int32(binary.LittleEndian.Uint32(buf[8:]))) / fscale
	return load1, load5, load15, nil
}
