package health

import (
	"os"
	"rgt-server/config"
	"rgt-server/log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type AlertType string

const (
	ALERT_CPU           AlertType = "CPU"
	ALERT_MEMORY        AlertType = "MEMORY"
	ALERT_DISK          AlertType = "DISK"
	ALERT_PENDING_LOGIN AlertType = "PENDING_LOGIN"
)

// PendingSession holds minimal info about a session pending login
type PendingSession struct {
	Id        int64
	StartTime time.Time
}

// HealthCallbacks provides the interface for the health checker to interact
// with the server without creating an import cycle.
type HealthCallbacks interface {
	PauseConnections()
	ResumeConnections()
	GetPendingLoginSessions() []PendingSession
}

type HealthChecker struct {
	callbacks    HealthCallbacks
	config       *config.ServerConfig
	timer        *time.Ticker
	unhealthy    atomic.Bool
	activeAlerts map[AlertType]uint
	alertsMutex  sync.RWMutex
	diskPath     string
	maxAlerts    map[AlertType]uint
}

func New(config *config.ServerConfig, callbacks HealthCallbacks) *HealthChecker {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	h := &HealthChecker{
		callbacks:    callbacks,
		config:       config,
		activeAlerts: make(map[AlertType]uint),
		diskPath:     dir,
		maxAlerts:    make(map[AlertType]uint),
	}
	h.maxAlerts[ALERT_CPU] = uint(config.MaxCpuAlerts().Get())
	h.maxAlerts[ALERT_MEMORY] = uint(config.MaxMemoryAlerts().Get())
	h.maxAlerts[ALERT_DISK] = uint(config.MaxDiskAlerts().Get())
	h.maxAlerts[ALERT_PENDING_LOGIN] = uint(config.MaxPendingLoginsAlerts().Get())
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
	log.Info("HealthChecker started.")
}

func (h *HealthChecker) Stop() {
	if h.timer != nil {
		t := h.timer
		h.timer = nil
		t.Stop()
		log.Info("HealthChecker stopped.")
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

func (h *HealthChecker) addAlert(alert AlertType) {
	h.alertsMutex.Lock()
	defer h.alertsMutex.Unlock()
	count, found := h.activeAlerts[alert]
	if found {
		h.activeAlerts[alert] = count + 1
	} else {
		h.activeAlerts[alert] = 1
	}
	log.Infof("HealthChecker. Alert incresead: %s", alert)
}

func (h *HealthChecker) clearAlert(alert AlertType) {
	h.alertsMutex.Lock()
	defer h.alertsMutex.Unlock()
	if _, found := h.activeAlerts[alert]; found {
		delete(h.activeAlerts, alert)
		log.Infof("HealthChecker. Alert cleared: %s", alert)
	}
}

func (h *HealthChecker) hasAlerts() bool {
	h.alertsMutex.RLock()
	defer h.alertsMutex.RUnlock()
	if len(h.activeAlerts) > 0 {
		for alertType, count := range h.activeAlerts {
			maxAlerts, found := h.maxAlerts[alertType]
			if !found || count > maxAlerts {
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
	h.checkPendingLogins()

	if h.hasAlerts() {
		if !h.unhealthy.Load() {
			h.unhealthy.Store(true)
			log.Infof("HealthChecker. Server unhealthy. Pausing new connections. Alerts: %s", h.alertsSummary())
			h.callbacks.PauseConnections()
		}
	} else {
		if h.unhealthy.Load() {
			h.unhealthy.Store(false)
			log.Info("HealthChecker. Server healthy. Resuming new connections.")
			h.callbacks.ResumeConnections()
		}
	}
}

func (h *HealthChecker) checkCPU() {
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
		h.addAlert(ALERT_CPU)
		log.Debugf("HealthChecker.checkCPU(). CPU usage %.1f%% exceeds threshold %.1f%%", cpuUsage, threshold)
	} else if cpuUsage <= resumeThreshold {
		h.clearAlert(ALERT_CPU)
	}
}

func (h *HealthChecker) checkMemory() {
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
		h.addAlert(ALERT_MEMORY)
		log.Debugf("HealthChecker.checkMemory(). Memory usage %.1f%% exceeds threshold %.1f%%", memUsage, threshold)
	} else if memUsage <= resumeThreshold {
		h.clearAlert(ALERT_MEMORY)
	}
}

func (h *HealthChecker) checkDisk() {
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
		h.addAlert(ALERT_DISK)
		log.Debugf("HealthChecker.checkDisk(). Disk usage %.1f%% exceeds threshold %.1f%%", diskUsage, threshold)
	} else if diskUsage <= resumeThreshold {
		h.clearAlert(ALERT_DISK)
	}
}

func (h *HealthChecker) checkPendingLogins() {
	timeout := h.config.HealthPendingLoginTimeout().Get()
	maxPending := h.config.HealthMaxPendingLogins().Get()
	if timeout <= 0 && maxPending <= 0 {
		return
	}
	sessions := h.callbacks.GetPendingLoginSessions()
	pendingCount := uint16(len(sessions))
	now := time.Now()
	for _, session := range sessions {
		if timeout > 0 && now.Sub(session.StartTime) > timeout {
			h.addAlert(ALERT_PENDING_LOGIN)
			log.Debugf("HealthChecker.checkPendingLogins(). Session %d pending login for %v exceeds timeout %v", session.Id, now.Sub(session.StartTime), timeout)
			return
		}
	}
	if maxPending > 0 && pendingCount > maxPending {
		h.addAlert(ALERT_PENDING_LOGIN)
		log.Debugf("HealthChecker.checkPendingLogins(). %d pending logins exceeds max %d", pendingCount, maxPending)
		return
	}
	h.clearAlert(ALERT_PENDING_LOGIN)
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
