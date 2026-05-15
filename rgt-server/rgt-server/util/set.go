package util

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Erros sentinela exportados para facilitar checagem pelo caller.
var (
	ErrDuplicate = errors.New("element already exists in set")
	ErrTimeout   = errors.New("timeout: set is full and limit was not raised in time")
	ErrInUse     = errors.New("set is in use")
	ErrIsFull    = errors.New("set is full")
)

// Set generic set with thread-safe and limit of capacity.
type Set[T comparable] struct {
	mu      sync.RWMutex
	items   map[T]struct{}
	limit   uint
	notFull *sync.Cond // signal to free space is available or limit increase
}

// New creates a new Set with the initial limit informed.
// limit <= 0 means no limit (unlimited capacity).
func NewSet[T comparable](limit uint) *Set[T] {
	s := &Set[T]{
		items: make(map[T]struct{}),
		limit: limit,
	}
	s.notFull = sync.NewCond(&s.mu)
	return s
}

// SetLimit updates the limit of elements of the set.
// limit <= 0 removes the limit.
// If the new limit is greater than the previous one, goroutines blocked in Add
// are notified immediately.
func (s *Set[T]) SetLimit(limit uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limit = limit
	// Broadcasts waiters; they will re-evaluate the condition.
	s.notFull.Broadcast()
}

// Limit returns the current limit (0 means unlimited).
func (s *Set[T]) Limit() uint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.limit
}

// Len returns the number of elements in the set.
func (s *Set[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// Contains returns if the element exists in the set.
func (s *Set[T]) Contains(elem T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.items[elem]
	return ok
}

// Add inserts an element into the set immediately, waiting if necessary for space.
//
//   - If the set is in use (another goroutine is adding an element), the method returns ErrInUse.
//   - If the set is not full, the insertion is immediate.
//   - If it is full, the method returns ErrIsFull.
//   - If the element already exists, the method returns ErrDuplicate.
func (s *Set[T]) Add(elem T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		if _, exists := s.items[elem]; exists {
			s.notFull.Signal()
			return ErrDuplicate
		}
		if s.limit > 0 && uint(len(s.items)) >= s.limit {
			s.notFull.Wait()
			continue
		}
		s.items[elem] = struct{}{}
		return nil
	}
}

// AddWait inserts an element into the set respecting the capacity limit.
//
//   - If the set is not full, the insertion is immediate.
//   - If it is full, the method blocks until there is space, the context
//     is canceled, or the timeout expires.
//   - Returns ErrDuplicate if the element already exists.
//   - Returns ErrTimeout if the timeout expires before there is space.
//   - Returns ctx.Err() if the context is canceled/expired.
func (s *Set[T]) AddWait(ctx context.Context, elem T, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Registra um único callback: quando ctx encerrar (timeout ou cancelamento
	// externo), Broadcast acorda todos os Waits deste Add.
	// stopAfterFunc() cancela o callback se Add retornar antes — evita um
	// Broadcast desnecessário após inserção bem-sucedida.
	stopAfterFunc := context.AfterFunc(ctx, func() {
		s.notFull.Broadcast()
	})
	defer stopAfterFunc()

	s.mu.Lock()
	defer s.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return resolveErr(ctx)
		default:
		}

		if _, exists := s.items[elem]; exists {
			return ErrDuplicate
		}

		if s.limit <= 0 || uint(len(s.items)) < s.limit {
			s.items[elem] = struct{}{}
			return nil
		}

		s.notFull.Wait()
	}
}

func resolveErr(ctx context.Context) error {
	if cause := context.Cause(ctx); cause != nil && !errors.Is(cause, ctx.Err()) {
		return cause
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ErrTimeout
	}
	return ctx.Err()
}

// Remove deletes an element from the set.
// Returns false if the element was not present.
func (s *Set[T]) Remove(elem T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[elem]; !ok {
		return false
	}
	delete(s.items, elem)
	s.notFull.Signal() // notify a goroutine blocked in Add
	return true
}

// Items returns a copy of the elements in the set.
func (s *Set[T]) Items() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]T, 0, len(s.items))
	for k := range s.items {
		result = append(result, k)
	}
	return result
}

// Clear remove todos os elementos do conjunto.
func (s *Set[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[T]struct{})
	s.notFull.Broadcast()
}
