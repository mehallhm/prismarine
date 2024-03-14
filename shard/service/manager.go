package service

import (
	"context"
	"github.com/rs/zerolog/log"
	"prismarine/shard/remote"
	"sync"
	"time"
)

type Manager struct {
	sync.RWMutex
	// client remote.Client
	servers []*Server
}

func NewManager(ctx context.Context) (*Manager, error) {
	m := &Manager{}
	if err := m.init(ctx); err != nil {
		return nil, err
	}
	return m, nil
}

// Len returns the number of stored servers
func (m *Manager) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.servers)
}

// Keys returns all the stored server UUIDs
func (m *Manager) Keys() []string {
	m.RLock()
	defer m.RUnlock()

	keys := make([]string, len(m.servers))
	for i, s := range m.servers {
		keys[i] = s.Id()
	}
	return keys
}

// All returns all the items in the collection
func (m *Manager) All() []*Server {
	m.RLock()
	defer m.RUnlock()
	return m.servers
}

// Add adds an item to the collection
func (m *Manager) Add(s *Server) {
	m.Lock()
	defer m.Unlock()
	m.servers = append(m.servers, s)
}

// TODO Get

// TODO Filter

// Find returns a single server matching the filter
func (m *Manager) Find(filter func(match *Server) bool) *Server {
	m.RLock()
	defer m.RUnlock()

	for _, v := range m.servers {
		if filter(v) {
			return v
		}
	}
	return nil
}

// Remove removes all items from the collection that match a filter function
func (m *Manager) Remove(filter func(match *Server) bool) {
	m.Lock()
	defer m.Unlock()

	r := make([]*Server, 0)
	for _, v := range m.servers {
		if !filter(v) {
			r = append(r, v)
		}
	}
}

func (m *Manager) InitServer(data remote.ServerData) (*Server, error) {
	s, err := New()
	if err != nil {
		return nil, err
	}

	return s, nil

}

func (m *Manager) init(ctx context.Context) error {
	log.Debug().Msg("Initializing Manager...")
	servers := make([]remote.ServerData, 1)
	servers[0] = remote.ServerData{
		Uuid: "89gt90",
	}

	start := time.Now()
	log.Debug().Msgf("Total game servers: %o", len(servers))

	// TODO Parallelize server initialization
	for _, data := range servers {
		data := data

		s, err := m.InitServer(data)
		if err != nil {
			log.Fatal().Msg("Failed to load server...")

			// TODO Fix w/ above comment.. should skip server if it fails to init
			continue
		}
		m.Add(s)
	}

	diff := time.Now().Sub(start)
	log.Debug().Msgf("Duration of startup: %s", diff)

	return nil
}
