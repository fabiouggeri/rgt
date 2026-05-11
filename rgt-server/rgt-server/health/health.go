package health

import (
	"fmt"
	"os"
	"rgt-server/log"
	"rgt-server/option"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type AlertType string

type HealthConfig interface {
	HealthMaxCpuAlerts() option.TypedOption[uint16]
	HealthMaxMemoryAlerts() option.TypedOption[uint16]
	HealthMaxDiskAlerts() option.TypedOption[uint16]
	HealthCpuThreshold() option.TypedOption[float64]
	HealthCpuResumeThreshold() option.TypedOption[float64]
	HealthMemThreshold() option.TypedOption[float64]
	HealthMemResumeThreshold() option.TypedOption[float64]
	HealthDiskThreshold() option.TypedOption[float64]
	HealthDiskResumeThreshold() option.TypedOption[float64]
	HealthCheckInterval() option.TypedOption[time.Duration]
	HealthMaxLoginTime() option.TypedOption[time.Duration]
	HealthMaxLoginsTimeout() option.TypedOption[uint16]
	HealthMaxLoginsTimeoutAlerts() option.TypedOption[uint16]
}

const (
	ALERT_CPU    AlertType = "CPU"
	ALERT_MEMORY AlertType = "MEMORY"
	ALERT_DISK   AlertType = "DISK"
)

// HealthCallbacks provides the interface for the health checker to interact
// with the server without creating an import cycle.
type ServiceCallbacks interface {
	Pause()
	Resume()
}

type HealthChecker struct {
	servicesCallbacks ServiceCallbacks
	config            HealthConfig
	timer             *time.Ticker
	unhealthy         atomic.Bool
	activeAlerts      map[AlertType]uint
	alertsMutex       sync.RWMutex
	diskPath          string
	maxAlerts         map[AlertType]uint
}

func New(config HealthConfig, servicesCallbacks ServiceCallbacks) *HealthChecker {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	h := &HealthChecker{
		servicesCallbacks: servicesCallbacks,
		config:            config,
		activeAlerts:      make(map[AlertType]uint),
		diskPath:          dir,
		maxAlerts:         make(map[AlertType]uint),
	}
	h.maxAlerts[ALERT_CPU] = uint(config.HealthMaxCpuAlerts().Get())
	h.maxAlerts[ALERT_MEMORY] = uint(config.HealthMaxMemoryAlerts().Get())
	h.maxAlerts[ALERT_DISK] = uint(config.HealthMaxDiskAlerts().Get())
	return h
}

func (h *HealthChecker) Start() {
	if h.timer != nil {
		return
	}
	interval := h.config.HealthCheckInterval().Get()
	if interval <= 0 {
		interval = 30 * time.Second
	}
	h.timer = time.NewTicker(interval)
	go h.healthCheckJob()
	log.Info("Health checker started.")
}

func (h *HealthChecker) Stop() {
	if h.timer != nil {
		t := h.timer
		h.timer = nil
		t.Stop()
		log.Info("Health checker stopped.")
	}
}

func (h *HealthChecker) IsHealthy() bool {
	return !h.unhealthy.Load()
}

func (h *HealthChecker) GetAlerts() []AlertType {
	h.alertsMutex.RLock()
	defer h.alertsMutex.RUnlock()
	alerts := make([]AlertType, 0, len(h.activeAlerts))
	for alert := range h.activeAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

func (h *HealthChecker) addAlert(alert AlertType, message string) {
	h.alertsMutex.Lock()
	defer h.alertsMutex.Unlock()
	count := h.activeAlerts[alert]
	count++
	h.activeAlerts[alert] = count
	log.Infof("%s. Alert %s incresead to %d", message, alert, count)
}

func (h *HealthChecker) clearAlert(alert AlertType) {
	h.alertsMutex.Lock()
	defer h.alertsMutex.Unlock()
	if _, found := h.activeAlerts[alert]; found {
		delete(h.activeAlerts, alert)
		log.Infof("Health checker. Alert cleared: %s", alert)
	}
}

func (h *HealthChecker) hasAlerts() bool {
	totalAlerts := uint(0)
	h.alertsMutex.RLock()
	defer h.alertsMutex.RUnlock()
	if len(h.activeAlerts) > 0 {
		for alertType, count := range h.activeAlerts {
			totalAlerts += count
			maxAlerts, found := h.maxAlerts[alertType]
			if found && maxAlerts > 0 && count >= maxAlerts {
				return true
			}
		}
	}
	return false
}

func (h *HealthChecker) healthCheckJob() {
	log.Debug("HealthChecker.healthCheckJob(). started.")
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in HealthChecker.healthCheckJob: %v", err)
		}
	}()
	for range h.timer.C {
		h.checkHealth()
	}
	log.Debug("HealthChecker.healthCheckJob(). stopped.")
}

func (h *HealthChecker) checkHealth() {
	h.checkCPU()
	h.checkMemory()
	h.checkDisk()

	if h.hasAlerts() {
		if !h.unhealthy.Load() {
			h.unhealthy.Store(true)
			log.Infof("Health checker. Server unhealthy. Pausing new connections. Alerts: %s", h.alertsSummary())
			h.servicesCallbacks.Pause()
		}
	} else {
		if h.unhealthy.Load() {
			h.unhealthy.Store(false)
			log.Info("Health checker. Server healthy. Resuming new connections.")
			h.servicesCallbacks.Resume()
		}
	}
}

func (h *HealthChecker) checkCPU() {
	if h.config.HealthMaxCpuAlerts().Get() == 0 {
		return
	}
	threshold := h.config.HealthCpuThreshold().Get()
	resumeThreshold := h.config.HealthCpuResumeThreshold().Get()
	if threshold <= 0 {
		return
	}
	percentages, err := cpu.Percent(0, false)
	if err != nil || len(percentages) == 0 {
		log.Debugf("HealthChecker.checkCPU(). Error getting CPU usage: %v", err)
		return
	}
	cpuUsage := percentages[0]
	if cpuUsage >= threshold {
		h.addAlert(ALERT_CPU, fmt.Sprintf("CPU Check. CPU usage %.1f%% exceeds threshold %.1f%%", cpuUsage, threshold))
	} else if cpuUsage <= resumeThreshold {
		h.clearAlert(ALERT_CPU)
	}
}

func (h *HealthChecker) checkMemory() {
	if h.config.HealthMaxMemoryAlerts().Get() == 0 {
		return
	}
	threshold := h.config.HealthMemThreshold().Get()
	resumeThreshold := h.config.HealthMemResumeThreshold().Get()
	if threshold <= 0 {
		return
	}
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Debugf("HealthChecker.checkMemory(). Error getting memory usage: %v", err)
		return
	}
	memUsage := vmStat.UsedPercent
	if memUsage >= threshold {
		h.addAlert(ALERT_MEMORY, fmt.Sprintf("Memory Check. Memory usage %.1f%% exceeds threshold %.1f%%", memUsage, threshold))
	} else if memUsage <= resumeThreshold {
		h.clearAlert(ALERT_MEMORY)
	}
}

func (h *HealthChecker) checkDisk() {
	if h.config.HealthMaxDiskAlerts().Get() == 0 {
		return
	}
	threshold := h.config.HealthDiskThreshold().Get()
	resumeThreshold := h.config.HealthDiskResumeThreshold().Get()
	if threshold <= 0 {
		return
	}
	usage, err := disk.Usage(h.diskPath)
	if err != nil {
		log.Debugf("HealthChecker.checkDisk(). Error getting disk usage: %v", err)
		return
	}
	diskUsage := usage.UsedPercent
	if diskUsage >= threshold {
		h.addAlert(ALERT_DISK, fmt.Sprintf("Disk Check. Disk usage %.1f%% exceeds threshold %.1f%%", diskUsage, threshold))
	} else if diskUsage <= resumeThreshold {
		h.clearAlert(ALERT_DISK)
	}
}

func (h *HealthChecker) alertsSummary() string {
	alerts := h.GetAlerts()
	if len(alerts) == 0 {
		return "none"
	}
	names := make([]string, len(alerts))
	for i, a := range alerts {
		names[i] = string(a)
	}
	return strings.Join(names, ", ")
}
