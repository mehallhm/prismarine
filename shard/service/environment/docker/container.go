package docker

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/pkg/errors"
	"prismarine/shard/service/environment"
)

func (i *Instance) Attach(ctx context.Context) error {
	if i.IsAttached() {
		return nil
	}

	opts := container.AttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	}

	// Set the stream again with the container.
	if st, err := i.client.ContainerAttach(ctx, i.Id, opts); err != nil {
		return errors.Wrap(err, "environment/docker: error while attaching to container")
	} else {
		i.SetStream(&st)
	}

	go func() {
		//pollCtx, cancel := context.WithCancel(context.Background())
		//defer cancel()
		defer i.stream.Close()
		defer func() {
			i.SetState(environment.ProcessOfflineState)
			i.SetStream(nil)
		}()

		//go func() {
		//	if err := i.
		//}()

	}()

	return nil
}
