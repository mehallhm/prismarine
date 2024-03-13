package service

import (
	"context"
	"sync"
)

type Server struct {
	sync.RWMutex
	ctx       context.Context
	ctxCancel *context.CancelFunc

	cfg Configuration

	//emitterLock sync.Mutex
	//powerLock *system.Locker

}

func New() (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := Server{
		ctx:       ctx,
		ctxCancel: &cancel,
	}

	return &s, nil
}
