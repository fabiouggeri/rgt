package health

import (
	"rgt-server/log"

	"github.com/shirou/gopsutil/v3/mem"
)

type memoryChecker struct {
	metricChecker
}

const (
	ALERT_MEMORY AlertType = "MEMORY"
)

var _ MetricChecker = &memoryChecker{}

func newMemoryChecker(config HealthConfig) *memoryChecker {
	return &memoryChecker{
		metricChecker: metricChecker{
			config: config,
			alerts: 0,
		},
	}
}

func (m *memoryChecker) Type() AlertType {
	return ALERT_MEMORY
}

func (m *memoryChecker) Alerts() uint {
	return m.alerts
}

func (m *memoryChecker) Check() bool {
	maxAlerts := m.config.HealthMaxMemoryAlerts().Get()
	if maxAlerts == 0 {
		return false
	}
	threshold := m.config.HealthMemThreshold().Get()
	resumeThreshold := m.config.HealthMemResumeThreshold().Get()
	if threshold <= 0 {
		return false
	}
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Debugf("memoryChecker.Check(). Error getting memory usage: %v", err)
		return false
	}
	memUsage := vmStat.UsedPercent
	if memUsage >= threshold {
		m.alerts++
		log.Infof("Memory Check. Memory usage %.1f%% exceeds threshold %.1f%%. Alert increased to %d", memUsage, threshold, m.alerts)
	} else if memUsage <= resumeThreshold {
		m.alerts = 0
		log.Infof("Health checker. Alert cleared: Disk")
	}
	return m.alerts >= uint(maxAlerts)
}

func (m *memoryChecker) Unhealth() bool {
	maxAlerts := m.config.HealthMaxMemoryAlerts().Get()
	return maxAlerts > 0 && m.alerts >= uint(maxAlerts)
}
