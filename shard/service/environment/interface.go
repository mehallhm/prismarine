package environment

import (
	"context"
	"prismarine/shard/service/events"
)

const (
	StateChangeEvent         = "state change"
	ResourceEvent            = "resources"
	DockerImagePullStarted   = "docker image pull started"
	DockerImagePullStatus    = "docker image pull status"
	DockerImagePullCompleted = "docker image pull completed"
)

const (
	ProcessOfflineState  = "offline"
	ProcessStartingState = "starting"
	ProcessRunningState  = "running"
	ProcessStoppingState = "stopping"
)

type Instance interface {
	Type() string

	Config() *Configuration

	Events() *events.Bus

	// Exists determines whether the Instance exists in the environment
	Exists() (bool, error)

	IsRunning(ctx context.Context) (bool, error)

	Prelude(ctx context.Context) error

	Start(ctx context.Context) error

	Stop(ctx context.Context) error

	Create() error

	State() string

	SetState(string)
}
