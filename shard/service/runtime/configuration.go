package runtime

import "sync"

type Settings struct {
	Labels map[string]string
}

type Configuration struct {
	sync.RWMutex

	settings Settings
}
