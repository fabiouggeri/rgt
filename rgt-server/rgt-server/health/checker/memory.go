package checker

import (
	"fmt"
	"rgt-server/health/metric"
)

// MemoryChecker measures used physical RAM as a percentage of total RAM.
//
// "Used" is defined as total − available, where "available" is whatever the
// OS considers immediately usable without swapping (this includes reclaimable
// page-cache on Linux, free+inactive pages on macOS, and UllAvailPhys on
// Windows).  This gives a realistic view of actual memory pressure rather than
// the raw free-pages figure.
//
// Platform dispatch at compile time:
//
//	sys_linux.go   → /proc/meminfo  (MemTotal, MemAvailable)
//	sys_darwin.go  → sysctl hw.memsize, vm.page_free_count, vm.page_inactive_count
//	sys_windows.go → GlobalMemoryStatusEx
type MemoryChecker struct{}

// NewMemoryChecker creates a MemoryChecker.
func NewMemoryChecker() *MemoryChecker {
	return &MemoryChecker{}
}

func (m *MemoryChecker) Name() string {
	return "MEM"
}

func (m *MemoryChecker) Unit() string {
	return "%"
}

// Check reads physical memory statistics and returns usage as a percentage.
func (m *MemoryChecker) Check() (metric.MetricValue, error) {
	stats, err := readMemStats()
	if err != nil {
		return metric.MetricValue{}, fmt.Errorf("memory checker: %w", err)
	}

	if stats.totalKB == 0 {
		return metric.MetricValue{}, fmt.Errorf("memory checker: total memory reported as zero")
	}

	usedKB := stats.totalKB - stats.availableKB
	usedPct := float64(usedKB) / float64(stats.totalKB) * 100

	return metric.MetricValue{
		Name:  m.Name(),
		Value: usedPct,
		Unit:  m.Unit(),
		Meta: map[string]string{
			"total_kb":     fmt.Sprintf("%d", stats.totalKB),
			"available_kb": fmt.Sprintf("%d", stats.availableKB),
			"used_kb":      fmt.Sprintf("%d", usedKB),
		},
	}, nil
}
