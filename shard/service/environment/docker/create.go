package docker

import (
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"sync"
)

var (
	_once   sync.Once
	_client *client.Client
)

func Create() (*client.Client, error) {
	var err error
	_once.Do(func() {
		_client, err = client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	})
	return _client, errors.Wrap(err, "could not create Docker client")
}
