package util

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func mustAdd[T comparable](t *testing.T, s *Set[T], elem T) {
	t.Helper()
	if err := s.AddWait(context.Background(), elem, time.Second); err != nil {
		t.Fatalf("AddWait(%v) unexpected error: %v", elem, err)
	}
}

// ---------------------------------------------------------------------------
// Básicos
// ---------------------------------------------------------------------------

func TestAdd_Basic(t *testing.T) {
	s := NewSet[int](5)
	mustAdd(t, s, 1)
	mustAdd(t, s, 2)
	if s.Len() != 2 {
		t.Fatalf("expected len 2, got %d", s.Len())
	}
}

func TestAdd_Duplicate(t *testing.T) {
	s := NewSet[string](5)
	mustAdd(t, s, "hello")
	if err := s.AddWait(context.Background(), "hello", time.Second); err != ErrDuplicate {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
}

func TestContains(t *testing.T) {
	s := NewSet[int](10)
	mustAdd(t, s, 42)
	if !s.Contains(42) {
		t.Fatal("expected Contains(42) == true")
	}
	if s.Contains(99) {
		t.Fatal("expected Contains(99) == false")
	}
}

func TestRemove(t *testing.T) {
	s := NewSet[int](5)
	mustAdd(t, s, 7)
	if !s.Remove(7) {
		t.Fatal("expected Remove to return true")
	}
	if s.Len() != 0 {
		t.Fatal("expected empty set after remove")
	}
	if s.Remove(7) {
		t.Fatal("expected Remove on absent element to return false")
	}
}

func TestClear(t *testing.T) {
	s := NewSet[int](5)
	mustAdd(t, s, 1)
	mustAdd(t, s, 2)
	s.Clear()
	if s.Len() != 0 {
		t.Fatalf("expected empty set, got len=%d", s.Len())
	}
}

// ---------------------------------------------------------------------------
// Limite e desbloqueio
// ---------------------------------------------------------------------------

func TestAdd_Timeout_WhenFull(t *testing.T) {
	s := NewSet[int](2)
	mustAdd(t, s, 1)
	mustAdd(t, s, 2)

	if err := s.AddWait(context.Background(), 3, 100*time.Millisecond); err != ErrTimeout {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}

func TestAdd_UnblockedAfterRemove(t *testing.T) {
	s := NewSet[int](1)
	mustAdd(t, s, 10)

	added := make(chan error, 1)
	go func() { added <- s.AddWait(context.Background(), 20, 2*time.Second) }()

	time.Sleep(50 * time.Millisecond)
	s.Remove(10)

	if err := <-added; err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !s.Contains(20) {
		t.Fatal("expected 20 to be in the set")
	}
}

func TestSetLimit_IncreasesCapacity(t *testing.T) {
	s := NewSet[int](1)
	mustAdd(t, s, 1)

	added := make(chan error, 1)
	go func() { added <- s.AddWait(context.Background(), 2, 2*time.Second) }()

	time.Sleep(50 * time.Millisecond)
	s.SetLimit(5)

	if err := <-added; err != nil {
		t.Fatalf("expected nil after SetLimit, got %v", err)
	}
}

func TestSetLimit_Unlimited(t *testing.T) {
	s := NewSet[int](0)
	for i := range 1000 {
		mustAdd(t, s, i)
	}
	if s.Len() != 1000 {
		t.Fatalf("expected 1000, got %d", s.Len())
	}
}

func TestContextCancellation(t *testing.T) {
	s := NewSet[int](1)
	mustAdd(t, s, 1)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.AddWait(ctx, 2, 5*time.Second) }()

	time.Sleep(30 * time.Millisecond)
	cancel()

	if err := <-done; err == nil {
		t.Fatal("expected error after context cancellation")
	}
}

// ---------------------------------------------------------------------------
// Concorrência com timeout forçado
// ---------------------------------------------------------------------------

// TestConcurrent_Timeout lança N goroutines tentando inserir num conjunto com
// limite L < N. As goroutines excedentes devem receber ErrTimeout; nenhuma
// outra categoria de erro é aceitável.
func TestConcurrent_Timeout(t *testing.T) {
	const (
		limit      = 20
		goroutines = 80
		timeout    = 200 * time.Millisecond
	)

	s := NewSet[int](limit)

	var (
		wg        sync.WaitGroup
		successes atomic.Int64
		timeouts  atomic.Int64
	)

	for i := range goroutines {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			switch err := s.AddWait(context.Background(), v, timeout); err {
			case nil:
				successes.Add(1)
			case ErrTimeout:
				timeouts.Add(1)
			default:
				t.Errorf("goroutine %d: unexpected error: %v", v, err)
			}
		}(i)
	}

	wg.Wait()

	if got := successes.Load(); got != limit {
		t.Errorf("successes: want %d, got %d", limit, got)
	}
	if got := timeouts.Load(); got != goroutines-limit {
		t.Errorf("timeouts: want %d, got %d", goroutines-limit, got)
	}
	if s.Len() != limit {
		t.Errorf("set size: want %d, got %d", limit, s.Len())
	}
}

// TestConcurrent_TimeoutThenDrain testa o cenário dinâmico: metade das
// goroutines entra imediatamente; a outra metade expira por timeout enquanto
// o conjunto permanece cheio. Depois o conjunto é drenado e uma nova rodada
// de inserções deve ter sucesso sem erros.
func TestConcurrent_TimeoutThenDrain(t *testing.T) {
	const (
		limit   = 10
		first   = 10 // preenchem o conjunto
		blocked = 15 // ficam esperando e sofrem timeout
		timeout = 150 * time.Millisecond
	)

	s := NewSet[int](limit)

	// Preenche o conjunto de forma síncrona.
	for i := range first {
		mustAdd(t, s, i)
	}

	// Lança goroutines que vão bloquear e expirar.
	var wg sync.WaitGroup
	var timedOut atomic.Int64

	for i := first; i < first+blocked; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			if err := s.AddWait(context.Background(), v, timeout); err == ErrTimeout {
				timedOut.Add(1)
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}

	wg.Wait()

	if got := timedOut.Load(); got != blocked {
		t.Errorf("expected %d timeouts, got %d", blocked, got)
	}

	// Drena o conjunto e insere uma nova rodada — agora sem bloqueio.
	s.Clear()
	for i := range limit {
		mustAdd(t, s, 1000+i)
	}
	if s.Len() != limit {
		t.Errorf("after drain: want %d elements, got %d", limit, s.Len())
	}
}

// TestConcurrent_TimeoutWhileSlowDrain mistura produtores (adicionam com
// timeout curto) e um consumidor lento (remove um elemento por vez com pausa).
// Valida que os totais de sucesso + timeout == goroutines e que não há deadlock.
func TestConcurrent_TimeoutWhileSlowDrain(t *testing.T) {
	const (
		limit      = 5
		goroutines = 100
		removeEach = 40 * time.Millisecond
		addTimeout = 100 * time.Millisecond
	)

	s := NewSet[int](limit)

	// Consumidor lento: abre uma vaga a cada removeEach.
	stopDrain := make(chan struct{})
	go func() {
		ticker := time.NewTicker(removeEach)
		defer ticker.Stop()
		next := 0
		for {
			select {
			case <-ticker.C:
				s.Remove(next)
				next++
			case <-stopDrain:
				return
			}
		}
	}()

	var (
		wg        sync.WaitGroup
		successes atomic.Int64
		timeouts  atomic.Int64
	)

	for i := range goroutines {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			switch err := s.AddWait(context.Background(), v, addTimeout); err {
			case nil:
				successes.Add(1)
			case ErrTimeout:
				timeouts.Add(1)
			default:
				t.Errorf("goroutine %d: unexpected error: %v", v, err)
			}
		}(i)
	}

	wg.Wait()
	close(stopDrain)

	total := successes.Load() + timeouts.Load()
	if total != goroutines {
		t.Errorf("successes(%d)+timeouts(%d) = %d, want %d",
			successes.Load(), timeouts.Load(), total, goroutines)
	}
	t.Logf("successes=%d timeouts=%d", successes.Load(), timeouts.Load())
}
