package docker

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"prismarine/shard/runtime"
	"time"
)

func (i *Instance) Prelude(ctx context.Context) error {
	// Always destroy and re-create the server container
	if err := i.client.ContainerRemove(ctx, i.Id, container.RemoveOptions{}); err != nil {
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

func (i *Instance) Start(ctx context.Context) error {
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
	if err := i.Prelude(ctx); err != nil {
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

	if err := i.client.ContainerStart(actx, i.Id, container.StartOptions{}); err != nil {
		return errors.Wrap(err, "runtime/docker: failed to start container")
	}

	sawError = false
	return nil
}

func (i *Instance) Stop(ctx context.Context) error {
	panic("Nope")
}
