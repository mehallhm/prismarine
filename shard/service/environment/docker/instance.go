package docker

import (
	"context"
	"github.com/docker/docker/client"
	"prismarine/shard/service/environment"
	"sync"
)

// Ensure that the Docker environment is implementing all methods from
// the base environment
var _ environment.Instance = (*Instance)(nil)

type Instance struct {
	sync.RWMutex

	Id string

	// The Instance configuration
	Configuration *environment.Configuration

	// The Docker client being used for this environment
	client *client.Client
}

func New(id string) (*Instance, error) {
	cli, err := Create()
	if err != nil {
		return nil, err
	}

	i := &Instance{
		Id:     id,
		client: cli,
	}

	return i, nil
}

// Type returns the type of Environment that that Instance is in
func (i *Instance) Type() string {
	return "docker"
}

// Exists determines if the container exists in this environment. The ID passed
// through should be the server UUID.
func (i *Instance) Exists() (bool, error) {
	_, err := i.ContainerInspect(context.Background())
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (i *Instance) IsRunning(ctx context.Context) (bool, error) {
	c, err := i.ContainerInspect(ctx)
	if err != nil {
		return false, err
	}
	return c.State.Running, nil
}

func (i *Instance) Config() *environment.Configuration {
	i.RLock()
	defer i.RUnlock()
	return i.Configuration
}
