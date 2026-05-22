// Package checkers provides ready-to-use monitor.Checker implementations for
// the most common host metrics (CPU, memory, disk, load average).
//
// Architecture — separating interface from implementation:
//
//	┌──────────────────────────────────────────────────────────────────┐
//	│  cpu.go / memory.go / disk.go / load_average.go                 │
//	│  Platform-agnostic Checker structs.  They call readXxx() which   │
//	│  returns the raw types defined in this file.                     │
//	└───────────┬───────────────────────────────────────────────────── ┘
//	            │ calls
//	   ┌────────┴──────────────────────────────────────────┐
//	   │  sys_linux.go    //go:build linux                 │
//	   │  sys_darwin.go   //go:build darwin                │
//	   │  sys_windows.go  //go:build windows               │
//	   │  sys_disk_unix.go    //go:build linux || darwin   │
//	   │  sys_disk_windows.go //go:build windows           │
//	   └────────────────────────────────────────────────── ┘
//
// Adding a new metric is always the same two steps:
//  1. Create a Checker struct in a new file (e.g. network.go) that calls a
//     readNetworkStats() function whose signature matches every platform.
//  2. Implement readNetworkStats() in sys_linux.go, sys_darwin.go, sys_windows.go.
package checker

// rawCPUSample holds a snapshot of cumulative CPU ticks split into
// "busy" and "idle" buckets.  The checker diffs two consecutive samples
// to compute utilisation — the same method used by top/htop/Task Manager.
type rawCPUSample struct {
	// total is the sum of all CPU tick categories (user+nice+sys+idle+…).
	total uint64
	// idle includes idle ticks and, on Linux, iowait ticks.
	idle uint64
}

// rawMemStats is a snapshot of physical memory in kilobytes.
type rawMemStats struct {
	// totalKB is the total amount of physical RAM installed.
	totalKB uint64
	// availableKB is the amount that can be given to processes without
	// swapping (includes reclaimable page cache on Linux, reports the
	// OS-provided "available" figure on macOS and Windows).
	availableKB uint64
}

// rawDiskStats holds the raw capacity figures for a single mount point.
type rawDiskStats struct {
	totalBytes uint64
	freeBytes  uint64
}
