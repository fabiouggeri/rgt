//go:build linux

package checker

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// readCPUSample parses /proc/stat to extract aggregate CPU tick counts.
// The kernel exposes cumulative ticks since boot; the caller diffs two
// consecutive samples to derive utilisation over an interval.
func readCPUSample() (rawCPUSample, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return rawCPUSample{}, fmt.Errorf("open /proc/stat: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		// Format: cpu  user nice system idle iowait irq softirq steal guest guest_nice
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return rawCPUSample{}, fmt.Errorf("/proc/stat: unexpected cpu line format")
		}
		u := func(s string) uint64 { v, _ := strconv.ParseUint(s, 10, 64); return v }

		var total uint64
		for _, f := range fields[1:] {
			total += u(f)
		}
		// Idle = idle + iowait (fields[4] + fields[5] when present).
		idle := u(fields[4])
		if len(fields) > 5 {
			idle += u(fields[5]) // iowait
		}
		return rawCPUSample{total: total, idle: idle}, nil
	}
	return rawCPUSample{}, fmt.Errorf("/proc/stat: cpu line not found")
}

// readMemStats parses /proc/meminfo.
// It uses MemAvailable (kernel 3.14+) which accounts for reclaimable
// page-cache, giving a realistic view of memory pressure.
func readMemStats() (rawMemStats, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return rawMemStats{}, fmt.Errorf("open /proc/meminfo: %w", err)
	}
	defer f.Close()

	want := map[string]*uint64{}
	var total, avail uint64
	want["MemTotal"] = &total
	want["MemAvailable"] = &avail

	found := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() && found < len(want) {
		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if dst, ok := want[key]; ok {
			valStr := strings.TrimSpace(parts[1])
			valStr = strings.Fields(valStr)[0] // strip " kB"
			*dst, _ = strconv.ParseUint(valStr, 10, 64)
			found++
		}
	}
	if err := scanner.Err(); err != nil {
		return rawMemStats{}, err
	}
	if total == 0 {
		return rawMemStats{}, fmt.Errorf("/proc/meminfo: MemTotal not found")
	}
	return rawMemStats{totalKB: total, availableKB: avail}, nil
}

// readLoadAvg parses /proc/loadavg and returns the 1, 5, and 15-minute
// exponentially-weighted moving averages of the run-queue length.
func readLoadAvg() (load1, load5, load15 float64, err error) {
	f, err := os.Open("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("open /proc/loadavg: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			return 0, 0, 0, fmt.Errorf("/proc/loadavg: unexpected format")
		}
		parse := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
		load1 = parse(fields[0])
		load5 = parse(fields[1])
		load15 = parse(fields[2])
	}
	return load1, load5, load15, scanner.Err()
}
