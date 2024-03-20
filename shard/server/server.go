package server

import (
	"context"
	"prismarine/shard/runtime"
	"prismarine/shard/runtime/events"
	"sync"
)

type Server struct {
	sync.RWMutex
	ctx       context.Context
	ctxCancel *context.CancelFunc

	cfg *Configuration

	instance runtime.Instance

	emitter *events.Bus

	// Tracks the installation process and prevents a server from
	// running two installer process at the same time
	installing   *runtime.AtomicBool
	restarting   *runtime.AtomicBool
	transferring *runtime.AtomicBool

	//emitterLock sync.Mutex
	powerLock *Locker
}

func New(config *Configuration) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := Server{
		ctx:       ctx,
		ctxCancel: &cancel,
		cfg:       config,

		installing:   runtime.NewAtomicBool(false),
		restarting:   runtime.NewAtomicBool(false),
		transferring: runtime.NewAtomicBool(false),

		powerLock: NewLocker(),

		emitter: &events.Bus{
			SinkPool: events.NewSinkPool(),
		},
	}

	return &s, nil
}

// CreateInstance creates the Environment Instance for the Server
func (s *Server) CreateInstance() error {
	return s.instance.Create()
}

// Id returns the UUID for the server
func (s *Server) Id() string {
	return s.Config().GetUuid()
}

// CtxCancel cancels the context assigned to this server instance, canceling all
// background tasks
func (s *Server) CtxCancel() {
	if s.ctxCancel != nil {
		(*s.ctxCancel)()
	}
}

// Context returns a context instance for the server
func (s *Server) Context() context.Context {
	return s.ctx
}
