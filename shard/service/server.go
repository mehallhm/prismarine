package service

import (
	"context"
	"prismarine/shard/service/environment"
	"sync"
)

type Server struct {
	sync.RWMutex
	ctx       context.Context
	ctxCancel *context.CancelFunc

	cfg *Configuration

	instance environment.Instance

	//emitterLock sync.Mutex
	//powerLock *system.Locker

}

func New(config *Configuration) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := Server{
		ctx:       ctx,
		ctxCancel: &cancel,
		cfg:       config,
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
