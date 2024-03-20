package runtime

import (
	"context"
	"prismarine/shard/runtime/events"
	"time"
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

	// Exists determines whether the Instance exists in the runtime
	Exists() (bool, error)

	IsRunning(ctx context.Context) (bool, error)

	Preflight(ctx context.Context) error

	Start(ctx context.Context) error

	Stop(ctx context.Context) error

	WaitForStop(ctx context.Context, duration time.Duration, terminate bool) error

	Create() error

	State() string

	SetState(string)
}
