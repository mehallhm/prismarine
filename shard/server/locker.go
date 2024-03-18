package server

import (
	"context"
	"errors"
	"sync"
)

var ErrLockerLocked = errors.New("locker: cannot acquire lock, already locked")

type Locker struct {
	sync.RWMutex
	ch chan bool
}

// NewLocker returns a new Locker
func NewLocker() *Locker {
	return &Locker{
		ch: make(chan bool, 1),
	}
}

// IsLocked returns the current state of the locker channel
func (l *Locker) IsLocked() bool {
	l.RLock()
	defer l.RUnlock()
	return len(l.ch) == 1
}

// Acquire will acquire the power lock if it is not currently locked. If it is
// already locked, acquire will fail to acquire the lock, and will return false.
func (l *Locker) Acquire() error {
	l.RLock()
	defer l.RUnlock()
	select {
	case l.ch <- true:
	default:
		return ErrLockerLocked
	}
	return nil
}

// TryAcquire will attempt to acquire a power-lock until the context provided
// is canceled.
func (l *Locker) TryAcquire(ctx context.Context) error {
	select {
	case l.ch <- true:
		return nil
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return ErrLockerLocked
			}
		}
		return nil
	}
}

// Release will drain the locker channel so that we can properly re-acquire it
// at a later time. If the channel is not currently locked this function is a
// no-op and will immediately return.
func (l *Locker) Release() {
	l.Lock()
	defer l.Unlock()
	select {
	case <-l.ch:
	default:
	}
}

// Destroy cleans up the power locker by closing the channel.
func (l *Locker) Destroy() {
	l.Lock()
	defer l.Unlock()
	if l.ch != nil {
		select {
		case <-l.ch:
		default:

		}
		close(l.ch)
	}
}
