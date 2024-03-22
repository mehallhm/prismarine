package runtime

import "sync"

type Settings struct {
	Labels map[string]string
}

type Configuration struct {
	sync.RWMutex

	Uuid string `json:"uuid"`

	Name        string `json:"name"`
	Description string `json:"description"`

	Suspended bool `json:"suspended"`

	Invocation string `json:"invocation"`

	Labels map[string]string `json:"labels"`

	Container struct {
		Image string `json:"image,omitempty"`
	} `json:"container,omitempty"`
}
