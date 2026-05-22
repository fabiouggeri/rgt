package metric

import (
	"sync"
	"time"
)

// MetricTracker keeps the runtime state for a single monitored metric.
type MetricTracker struct {
	mu         sync.Mutex
	config     ThresholdConfig
	state      MetricState
	degradedAt time.Time // when the metric first crossed UpperThreshold
	lastValue  MetricValue
}

// NewMetricTracker creates a tracker initialised to the Healthy state.
func NewMetricTracker(cfg ThresholdConfig) *MetricTracker {
	return &MetricTracker{
		config: cfg,
		state:  STATE_HEALTHY,
	}
}

// transitionResult is what the tracker returns after evaluating a new sample.
type transitionResult struct {
	PreviousState MetricState
	NewState      MetricState
	Value         MetricValue
}

// Evaluate feeds a new measurement into the state machine and returns the
// transition that occurred (if any). It is safe to call from multiple goroutines.
//
// State-machine transitions:
//
//	Healthy   → Degraded   : value > UpperThreshold
//	Degraded  → Healthy    : value ≤ UpperThreshold (spike resolved before grace period)
//	Degraded  → Critical   : value > UpperThreshold AND time since degradedAt ≥ GracePeriod
//	Critical  → Recovering : value < LowerThreshold
//	Recovering→ Healthy    : value ≤ UpperThreshold (cool-down complete)
//	Recovering→ Degraded   : value > UpperThreshold (bounced back up)
func (t *MetricTracker) Evaluate(v MetricValue) transitionResult {
	t.mu.Lock()
	defer t.mu.Unlock()

	prev := t.state
	t.lastValue = v
	now := v.Timestamp

	switch t.state {

	case STATE_HEALTHY:
		if v.Value > t.config.UpperThreshold {
			t.state = STATE_DEGRADED
			t.degradedAt = now
		}

	case STATE_DEGRADED:
		if v.Value <= t.config.UpperThreshold {
			// Spike resolved before grace period elapsed — back to normal.
			t.state = STATE_HEALTHY
		} else if now.Sub(t.degradedAt) >= t.config.GracePeriod {
			// Has been above threshold long enough → Critical.
			t.state = STATE_CRITICAL
		}

	case STATE_CRITICAL:
		if v.Value < t.config.LowerThreshold {
			t.state = STATE_RECOVERING
		}

	case STATE_RECOVERING:
		if v.Value <= t.config.UpperThreshold {
			// Fully recovered.
			t.state = STATE_HEALTHY
		} else if v.Value > t.config.UpperThreshold {
			// Still high — go back to Degraded and restart the grace-period clock.
			t.state = STATE_DEGRADED
			t.degradedAt = now
		}
	}

	return transitionResult{
		PreviousState: prev,
		NewState:      t.state,
		Value:         v,
	}
}

// CurrentState returns the current state without modifying it.
func (t *MetricTracker) CurrentState() MetricState {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state
}

// Snapshot returns a copy of the last recorded value.
func (t *MetricTracker) Snapshot() MetricValue {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastValue
}
