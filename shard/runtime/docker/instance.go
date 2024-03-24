package docker

import (
	"context"
	"fmt"
	"prismarine/shard/runtime"
	"prismarine/shard/runtime/events"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Ensure that the Docker runtime is implementing all methods from
// the base runtime
var _ runtime.Instance = (*Instance)(nil)

type Metadata struct {
	Image string
	Stop  string
}

type Instance struct {
	runtime.RuntimeInstance

	Id string

	// The Instance configuration
	meta *Metadata

	// The Docker client being used for this runtime
	client *client.Client

	// Controls the hijacked response stream which only exists when
	// attached to the running docker instance
	stream *types.HijackedResponse

	state *runtime.AtomicString

	emitter *events.Bus
}

func New(id string, m *Metadata) (*Instance, error) {
	cli, err := Create()
	if err != nil {
		return nil, err
	}

	i := &Instance{
		Id:     id,
		client: cli,
		meta:   m,

		state: runtime.NewAtomicString(runtime.ProcessOfflineState),
	}

	return i, nil
}

// Type returns the type of Environment that that Instance is in
func (i *Instance) Type() string {
	return "docker"
}

// Exists determines if the container exists in this runtime. The ID passed
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

// Config returns the Configuration of the Instance
func (i *Instance) Config() *runtime.Configuration {
	i.RLock()
	defer i.RUnlock()
	return i.Cfg
}

// State returns the state of the Instance
func (i *Instance) State() string {
	return i.state.Load()
}

// Events returns an event bus for the instance
func (i *Instance) Events() *events.Bus {
	return i.emitter
}

// IsAttached determines if this process is currently attached to
// the container instance by checking if the stream is nil or not
func (i *Instance) IsAttached() bool {
	i.RLock()
	defer i.RUnlock()
	return i.stream != nil
}

// SetStream sets the current stream value from the Docker client. If a nil
// value is provided we assume that the stream is no longer operational and the
// instance is effectively offline
func (i *Instance) SetStream(s *types.HijackedResponse) {
	i.Lock()
	defer i.Unlock()
	i.stream = s
}

// SetState sets the state of the runtime. This emits an event that server's
// can hook into to take their own actions and track their own state based on
// the runtime.
func (i *Instance) SetState(state string) {
	if state != runtime.ProcessOfflineState &&
		state != runtime.ProcessStartingState &&
		state != runtime.ProcessRunningState &&
		state != runtime.ProcessStoppingState {
		panic(fmt.Errorf("invalid server state received: %s", state))
	}

	// Emit the event to any listeners that are currently registered.
	if i.State() != state {
		// If the state changed make sure we update the internal tracking to note that.
		i.state.Store(state)
		i.Events().Publish(runtime.StateChangeEvent, state)
	}
}
