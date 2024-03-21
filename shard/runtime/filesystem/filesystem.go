package filesystem

import "sync"

type Filesystem struct {
	sync.RWMutex
}
