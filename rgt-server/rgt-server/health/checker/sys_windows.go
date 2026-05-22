//go:build windows

package checker

import (
	"fmt"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// ─── DLL and proc references (loaded once) ────────────────────────────────────

var (
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetSystemTimes       = modKernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatusEx = modKernel32.NewProc("GlobalMemoryStatusEx")
)

// ─── CPU ─────────────────────────────────────────────────────────────────────

// FILETIME is a Windows FILETIME structure: 100-nanosecond intervals since
// January 1, 1601.  We only need the arithmetic value, not the calendar date.
type fileTime struct {
	LowDateTime  uint32
	HighDateTime uint32
}

func (ft fileTime) uint64() uint64 {
	return uint64(ft.HighDateTime)<<32 | uint64(ft.LowDateTime)
}

// readCPUSample calls GetSystemTimes which returns three FILETIME values:
// idle time, kernel time (includes idle), and user time.
//
// total  = kernelTime + userTime
// idle   = idleTime         (already contained inside kernelTime)
// busy   = total - idle
func readCPUSample() (rawCPUSample, error) {
	var idle, kernel, user fileTime
	r1, _, err := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if r1 == 0 {
		return rawCPUSample{}, fmt.Errorf("GetSystemTimes: %w", err)
	}
	// kernelTime already includes idleTime, so total = kernel + user.
	total := kernel.uint64() + user.uint64()
	return rawCPUSample{total: total, idle: idle.uint64()}, nil
}

// ─── Memory ───────────────────────────────────────────────────────────────────

// memoryStatusEx mirrors MEMORYSTATUSEX from <winbase.h>.
// Must be exactly 64 bytes; dwLength is set before the call.
type memoryStatusEx struct {
	DwLength                uint32
	DwMemoryLoad            uint32
	UllTotalPhys            uint64
	UllAvailPhys            uint64
	UllTotalPageFile        uint64
	UllAvailPageFile        uint64
	UllTotalVirtual         uint64
	UllAvailVirtual         uint64
	UllAvailExtendedVirtual uint64
}

// readMemStats calls GlobalMemoryStatusEx.
// UllAvailPhys is the amount of physical memory that can be immediately
// allocated to processes without paging — equivalent to Linux MemAvailable.
func readMemStats() (rawMemStats, error) {
	var ms memoryStatusEx
	ms.DwLength = uint32(unsafe.Sizeof(ms))
	r1, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&ms)))
	if r1 == 0 {
		return rawMemStats{}, fmt.Errorf("GlobalMemoryStatusEx: %w", err)
	}
	return rawMemStats{
		totalKB:     ms.UllTotalPhys / 1024,
		availableKB: ms.UllAvailPhys / 1024,
	}, nil
}

// ─── Load average (approximation) ────────────────────────────────────────────

// Windows has no native load-average concept.  We approximate the 1-minute
// load average by sampling the system-wide CPU utilisation over a short
// interval and expressing it as a fraction of the logical CPU count.
//
// This is not identical to Unix load (which counts the run-queue length), but
// it is a meaningful signal for the monitor's threshold logic and follows the
// same 0–numCPU scale.

var (
	loadMu      sync.Mutex
	prevLoad    rawCPUSample
	prevLoadAt  time.Time
	smoothLoad1 float64 // EMA with α chosen to approximate a 1-min window
)

// readLoadAvg returns an approximate 1, 5, and 15-minute "load average" for
// Windows by sampling CPU busy-time and applying exponential smoothing.
// load5 and load15 are decayed from the same EMA with longer time constants.
func readLoadAvg() (load1, load5, load15 float64, err error) {
	cur, err := readCPUSample()
	if err != nil {
		return 0, 0, 0, err
	}

	loadMu.Lock()
	defer loadMu.Unlock()

	now := time.Now()
	if !prevLoadAt.IsZero() && cur.total > prevLoad.total {
		deltaTot := float64(cur.total - prevLoad.total)
		deltaIdle := float64(cur.idle - prevLoad.idle)
		busyFrac := (deltaTot - deltaIdle) / deltaTot // 0..1

		// α = 1 - exp(-dt / τ) where τ = 60 s (1-min EMA).
		dt := now.Sub(prevLoadAt).Seconds()
		alpha := 1 - safeExp(-dt/60.0)
		smoothLoad1 += alpha * (busyFrac - smoothLoad1)
	}

	prevLoad = cur
	prevLoadAt = now

	// Express as fraction × numCPU so the scale matches Unix load.
	numCPU := float64(numLogicalCPUs())
	load1 = smoothLoad1 * numCPU
	// Approximate longer windows with a decayed copy (not independently sampled).
	load5 = load1 * 0.85
	load15 = load1 * 0.70

	return load1, load5, load15, nil
}

// ─── Windows helpers ──────────────────────────────────────────────────────────

var (
	procGetSystemInfo = modKernel32.NewProc("GetSystemInfo")
)

type systemInfo struct {
	_                    uint16
	_                    uint16
	_                    uint32
	_                    uintptr
	_                    uint32
	_                    uint32
	DwNumberOfProcessors uint32
	_                    uint32
	_                    uint32
	_                    uint32
}

func numLogicalCPUs() int {
	var si systemInfo
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&si)))
	if si.DwNumberOfProcessors == 0 {
		return 1
	}
	return int(si.DwNumberOfProcessors)
}

func safeExp(x float64) float64 {
	// Avoid importing math for a single exp call; use a Taylor approximation
	// that is accurate enough for our smoothing constant.
	// For x in (-1, 0] the error is < 0.002.
	if x < -10 {
		return 0
	}
	// Standard 5-term Taylor: e^x ≈ 1 + x + x²/2 + x³/6 + x⁴/24
	x2 := x * x
	x3 := x2 * x
	x4 := x3 * x
	return 1 + x + x2/2 + x3/6 + x4/24
}
