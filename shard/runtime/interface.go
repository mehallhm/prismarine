package runtime

import (
	"context"
	"os"
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
	// Type returns the type of runtime
	Type() string

	// Config returns the runtime configuration
	Config() *Configuration

	// Events returns an event emitter instance that can be hooked into for
	// different events that are fired by the runtime
	Events() *events.Bus

	// Exists determines in the server instance exists
	Exists() (bool, error)

	// IsRunning determines if the runtime is currently active and runnining
	// a server process for this specific instance
	IsRunning(ctx context.Context) (bool, error)

	// Preflight runs before an instance is started
	Preflight(ctx context.Context) error

	// Start starts a server insance. If the server is not in a state where it
	// can be started an error should be returned
	Start(ctx context.Context) error

	// Stops stops a server instance. If the server is already stopped an
	// error will not be returned
	Stop(ctx context.Context) error

	// WaitForStop waits for a server Instance to stop gracefully. If it
	// does not stop in the given duration, it will either error or terminate
	WaitForStop(ctx context.Context, duration time.Duration, terminate bool) error

	// Terminate stops a running server instance using the provided signal. An
	// error is not thrown if it is already stopped
	Terminate(ctx context.Context, signal os.Signal) error

	// Destroy destroys the instance
	Destroy() error

	// ExitState returns the exit state of the process
	ExitState() (uint32, bool, error)

	// Create creates the necessary instance for running the server process
	Create() error

	// Attach attaches to the server console environment and allows piping
	// the output to an internal tool to monitor output. Also allows sending data
	// in to stdin
	Attach(ctx context.Context) error

	// SendCommand sends a provided command to the instance
	SendCommand(string) error

	// ReadLog reads the log file for the process from the end backwards until
	// the provided number of lines is met
	ReadLog(int) ([]string, error)

	// State retuns the current state of the instance
	State() string

	// SetState sets the current state of the instance
	SetState(string)

	// Uptime returns the current instance uptime in milliseconds
	Uptime(ctx context.Context) (int64, error)

	// SetLogCallback sets the callback that the container's log
	// output will be passed to
	SetLogCallback(func([]byte))
}
