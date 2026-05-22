package checker

import (
	"fmt"
	"rgt-server/health/metric"
	"sync"
	"time"
)

// CPUChecker measures overall CPU utilisation by computing the ratio of busy
// ticks to total ticks between two consecutive samples.
//
// It is platform-agnostic: the actual sampling is delegated to readCPUSample(),
// whose implementation is selected at compile time by Go build tags:
//
//	sys_linux.go   → /proc/stat
//	sys_darwin.go  → sysctl kern.cp_time
//	sys_windows.go → GetSystemTimes
type CPUChecker struct {
	mu         sync.Mutex
	prevSample rawCPUSample
	prevAt     time.Time
}

// NewCPUChecker creates a CPUChecker and takes an initial baseline reading so
// that the very first call to Check() returns a meaningful value.
func NewCPUChecker() *CPUChecker {
	c := &CPUChecker{}
	sample, _ := readCPUSample() // baseline; error silently ignored
	c.prevSample = sample
	c.prevAt = time.Now()
	return c
}

func (c *CPUChecker) Name() string {
	return "CPU"
}

func (c *CPUChecker) Unit() string {
	return "%"
}

// Check samples the CPU tick counters, diffs them against the previous sample,
// and returns usage as a percentage in [0, 100].
func (c *CPUChecker) Check() (metric.MetricValue, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cur, err := readCPUSample()
	if err != nil {
		return metric.MetricValue{}, fmt.Errorf("cpu checker: %w", err)
	}

	deltaTot := float64(cur.total - c.prevSample.total)
	deltaIdle := float64(cur.idle - c.prevSample.idle)

	var usage float64
	if deltaTot > 0 {
		usage = (1 - deltaIdle/deltaTot) * 100
	}

	now := time.Now()
	c.prevSample = cur
	c.prevAt = now

	return metric.MetricValue{
		Name:  c.Name(),
		Value: usage,
		Unit:  c.Unit(),
		Meta: map[string]string{
			"sampled_at": now.Format(time.RFC3339),
		},
	}, nil
}
