package health

import (
	"rgt-server/log"
	"rgt-server/option"
	"strings"
	"sync/atomic"
	"time"
)

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
}

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
	diskPath          string
	metricsCheckers   []MetricChecker
}

func New(config HealthConfig, servicesCallbacks ServiceCallbacks) *HealthChecker {
	h := &HealthChecker{
		servicesCallbacks: servicesCallbacks,
		config:            config,
	}
	h.metricsCheckers = []MetricChecker{
		newCpuChecker(config),
		newMemoryChecker(config),
		newDiskChecker(config),
	}
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

func (h *HealthChecker) hasAlerts() bool {
	for _, checker := range h.metricsCheckers {
		if checker.Alerts() > 0 {
			return true
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
	hasAlerts := false
	for _, metricChecker := range h.metricsCheckers {
		if metricChecker.Check() {
			hasAlerts = true
		}
	}
	if hasAlerts {
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

func (h *HealthChecker) GetAlerts() []AlertType {
	alerts := make([]AlertType, 0)
	for _, checker := range h.metricsCheckers {
		if checker.Unhealth() {
			alerts = append(alerts, checker.Type())
		}
	}
	return alerts
}

func (h *HealthChecker) alertsSummary() string {
	alerts := h.GetAlerts()
	if len(alerts) == 0 {
		return "none"
	}
	names := make([]string, len(alerts))
	for i, alert := range alerts {
		names[i] = string(alert)
	}
	return strings.Join(names, ", ")
}
