//go:build linux || darwin

package checker

import (
	"fmt"
	"syscall"
)

// readDiskStats calls statfs(2) — available on both Linux and macOS —
// and returns the raw capacity figures for the given mount point.
//
// We use Bfree (blocks free to root) rather than Bavail (blocks free
// to unprivileged processes) so that the usage percentage reflects the
// true on-disk occupancy, not the user-visible one.  Operators monitoring
// disk health care about the total fill level.
func readDiskStats(mountPoint string) (rawDiskStats, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(mountPoint, &stat); err != nil {
		return rawDiskStats{}, fmt.Errorf("statfs(%q): %w", mountPoint, err)
	}

	blockSize := uint64(stat.Bsize)
	total := stat.Blocks * blockSize
	free := stat.Bfree * blockSize

	return rawDiskStats{
		totalBytes: total,
		freeBytes:  free,
	}, nil
}
