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

type Instance struct {
	runtime.RuntimeInstance

	// The Docker client being used for this runtime
	client *client.Client

	// Controls the hijacked response stream which only exists when
	// attached to the running docker instance
	stream *types.HijackedResponse

	state *runtime.AtomicString

	emitter *events.Bus
}

func New(config *runtime.Configuration) (*Instance, error) {
	cli, err := Create()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	i := &Instance{
		RuntimeInstance: runtime.RuntimeInstance{
			Ctx:       ctx,
			CtxCancel: &cancel,

			Cfg: config,

			Transferring: runtime.NewAtomicBool(false),
			Restoring:    runtime.NewAtomicBool(false),
			Installing:   runtime.NewAtomicBool(false),

			Powerlock: runtime.NewLocker(),
		},
		client: cli,

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

func (i *Instance) ContextCancel() {
	if i.CtxCancel != nil {
		(*i.CtxCancel)()
	}
}

func (i *Instance) Context() context.Context {
	return i.Ctx
}

func (i *Instance) Destroy() error {
	panic("Not implimented")
}

func (i *Instance) ExitState() (uint32, bool, error) {
	panic("not implimented")
}

func (i *Instance) ReadLog(depth int) ([]string, error) {
	panic("Not implimented")
}

func (i *Instance) SendCommand(cmd string) error {
	panic("not implimented")
}

func (i *Instance) SetLogCallback(func([]byte)) {
	panic("Not implimented")
}

func (i *Instance) Uptime(ctx context.Context) (int64, error) {
	panic("not implimented")
}
