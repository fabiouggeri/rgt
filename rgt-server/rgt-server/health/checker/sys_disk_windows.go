//go:build windows

package checker

import (
	"fmt"
	"syscall"
	"unsafe"
)

var procGetDiskFreeSpaceExW = modKernel32.NewProc("GetDiskFreeSpaceExW")

// readDiskStats calls GetDiskFreeSpaceExW for the given path (drive root
// or directory, e.g. "C:\\", "D:\\data").
//
// The three OUT parameters are:
//   - lpFreeBytesAvailableToCaller  (for the current user, respecting quotas)
//   - lpTotalNumberOfBytes          (total disk capacity)
//   - lpTotalNumberOfFreeBytes      (total free, ignoring quotas)
//
// We use lpTotalNumberOfFreeBytes (not the caller-quota-limited one) to be
// consistent with the Unix implementation which also uses Bfree (root free).
func readDiskStats(mountPoint string) (rawDiskStats, error) {
	pathPtr, err := syscall.UTF16PtrFromString(mountPoint)
	if err != nil {
		return rawDiskStats{}, fmt.Errorf("UTF16PtrFromString(%q): %w", mountPoint, err)
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	r1, _, callErr := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if r1 == 0 {
		return rawDiskStats{}, fmt.Errorf("GetDiskFreeSpaceExW(%q): %w", mountPoint, callErr)
	}

	return rawDiskStats{
		totalBytes: totalBytes,
		freeBytes:  totalFreeBytes,
	}, nil
}
