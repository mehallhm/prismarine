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

func (c *Configuration) GetUuid() string {
	c.RLock()
	defer c.RUnlock()
	return c.Uuid
}
