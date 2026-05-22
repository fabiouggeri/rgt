package checker

import (
	"fmt"
	"rgt-server/health/metric"
	"runtime"
)

// LoadAverageChecker reports the 1-minute load average as a percentage of the
// number of logical CPUs.  A value of 100 % means the system is fully saturated
// (one unit of work queued per CPU), matching the Unix convention where a load
// of N on an N-CPU machine equals 100 % utilisation of the scheduling capacity.
//
// Platform notes:
//   - Linux  : reads /proc/loadavg — native EWM averages from the kernel scheduler.
//   - macOS  : reads sysctl vm.loadavg — same semantics, native kernel values.
//   - Windows: no native load average exists; the value is approximated by
//     sampling CPU busy-time and applying exponential smoothing with a 60-second
//     time constant.  It is directionally correct but not strictly comparable
//     to Unix load averages.
//
// Platform dispatch at compile time:
//
//	sys_linux.go   → /proc/loadavg
//	sys_darwin.go  → sysctl vm.loadavg
//	sys_windows.go → GetSystemTimes + EMA
type LoadAverageChecker struct {
	numCPU int
}

// NewLoadAverageChecker creates a LoadAverageChecker.
// If numCPU ≤ 0, runtime.NumCPU() is used automatically.
func NewLoadAverageChecker(numCPU int) *LoadAverageChecker {
	if numCPU <= 0 {
		numCPU = runtime.NumCPU()
	}
	return &LoadAverageChecker{
		numCPU: numCPU,
	}
}

func (l *LoadAverageChecker) Name() string {
	return "LOAD_AVERAGE_1M"
}

func (l *LoadAverageChecker) Unit() string {
	return "%"
}

// Check returns the 1-minute load average as a percentage of CPU capacity,
// plus the 5- and 15-minute values in Meta for informational purposes.
func (l *LoadAverageChecker) Check() (metric.MetricValue, error) {
	load1, load5, load15, err := readLoadAvg()
	if err != nil {
		return metric.MetricValue{}, fmt.Errorf("load average checker: %w", err)
	}

	// Percentage: load1=numCPU means 100 % of scheduling capacity is saturated.
	pct := (load1 / float64(l.numCPU)) * 100

	return metric.MetricValue{
		Name:  l.Name(),
		Value: pct,
		Unit:  l.Unit(),
		Meta: map[string]string{
			"load_1m":  fmt.Sprintf("%.2f", load1),
			"load_5m":  fmt.Sprintf("%.2f", load5),
			"load_15m": fmt.Sprintf("%.2f", load15),
			"num_cpu":  fmt.Sprintf("%d", l.numCPU),
		},
	}, nil
}
