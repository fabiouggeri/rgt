package checker

import (
	"fmt"
	"rgt-server/health/metric"
)

// DiskChecker measures used disk space as a percentage of total capacity for a
// given mount point or drive root.
//
// Typical values:
//   - Linux / macOS : "/"  "/data"  "/var"
//   - Windows       : `C:\`  `D:\`
//
// Platform dispatch at compile time:
//
//	sys_disk_unix.go    (linux || darwin) → statfs(2)
//	sys_disk_windows.go (windows)         → GetDiskFreeSpaceExW
type DiskChecker struct {
	mountPoint string
}

// NewDiskChecker creates a DiskChecker for the given mount point.
func NewDiskChecker(mountPoint string) *DiskChecker {
	return &DiskChecker{
		mountPoint: mountPoint,
	}
}

// Name returns a unique identifier that includes the mount point, so multiple
// DiskChecker instances can coexist without key collisions in the monitor.
func (d *DiskChecker) Name() string {
	return fmt.Sprintf("DISK(%s)", d.mountPoint)
}

func (d *DiskChecker) Unit() string {
	return "%"
}

// Check reads disk usage for the configured mount point and returns it as a
// percentage.
func (d *DiskChecker) Check() (metric.MetricValue, error) {
	stats, err := readDiskStats(d.mountPoint)
	if err != nil {
		return metric.MetricValue{}, fmt.Errorf("disk checker (%s): %w", d.mountPoint, err)
	}

	if stats.totalBytes == 0 {
		return metric.MetricValue{}, fmt.Errorf("disk checker (%s): total size reported as zero", d.mountPoint)
	}

	usedBytes := stats.totalBytes - stats.freeBytes
	usedPct := float64(usedBytes) / float64(stats.totalBytes) * 100

	gb := func(b uint64) string { return fmt.Sprintf("%.2f", float64(b)/1e9) }

	return metric.MetricValue{
		Name:  d.Name(),
		Value: usedPct,
		Unit:  d.Unit(),
		Meta: map[string]string{
			"mount":    d.mountPoint,
			"total_gb": gb(stats.totalBytes),
			"used_gb":  gb(usedBytes),
			"free_gb":  gb(stats.freeBytes),
		},
	}, nil
}
