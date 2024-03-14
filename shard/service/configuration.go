package service

import "sync"

type Configuration struct {
	sync.RWMutex

	Uuid        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Container struct {
		Image string `json:"image"`
	} `json:"container"`
}

func (s *Server) Config() *Configuration {
	s.cfg.RLock()
	defer s.cfg.RUnlock()
	return &s.cfg
}

func (c *Configuration) GetUuid() string {
	c.RLock()
	defer c.RUnlock()
	return c.Uuid
}
