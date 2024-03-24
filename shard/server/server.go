package server

import (
	"context"
	"os"
	"prismarine/shard/runtime"
	"prismarine/shard/runtime/events"
	"sync"

	"github.com/charmbracelet/log"
)

type Server struct {
	sync.RWMutex
	ctx       context.Context
	ctxCancel *context.CancelFunc

	instance runtime.Instance

	emitter *events.Bus

	// Tracks the installation process and prevents a server from
	// running two installer process at the same time
	installing   *runtime.AtomicBool
	restarting   *runtime.AtomicBool
	transferring *runtime.AtomicBool

	powerLock *runtime.Locker

	// emitterLock sync.Mutex

	log log.Logger
}

func New() (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := Server{
		ctx:       ctx,
		ctxCancel: &cancel,

		installing:   runtime.NewAtomicBool(false),
		restarting:   runtime.NewAtomicBool(false),
		transferring: runtime.NewAtomicBool(false),

		powerLock: runtime.NewLocker(),

		emitter: &events.Bus{
			SinkPool: events.NewSinkPool(),
		},

		log: *log.New(os.Stderr),
	}

	return &s, nil
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
