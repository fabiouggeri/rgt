package metric

import "time"

// MetricState represents the current state of a monitored metric.
type MetricState int

const (
	// STATE_HEALTHY: metric is within normal operating range.
	STATE_HEALTHY MetricState = iota
	// STATE_DEGRADED: metric exceeded the upper threshold but hasn't held for long enough
	// to trigger a critical callback yet.
	STATE_DEGRADED
	// STATE_CRITICAL: metric has been above the upper threshold for longer than GracePeriod,
	// the stop callback has been fired.
	STATE_CRITICAL
	// STATE_RECOVERING: metric dropped below the lower threshold while in Critical state;
	// the restart callback has been fired.
	STATE_RECOVERING
)

func (s MetricState) String() string {
	switch s {
	case STATE_HEALTHY:
		return "Healthy"
	case STATE_DEGRADED:
		return "Degraded"
	case STATE_CRITICAL:
		return "Critical"
	case STATE_RECOVERING:
		return "Recovering"
	default:
		return "Unknown"
	}
}

// MetricValue holds a single measurement result returned by a Checker.
type MetricValue struct {
	Name      string            // Name of the metric (e.g. "cpu", "memory", "disk:/").
	Value     float64           // Value is the measured percentage (0–100).
	Unit      string            // Unit is a human-readable label (e.g. "%", "MB").
	Timestamp time.Time         // Timestamp when the measurement was taken.
	Meta      map[string]string // Meta carries optional extra key/value pairs provided by the checker.
}

// ThresholdConfig defines the hysteresis band and grace period for a metric.
//
//	┌────────────────────────────────────────────────────────────┐
//	│  Value                                                     │
//	│   100 ─────────────────────────────────────────────────── │
//	│        ▲  above UpperThreshold for ≥ GracePeriod          │
//	│        │  → OnCritical callback fires                      │
//	│ Upper ─┼─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
//	│        │  (Degraded zone)                                  │
//	│ Lower ─┼─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
//	│        │  below LowerThreshold (while Critical)            │
//	│        ▼  → OnRecovery callback fires                      │
//	│     0 ─────────────────────────────────────────────────── │
//	└────────────────────────────────────────────────────────────┘
type ThresholdConfig struct {
	UpperThreshold float64       // if Value exceeds this, the metric enters Degraded state.
	LowerThreshold float64       // if Value falls below this while Critical, it triggers recovery.
	GracePeriod    time.Duration // how long the metric must stay above UpperThreshold before OnCritical is called. Prevents flapping on momentary spikes.
}

// CallbackFunc is the signature for stop/restart callbacks.
// metricName is the name of the metric that triggered the event.
// value is the measurement that caused the state transition.
type CallbackFunc func(metricName string, value MetricValue)

// Checker is the extension point: implement this interface to add any new metric.
type Checker interface {
	Name() string                //returns the unique identifier for this checker (used as map key and log label).
	Check() (MetricValue, error) // Check performs the measurement and returns a MetricValue or an error.
	Unit() string                // Unit returns a human-readable label for the metric's unit (e.g. "%", "MB").
}
