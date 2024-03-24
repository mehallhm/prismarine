package runtime

import "sync"

type Settings struct {
	Labels map[string]string
}

type Configuration struct {
	*sync.RWMutex

	Uuid string `json:"uuid"`

	Name        string `json:"name"`
	Description string `json:"description"`

	Invocation string `json:"invocation"`

	Container struct {
		Labels map[string]string `json:"labels"`
		Image  string            `json:"image,omitempty"`
	} `json:"container,omitempty"`

	Suspended bool `json:"suspended"`
}
