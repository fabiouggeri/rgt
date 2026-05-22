// Package monitor provides a background health-monitoring engine that collects
// host metrics at configurable intervals, tracks state transitions via a
// hysteresis band, and fires stop/restart callbacks when thresholds are breached.
//
// Extending the monitor is as simple as implementing the Checker interface and
// registering the checker with RegisterChecker.
package health

import (
	"context"
	"rgt-server/health/checker"
	"rgt-server/health/metric"
	"rgt-server/log"
	"rgt-server/option"
	"sync"
	"time"
)

type HealthConfig interface {
	HealthCheckInterval() option.TypedOption[time.Duration]
	HealthCpuResumeThreshold() option.TypedOption[float64]
	HealthCpuThreshold() option.TypedOption[float64]
	HealthCpuGracePeriod() option.TypedOption[time.Duration]
	HealthMemThreshold() option.TypedOption[float64]
	HealthMemGracePeriod() option.TypedOption[time.Duration]
	HealthMemResumeThreshold() option.TypedOption[float64]
	HealthDiskThreshold() option.TypedOption[float64]
	HealthDiskResumeThreshold() option.TypedOption[float64]
	HealthDiskGracePeriod() option.TypedOption[time.Duration]
	HealthDiskPath() option.TypedOption[string]
}

// HealthCallbacks provides the interface for the health checker to interact
// with the server without creating an import cycle.
type ServiceCallbacks interface {
	Pause()
	Resume()
}

// CheckerConfig bundles a Checker with its threshold settings.
type CheckerConfig struct {
	Checker   metric.Checker
	Threshold metric.ThresholdConfig
}
type configMetricTracker struct {
	cfg     CheckerConfig
	tracker *metric.MetricTracker
}

// HealthMonitor runs registered Checkers in the background and drives the
// per-metric state machine, firing callbacks on critical events and recoveries.
type HealthMonitor struct {
	interval      time.Duration
	callbacks     ServiceCallbacks
	onStateChange func(prev, next metric.MetricState, value metric.MetricValue)
	mu            sync.RWMutex
	checkers      map[string]CheckerConfig
	trackers      map[string]*metric.MetricTracker
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// New creates and returns a HealthChecker.  Call Start to begin background
// collection. Call RegisterChecker before or after Start (it is safe to add
// checkers while running).
func New(interval time.Duration, callbacks ServiceCallbacks, onStateChange func(prev, next metric.MetricState, value metric.MetricValue)) (*HealthMonitor, error) {
	return &HealthMonitor{
		interval:      interval,
		callbacks:     callbacks,
		onStateChange: onStateChange,
		checkers:      make(map[string]CheckerConfig),
		trackers:      make(map[string]*metric.MetricTracker),
	}, nil
}

func NewDefault(cfg HealthConfig, callbacks ServiceCallbacks) (*HealthMonitor, error) {
	healthChecker, err := New(cfg.HealthCheckInterval().Get(), callbacks, nil)
	if err != nil {
		return nil, err
	}
	if cfg.HealthCpuThreshold().Get() > 0 {
		healthChecker.RegisterChecker(CheckerConfig{
			Checker: checker.NewCPUChecker(),
			Threshold: metric.ThresholdConfig{
				UpperThreshold: cfg.HealthCpuThreshold().Get(),
				LowerThreshold: cfg.HealthCpuResumeThreshold().Get(),
				GracePeriod:    cfg.HealthCpuGracePeriod().Get(),
			},
		})
	}

	if cfg.HealthMemThreshold().Get() > 0 {
		healthChecker.RegisterChecker(CheckerConfig{
			Checker: checker.NewMemoryChecker(),
			Threshold: metric.ThresholdConfig{
				UpperThreshold: cfg.HealthMemThreshold().Get(),
				LowerThreshold: cfg.HealthMemResumeThreshold().Get(),
				GracePeriod:    cfg.HealthMemGracePeriod().Get(),
			},
		})
	}

	if cfg.HealthDiskThreshold().Get() > 0 {
		healthChecker.RegisterChecker(CheckerConfig{
			Checker: checker.NewDiskChecker(cfg.HealthDiskPath().Get()),
			Threshold: metric.ThresholdConfig{
				UpperThreshold: cfg.HealthDiskThreshold().Get(),
				LowerThreshold: cfg.HealthDiskResumeThreshold().Get(),
				GracePeriod:    cfg.HealthDiskGracePeriod().Get(),
			},
		})
	}

	// if cfg.HealthLoadAverageThreshold().Get() > 0 {
	// healthChecker.RegisterChecker(CheckerConfig{
	// 	Checker: checker.NewLoadAverageChecker(0), // 0 → auto-detect
	// 	Threshold: metric.ThresholdConfig{
	// 		UpperThreshold: cfg.HealthLoadAverageThreshold().Get(), // 80.0,
	// 		LowerThreshold: cfg.HealthLoadAverageResumeThreshold().Get(), // 50.0,
	// 		GracePeriod:    cfg.HealthLoadAverageGracePeriod().Get(), // 15 * time.Second,
	// 	},
	// })
	// }

	return healthChecker, nil
}

// RegisterChecker adds (or replaces) a checker at runtime.
// If the monitor is already running the new checker will be picked up on the
// very next tick.
func (h *HealthMonitor) RegisterChecker(cfg CheckerConfig) {
	name := cfg.Checker.Name()
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[name] = cfg
	h.trackers[name] = metric.NewMetricTracker(cfg.Threshold)
	log.Infof("%s checker registered. Lower Threshold: %g%s, Upper Threshold: %g%s, Grace Period: %v",
		name,
		cfg.Threshold.LowerThreshold, cfg.Checker.Unit(),
		cfg.Threshold.UpperThreshold, cfg.Checker.Unit(),
		cfg.Threshold.GracePeriod)
}

// UnregisterChecker removes a checker by name.
func (h *HealthMonitor) UnregisterChecker(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.checkers, name)
	delete(h.trackers, name)
	log.Infof("Checker %s unregistered", name)
}

// Start launches the background collection loop. It is idempotent; calling
// Start on an already-running monitor is a no-op.
func (h *HealthMonitor) Start(ctx context.Context) {
	h.mu.Lock()
	if h.cancel != nil { // already running
		h.mu.Unlock()
		return
	}
	ctx, h.cancel = context.WithCancel(ctx)
	h.mu.Unlock()

	h.wg.Add(1)
	go h.loop(ctx)
	log.Infof("Health monitor started with interval %v", h.interval)
}

// Stop gracefully shuts down the monitor and waits for the background
// goroutine to exit.
func (h *HealthMonitor) Stop() {
	h.mu.Lock()
	if h.cancel == nil {
		h.mu.Unlock()
		return
	}
	h.cancel()
	h.cancel = nil
	h.mu.Unlock()

	h.wg.Wait()
	log.Info("Health monitor stopped")
}

// Snapshot returns a point-in-time copy of every tracked metric value.
func (h *HealthMonitor) Snapshot() map[string]metric.MetricValue {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make(map[string]metric.MetricValue, len(h.trackers))
	for name, t := range h.trackers {
		out[name] = t.Snapshot()
	}
	return out
}

// States returns the current MetricState for every registered checker.
func (h *HealthMonitor) States() map[string]metric.MetricState {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make(map[string]metric.MetricState, len(h.trackers))
	for name, t := range h.trackers {
		out[name] = t.CurrentState()
	}
	return out
}

// ─── internal ────────────────────────────────────────────────────────────────

func (h *HealthMonitor) loop(ctx context.Context) {
	defer h.wg.Done()

	// Run immediately on start, then on every tick.
	h.runAll()

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.runAll()
		}
	}
}

func (h *HealthMonitor) runAll() {
	h.mu.RLock()
	// Copy references so we don't hold the lock during slow I/O.
	items := make([]configMetricTracker, 0, len(h.checkers))
	for name, cfg := range h.checkers {
		items = append(items, configMetricTracker{
			cfg,
			h.trackers[name],
		})
	}
	h.mu.RUnlock()

	for _, item := range items {
		h.runOne(item.cfg, item.tracker)
	}
}

func (h *HealthMonitor) runOne(cfg CheckerConfig, tracker *metric.MetricTracker) {
	value, err := cfg.Checker.Check()
	if err != nil {
		log.Warnf("Checker error. checker: %s error: %v", cfg.Checker.Name(), err)
		return
	}

	value.Timestamp = time.Now()
	result := tracker.Evaluate(value)

	log.Debugf("Metric sampled. metric=%s value=%.2f%s state=%s", value.Name, value.Value, value.Unit, result.NewState)

	// Notify on every state change.
	if result.PreviousState != result.NewState {
		log.Infof("State transition on metric %s from %v to %v. value=%.2f%s",
			value.Name,
			result.PreviousState,
			result.NewState,
			value.Value, value.Unit)

		if h.onStateChange != nil {
			h.onStateChange(result.PreviousState, result.NewState, value)
		}
	}

	// Fire domain callbacks on the important transitions.
	switch {
	case result.PreviousState == metric.STATE_DEGRADED && result.NewState == metric.STATE_CRITICAL:
		log.Warnf("CRITICAL threshold reached on metric %s — firing OnCritical callback. value=%.2f%s upper_threshold=%v grace_period=%v",
			value.Name,
			value.Value, value.Unit,
			cfg.Threshold.UpperThreshold,
			cfg.Threshold.GracePeriod)
		h.callbacks.Pause()

	case result.PreviousState == metric.STATE_CRITICAL && result.NewState == metric.STATE_RECOVERING:
		log.Infof("Metric %s recovered — firing OnRecovery callback. value=%.2f%s lower_threshold=%v",
			value.Name,
			value.Value, value.Unit,
			cfg.Threshold.LowerThreshold)
		h.callbacks.Resume()
	}
}
