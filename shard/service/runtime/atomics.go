package runtime

import "sync"

type AtomicString struct {
	sync.RWMutex
	v string
}

func (as *AtomicString) Store(v string) {
	as.Lock()
	defer as.Unlock()
	as.v = v
}

func (as *AtomicString) Load() string {
	as.Lock()
	defer as.Unlock()
	return as.v
}
