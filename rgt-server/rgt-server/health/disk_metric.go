package health

import (
	"os"
	"rgt-server/log"

	"github.com/shirou/gopsutil/v3/disk"
)

type diskChecker struct {
	metricChecker
	diskPath string
}

const (
	ALERT_DISK AlertType = "DISK"
)

var _ MetricChecker = &diskChecker{}

func newDiskChecker(config HealthConfig) *diskChecker {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	return &diskChecker{
		metricChecker: metricChecker{
			config: config,
			alerts: 0,
		},
		diskPath: dir,
	}
}

func (d *diskChecker) Type() AlertType {
	return ALERT_DISK
}

func (d *diskChecker) Alerts() uint {
	return d.alerts
}

func (d *diskChecker) Check() bool {
	maxAlerts := d.config.HealthMaxDiskAlerts().Get()
	if maxAlerts == 0 {
		return false
	}
	threshold := d.config.HealthDiskThreshold().Get()
	resumeThreshold := d.config.HealthDiskResumeThreshold().Get()
	if threshold <= 0 {
		return false
	}
	usage, err := disk.Usage(d.diskPath)
	if err != nil {
		log.Debugf("diskChecker.Check(). Error getting disk usage: %v", err)
		return false
	}
	diskUsage := usage.UsedPercent
	if diskUsage >= threshold {
		d.alerts++
		log.Infof("Disk Check. Disk usage %.1f%% exceeds threshold %.1f%%. Alert increased to %d", diskUsage, threshold, d.alerts)
	} else if diskUsage <= resumeThreshold {
		d.alerts = 0
		log.Infof("Health checker. Alert cleared: Disk")
	}
	return d.alerts >= uint(maxAlerts)
}

func (d *diskChecker) Unhealth() bool {
	maxAlerts := d.config.HealthMaxDiskAlerts().Get()
	return maxAlerts > 0 && d.alerts >= uint(maxAlerts)
}
