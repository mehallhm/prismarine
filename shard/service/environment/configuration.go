package environment

import "sync"

type Configuration struct {
	sync.RWMutex
}
