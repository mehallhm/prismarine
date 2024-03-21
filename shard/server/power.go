package server

import (
	"context"
	"fmt"
	"prismarine/shard/runtime"
	"time"

	"github.com/charmbracelet/log"
	"github.com/pkg/errors"
)

type PowerAction string

// The power actions that can be performed for a given server. This taps into the given server
// environment and performs them in a way that prevents a race condition from occurring. For
// example, sending two "start" actions back to back will not process the second action until
// the first action has been completed.
//
// This utilizes a workerpool with a limit of one worker so that all the actions execute
// in a sync manner.
const (
	PowerActionStart   = "start"
	PowerActionStop    = "stop"
	PowerActionRestart = "restart"
	PowerActionKill    = "kill"
)

// IsValid checks if the power action being received is valid
func (pa PowerAction) IsValid() bool {
	return pa == PowerActionStart ||
		pa == PowerActionStop ||
		pa == PowerActionRestart ||
		pa == PowerActionKill
}

func (pa PowerAction) IsStart() bool {
	return pa == PowerActionStart || pa == PowerActionRestart
}

// ExecutingPowerAction checks if there is currently a power action
// being processed for the server
func (s *Server) ExecutingPowerAction() bool {
	return s.powerLock.IsLocked()
}

// HandlePowerAction is a helper function that can receive a power action and then process the
// actions that need to occur for it. This guards against someone calling Start() twice at the
// same time, or trying to restart while another restart process is currently running.
//
// However, the code design for the daemon does depend on the user correctly calling this
// function rather than making direct calls to the start/stop/restart functions on the
// runtime struct.
func (s *Server) HandlePowerAction(action PowerAction, waitSeconds ...int) error {
	if s.installing.Load() {
		return errors.New("server is installing")
	} else if s.restarting.Load() {
		return errors.New("server is restarting")
	} else if s.transferring.Load() {
		return errors.New("server is transferring")
	}

	// lockId, _ := uuid.NewUUID()
	// Server logging
	// log := s.Log().WithField("lock_id", lockId.String()). ...

	cleanup := func() {
		log.Debug("releasing exclusive lock for power action")
		s.powerLock.Release()
	}

	var wait int
	if len(waitSeconds) > 0 && waitSeconds[0] > 0 {
		wait = waitSeconds[0]
	}

	log.Debug("wait_seconds:", waitSeconds, "acquiring power action lock")

	if action == PowerActionKill {
		// Still try to acquire the lock if terminating, and it is available, just so that
		// other power actions are blocked until it has completed. However, if it cannot be
		// acquired we won't stop the entire process.
		//
		// If we did successfully acquire the lock, make sure we release it once we're done
		// execution the power actions.
		if err := s.powerLock.Acquire(); err != nil {
			log.Debug("acquired exclusive lock on power actions, processing event...")
			defer cleanup()
		} else {
			log.Warn("failed to acquire exclusive lock, ignoring failure for termination event")
		}
	} else {
		// Determines if we should wait for the lock or not. If a value greater than 0 is passed
		// into this function we will wait that long for a lock to be acquired.
		if wait == 0 {
			// If no wait duration was provided we will attempt to immediately acquire the lock
			// and bail out with a context deadline error if it is not acquired immediately.
			if err := s.powerLock.Acquire(); err != nil {
				return errors.Wrap(err, "failed to acquire exclusive power lock")
			}
		} else {
			ctx, cancel := context.WithTimeout(s.ctx, time.Second*time.Duration(wait))
			defer cancel()

			// Attempt to acquire a lock on the power action lock for up to 30 seconds. If more
			// time than that passes an error will be propagated back up the chain and this
			// request will be aborted.
			if err := s.powerLock.TryAcquire(ctx); err != nil {
				return errors.Wrap(err, fmt.Sprintf("could not acquire lock on power action after %d seconds", wait))
			}
		}
		log.Debug("acquired lock on power actions, processing event...")
		defer cleanup()
	}

	switch action {
	case PowerActionStart:
		if s.instance.State() != runtime.ProcessOfflineState {
			return errors.New("already running")
		}

		if err := s.Prelude(); err != nil {
			return err
		}

		return s.instance.Start(s.Context())
	case PowerActionStop:
		fallthrough
	case PowerActionRestart:
		if err := s.instance.WaitForStop(s.Context(), time.Second*10, true); err != nil {
			return err
		}

		if action == PowerActionStop {
			return nil
		}

		if err := s.Prelude(); err != nil {
			return err
		}

		return s.instance.Start(s.Context())
	case PowerActionKill:
		panic("not implemented")
	}

	return errors.New("attempting to handle unknown power action")
}

// Runs before the
func (s *Server) Prelude() error {
	return nil
}
