package docker

import (
	"context"
	"fmt"
	"os"
	"prismarine/shard/runtime"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

func (i *Instance) Preflight(ctx context.Context) error {
	// Always destroy and re-create the server container
	if err := i.client.ContainerRemove(ctx, i.Cfg.Uuid, container.RemoveOptions{}); err != nil {
		if !client.IsErrNotFound(err) {
			return errors.Wrap(err, "runtime/docker: failed to remove container")
		}
	}

	// The Create() function will check if the container exists in the first place, and if
	// so just silently return without an error. Otherwise, it will try to create the necessary
	// container and data storage directory.
	//
	// This won't actually run an installation process however, it is just here to ensure the
	// runtime gets created properly if it is missing and the server is started. We're making
	// an assumption that all the files will still exist at this point.
	if err := i.Create(); err != nil {
		return err
	}

	return nil
}

func (i *Instance) Start(ctx context.Context, skipLock bool, waitSeconds int) error {
	log.Debug("starting instance...")

	cleanup, err := i.AttemptPowerlock(ctx, skipLock, waitSeconds)
	if err != nil {
		return err
	}
	defer cleanup()

	if i.State() != runtime.ProcessOfflineState {
		return errors.New("Container already running")
	}

	sawError := false

	defer func() {
		if sawError {
			// If we don't set it to stopping first, you'll trigger crash detection which
			// we don't want to do at this point since it'll just immediately try to do the
			// exact same action that lead to it crashing in the first place...
			i.SetState(runtime.ProcessStoppingState)
			i.SetState(runtime.ProcessOfflineState)
		}
	}()

	if c, err := i.ContainerInspect(ctx); err != nil {
		// Do nothing if the container is not found, we just don't want to continue
		// to the next block of code here. This check was inlined here to guard against
		// a nil-pointer when checking c.State below.
		//
		// @see https://github.com/pterodactyl/panel/issues/2000
		if !client.IsErrNotFound(err) {
			return errors.Wrap(err, "runtime/docker: failed to inspect container")
		}
	} else {
		if c.State.Running {
			i.SetState(runtime.ProcessOfflineState)

			return i.Attach(ctx)
		}

		// TODO Log crap
	}

	i.SetState(runtime.ProcessStartingState)

	// Pretend we have seen an error for now
	sawError = true

	// Run the before start function and wait for it to finish. This will validate that the container
	// exists on the system, and rebuild the container if that is required for server booting to
	// occur.
	if err := i.Preflight(ctx); err != nil {
		return errors.Wrap(err, "runtime/docker: failed to run prelude")
	}

	// If we cannot start & attach to the container in 30 seconds something has gone
	// quite sideways, and we should stop trying to avoid a hanging situation.
	actx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	// You must attach to the instance _before_ you start the container. If you do this
	// in the opposite order you'll enter a deadlock condition where we're attached to
	// the instance successfully, but the container has already stopped and you'll get
	// the entire program into a very confusing state.
	//
	// By explicitly attaching to the instance before we start it, we can immediately
	// react to errors/output stopping/etc. when starting.
	if err := i.Attach(actx); err != nil {
		return errors.Wrap(err, "runtime/docker: failed to attach to container")
	}

	if err := i.client.ContainerStart(actx, i.Cfg.Uuid, container.StartOptions{}); err != nil {
		return errors.Wrap(err, "runtime/docker: failed to start container")
	}

	sawError = false
	return nil
}

// Stop stops the container that the server is running in. This will allow up to
// 30 seconds to pass before the container is forcefully terminated if we are
// trying to stop it without using a command sent into the instance.
//
// You most likely want to be using WaitForStop() rather than this function,
// since this will return as soon as the command is sent, rather than waiting
// for the process to be completed stopped.
func (i *Instance) Stop(ctx context.Context, skipLock bool, waitSeconds int) error {
	cleanup, err := i.AttemptPowerlock(ctx, skipLock, waitSeconds)
	if err != nil {
		return err
	}
	defer cleanup()

	i.Lock()
	s := i.Cfg.Stop
	i.Unlock()

	// A native "stop" as the Type field value will just skip over all of this
	// logic and end up only executing the container stop command (which may or
	// may not work as expected).
	if s == "" {
		if s == "" {
			log.Debug("no stop configuration set, using terminate command")
		}

		signal := os.Kill

		// Skipping some stuff
		return i.Terminate(ctx, signal, true, 0)
	}

	// If the process is already offline don't switch it back to stopping. Just leave it how
	// it is and continue through to the stop handling for the process
	if i.state.Load() != runtime.ProcessOfflineState {
		i.SetState(runtime.ProcessStoppingState)
	}

	// // Only attempt to send the stop command if we are attached
	// if i.IsAttached() && false {
	//   return i.SendCommand(s.)
	// }

	// Allow the stop action to run for however long it takes, similar to executing a command
	// and using a different logic pathway to wait for the container to stop successfully.
	//
	// Using a negative timeout here will allow the container to stop gracefully,
	// rather than forcefully terminating it.  Value is in seconds, but -1 is
	// treated as indefinitely.
	timeout := -1
	if err := i.client.ContainerStop(ctx, i.Cfg.Uuid, container.StopOptions{Timeout: &timeout}); err != nil {
		if client.IsErrNotFound(err) {
			i.SetStream(nil)
			i.SetState(runtime.ProcessOfflineState)
			return nil
		}

		return errors.Wrap(err, "runtime/docker: cannot stop container")
	}

	return nil
}

// WaitForStop attempts to gracefully stop a server using the defined stop
// command. If the server does not stop after seconds have passed, an error will
// be returned, or the instance will be terminated forcefully depending on the
// value of the second argument.
//
// Calls to Environment.Terminate() in this function use the context passed
// through since we don't want to prevent termination of the server instance
// just because the context.WithTimeout() has expired.
func (i *Instance) WaitForStop(ctx context.Context, duration time.Duration, terminate bool, skipLock bool, waitSeconds int) error {
	cleanup, err := i.AttemptPowerlock(ctx, skipLock, waitSeconds)
	if err != nil {
		return err
	}
	defer cleanup()

	tctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// If the parent context is cancled, abored the timed context for termination
	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-tctx.Done():
			break
		}
	}()

	doTermination := func(s string) error {
		log.Warnf("Terminating with %s", s)
		return i.Terminate(ctx, os.Kill, true, 0)
	}

	// We pass through the timed context for this stop action so that if one of the
	// internal docker calls fails to ever finish before we've exhausted the time limit
	// the resources get cleaned up, and the exection is stopped.
	if err := i.Stop(tctx, true, 0); err != nil {
		if terminate && errors.Is(err, context.DeadlineExceeded) {
			return doTermination("stop")
		}
		return err
	}

	// Block the return of this function until the container as been marked as no
	// longer running. If this wait does not end by the time seconds have passed,
	// attempt to terminate the container, or return an error.
	ok, errChan := i.client.ContainerWait(tctx, i.Cfg.Uuid, container.WaitConditionNotRunning)
	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			if terminate {
				return doTermination("parent-context")
			}
			return err
		}
	case err := <-errChan:
		if err == nil || client.IsErrNotFound(err) {
			return nil
		}
		if terminate {
			if !errors.Is(err, context.DeadlineExceeded) {
				log.Error(err, "error while waiting for container stop")
			}

			return doTermination("wait")
		}

		return errors.Wrapf(err, "runtime/docker: error waiting on conatainer to stop")
	case <-ok:
	}

	return nil
}

// Terminate forcefully terminates the container using the signal provided
func (i *Instance) Terminate(ctx context.Context, signal os.Signal, skipLock bool, waitSeconds int) error {
	log.Warnf("Terminating instance %s", i.Id)

	cleanup, err := i.AttemptPowerlock(ctx, skipLock, waitSeconds)
	if err != nil {
		return err
	}
	defer cleanup()

	c, err := i.ContainerInspect(ctx)
	if err != nil {
		if client.IsErrNotFound(err) {
			return nil
		}
		return errors.WithStack(err)
	}

	if !c.State.Running {
		if i.state.Load() != runtime.ProcessOfflineState {
			i.SetState(runtime.ProcessStoppingState)
			i.SetState(runtime.ProcessOfflineState)
		}

		return nil
	}

	// Set to stopping first to prevent crash detection
	i.SetState(runtime.ProcessStoppingState)
	sig := strings.TrimSuffix(strings.TrimPrefix(signal.String(), "signal "), "ed")
	if err := i.client.ContainerKill(ctx, i.Cfg.Uuid, sig); err != nil && !client.IsErrNotFound(err) {
		return errors.WithStack(err)
	}

	i.SetState(runtime.ProcessOfflineState)

	return nil
}

func (i *Instance) Restart(ctx context.Context) error {
	if err := i.WaitForStop(ctx, time.Second*10, true, false, 0); err != nil {
		return err
	}

	return i.Start(i.Ctx, false, 0)
}

func (i *Instance) AttemptPowerlock(ctx context.Context, skipLock bool, waitSeconds int) (func(), error) {
	if i.Installing.Load() {
		return nil, errors.New("server is installing")
	} else if i.Restoring.Load() {
		return nil, errors.New("server is restoring")
	} else if i.Transferring.Load() {
		return nil, errors.New("server is transferring")
	}

	cleanup := func() {
		log.Debug("Releasing powerlock...")
		i.Powerlock.Release()
	}

	if waitSeconds > 0 && !skipLock {
		lockCtx, cancel := context.WithTimeout(i.Ctx, time.Second*time.Duration(waitSeconds))
		defer cancel()

		// Attempt to acquire a lock on the power action lock for up to 30 seconds. If more
		// time than that passes an error will be propagated back up the chain and this
		// request will be aborted.
		if err := i.Powerlock.TryAcquire(lockCtx); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not acquire lock on power action after %d seconds", waitSeconds))
		}

		return cleanup, nil
	}

	if err := i.Powerlock.Acquire(); err != nil {
		log.Warn("failed to aquire powerlock...")
		if skipLock {
			log.Debug("skipping lock due to skiplock")
			return func() {}, nil
		}
		return nil, errors.Wrap(err, "failed to aquire powerlock")
	}

	return cleanup, nil
}
