package health

import (
	"context"
	"rgt-server/health/metric"
	"sync"
	"testing"
	"time"
)

// ─── stub checker ────────────────────────────────────────────────────────────

type stubChecker struct {
	mu   sync.Mutex
	name string
	val  float64
}

type callbackInterface struct{}

func (s *stubChecker) Name() string { return s.name }
func (s *stubChecker) Unit() string { return "%" }
func (s *stubChecker) Check() (metric.MetricValue, error) {
	s.mu.Lock()
	v := s.val
	s.mu.Unlock()
	return metric.MetricValue{Name: s.name, Value: v, Unit: "%", Timestamp: time.Now()}, nil
}

func (s *stubChecker) set(v float64) {
	s.mu.Lock()
	s.val = v
	s.mu.Unlock()
}

func (c *callbackInterface) Pause() {
}

func (c *callbackInterface) Resume() {
}

// ─── helper ──────────────────────────────────────────────────────────────────

func buildMon(t *testing.T, interval time.Duration, c *callbackInterface, onChange func(prev, next metric.MetricState, v metric.MetricValue)) *HealthMonitor {
	t.Helper()
	m, err := New(interval, c, onChange)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m
}

// ─── tests ───────────────────────────────────────────────────────────────────

// TestHealthyToNoCritical verifies that a value below upper threshold never
// fires the OnCritical callback.
func TestHealthyToNoCritical(t *testing.T) {
	critFired := false
	m := buildMon(t, 20*time.Millisecond, &callbackInterface{},
		func(prev, next metric.MetricState, v metric.MetricValue) {
			if prev != metric.STATE_CRITICAL && next == metric.STATE_CRITICAL {
				critFired = true
			}
		})
	chk := &stubChecker{name: "cpu", val: 50}
	m.RegisterChecker(CheckerConfig{
		Checker: chk,
		Threshold: metric.ThresholdConfig{
			UpperThreshold: 80, LowerThreshold: 50, GracePeriod: 50 * time.Millisecond,
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)
	time.Sleep(120 * time.Millisecond)
	cancel()
	m.Stop()

	if critFired {
		t.Error("OnCritical should not fire when value stays below upper threshold")
	}
}

// TestCriticalCallbackFires verifies OnCritical fires after grace period.
func TestCriticalCallbackFires(t *testing.T) {
	critCh := make(chan string, 1)
	m := buildMon(t, 20*time.Millisecond, &callbackInterface{},
		func(prev, next metric.MetricState, v metric.MetricValue) {
			if prev != metric.STATE_CRITICAL && next == metric.STATE_CRITICAL {
				critCh <- v.Name
			}
		})
	chk := &stubChecker{name: "mem", val: 95}
	m.RegisterChecker(CheckerConfig{
		Checker: chk,
		Threshold: metric.ThresholdConfig{
			UpperThreshold: 80, LowerThreshold: 50, GracePeriod: 60 * time.Millisecond,
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)
	defer cancel()
	defer m.Stop()

	select {
	case name := <-critCh:
		if name != "mem" {
			t.Errorf("expected metric 'mem', got %q", name)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("OnCritical was not called within timeout")
	}
}

// TestRecoveryCallbackFires verifies OnRecovery fires after metric drops below lower.
func TestRecoveryCallbackFires(t *testing.T) {
	recoverCh := make(chan string, 1)
	m := buildMon(t, 20*time.Millisecond, &callbackInterface{},
		func(prev, next metric.MetricState, v metric.MetricValue) {
			if prev != metric.STATE_CRITICAL && next == metric.STATE_CRITICAL {
				recoverCh <- v.Name
			}
		})
	chk := &stubChecker{name: "disk", val: 95}
	m.RegisterChecker(CheckerConfig{
		Checker: chk,
		Threshold: metric.ThresholdConfig{
			UpperThreshold: 80, LowerThreshold: 50, GracePeriod: 40 * time.Millisecond,
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)
	defer cancel()
	defer m.Stop()

	// Wait for it to go Critical first.
	time.Sleep(200 * time.Millisecond)
	// Now drop below lower threshold.
	chk.set(30)

	select {
	case name := <-recoverCh:
		if name != "disk" {
			t.Errorf("expected metric 'disk', got %q", name)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("OnRecovery was not called within timeout")
	}
}

// TestCriticalNotFiredBeforeGrace verifies that spikes shorter than GracePeriod
// do NOT trigger OnCritical.
func TestCriticalNotFiredBeforeGrace(t *testing.T) {
	critFired := false
	m := buildMon(t, 20*time.Millisecond, &callbackInterface{},
		func(prev, next metric.MetricState, v metric.MetricValue) {
			if prev != metric.STATE_CRITICAL && next == metric.STATE_CRITICAL {
				critFired = true
			}
		})
	chk := &stubChecker{name: "cpu", val: 90}
	m.RegisterChecker(CheckerConfig{
		Checker: chk,
		Threshold: metric.ThresholdConfig{
			UpperThreshold: 80, LowerThreshold: 50, GracePeriod: 500 * time.Millisecond,
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)

	// Spike for a short time, then resolve.
	time.Sleep(80 * time.Millisecond)
	chk.set(30)
	time.Sleep(100 * time.Millisecond)
	cancel()
	m.Stop()

	if critFired {
		t.Error("OnCritical must NOT fire when spike resolves before GracePeriod")
	}
}

// TestRegisterCheckerAtRuntime verifies checkers can be added while running.
func TestRegisterCheckerAtRuntime(t *testing.T) {
	critCh := make(chan string, 4)
	m := buildMon(t, 20*time.Millisecond, &callbackInterface{},
		func(prev, next metric.MetricState, v metric.MetricValue) {
			if prev != metric.STATE_CRITICAL && next == metric.STATE_CRITICAL {
				critCh <- v.Name
			}
		})
	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)
	defer cancel()
	defer m.Stop()

	// Register a checker AFTER Start.
	time.Sleep(30 * time.Millisecond)
	chk := &stubChecker{name: "load", val: 99}
	m.RegisterChecker(CheckerConfig{
		Checker: chk,
		Threshold: metric.ThresholdConfig{
			UpperThreshold: 80, LowerThreshold: 50, GracePeriod: 30 * time.Millisecond,
		},
	})

	select {
	case name := <-critCh:
		if name != "load" {
			t.Errorf("expected 'load', got %q", name)
		}
	case <-time.After(600 * time.Millisecond):
		t.Fatal("OnCritical not fired for dynamically registered checker")
	}
}
