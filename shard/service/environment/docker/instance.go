package docker

import (
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
