package service

import "sync"

type Manager struct {
	mu      sync.RWMutex
	servers []*Server
}

func NewManager() (*Manager, error) {
	return nil, nil
}
