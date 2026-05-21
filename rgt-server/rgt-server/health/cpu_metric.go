package health

import (
	"rgt-server/log"

	"github.com/shirou/gopsutil/v3/cpu"
)

type cpuChecker struct {
	metricChecker
}

const (
	ALERT_CPU AlertType = "CPU"
)

var _ MetricChecker = &cpuChecker{}

func newCpuChecker(config HealthConfig) *cpuChecker {
	return &cpuChecker{
		metricChecker: metricChecker{
			config: config,
			alerts: 0,
		},
	}
}

func (c *cpuChecker) Type() AlertType {
	return ALERT_CPU
}

func (c *cpuChecker) Check() bool {
	maxAlerts := c.config.HealthMaxCpuAlerts().Get()
	if maxAlerts == 0 {
		return false
	}
	threshold := c.config.HealthCpuThreshold().Get()
	resumeThreshold := c.config.HealthCpuResumeThreshold().Get()
	if threshold <= 0 {
		return false
	}
	percentages, err := cpu.Percent(0, false)
	if err != nil || len(percentages) == 0 {
		log.Debugf("cpuChecker.Check(). Error getting CPU usage: %v", err)
		return false
	}
	cpuUsage := percentages[0]
	if cpuUsage >= threshold {
		c.alerts++
		log.Infof("CPU Check. CPU usage %.1f%% exceeds threshold %.1f%%. Alert increased to %d", cpuUsage, threshold, c.alerts)
	} else if cpuUsage <= resumeThreshold {
		c.alerts = 0
		log.Infof("Health checker. Alert cleared: CPU")
	}
	return c.alerts >= uint(maxAlerts)
}

func (c *cpuChecker) Unhealth() bool {
	maxAlerts := c.config.HealthMaxCpuAlerts().Get()
	return maxAlerts > 0 && c.alerts >= uint(maxAlerts)
}

func (c *cpuChecker) Alerts() uint {
	return c.alerts
}
