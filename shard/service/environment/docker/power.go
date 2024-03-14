package docker

import (
	"context"
)

func (i *Instance) Prelude(ctx context.Context) error {
	// TODO
	panic("Implementation needed!")
}

func (i *Instance) Start(ctx context.Context) error {
	//sawError := false
	//
	//defer func() {
	//	if sawError {
	//		// TODO
	//	}
	//}()
	//
	//if c, err := e.ContainerInspect(ctx); err != nil {
	//	// Do nothing if the container is not found, we just don't want to continue
	//	// to the next block of code here. This check was inlined here to guard against
	//	// a nil-pointer when checking c.State below.
	//	//
	//	// @see https://github.com/pterodactyl/panel/issues/2000
	//	if !client.IsErrNotFound(err) {
	//		return errors.WrapIf(err, "environment/docker: failed to inspect container")
	//	}
	//} else {
	//	// If the server is running update our internal state and continue on with the attach.
	//	if c.State.Running {
	//		i.SetState(environment.ProcessRunningState)
	//
	//		return i.Attach(ctx)
	//	}
	//}
	//
	//sawError = true
	//
	panic("Not implimented yet")
}

func (i *Instance) Create() error {
	panic("Nope")
}

func (i *Instance) Stop(ctx context.Context) error {
	panic("Nope")
}
