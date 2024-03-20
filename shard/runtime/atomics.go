package runtime

import (
	"encoding/json"
	"sync"
)

type AtomicString struct {
	sync.RWMutex
	v string
}

func NewAtomicString(v string) *AtomicString {
	return &AtomicString{v: v}
}

func (as *AtomicString) Store(v string) {
	as.Lock()
	defer as.Unlock()
	as.v = v
}

func (as *AtomicString) Load() string {
	as.RLock()
	defer as.RUnlock()
	return as.v
}

type AtomicBool struct {
	sync.RWMutex
	v bool
}

func NewAtomicBool(v bool) *AtomicBool {
	return &AtomicBool{v: v}
}

func (ab *AtomicBool) Store(v bool) {
	ab.Lock()
	defer ab.Unlock()
	ab.v = v
}

// SwapIf stores the value "v" if the current value stored in the AtomicBool is
// the opposite boolean value. If successfully swapped, the response is "true",
// otherwise "false" is returned.
func (ab *AtomicBool) SwapIf(v bool) bool {
	ab.Lock()
	defer ab.Unlock()
	if ab.v != v {
		ab.v = v
		return true
	}
	return false
}

func (ab *AtomicBool) Load() bool {
	ab.RLock()
	defer ab.RUnlock()
	return ab.v
}

func (ab *AtomicBool) UnmarshalJson(b []byte) error {
	ab.Lock()
	defer ab.Unlock()
	return json.Unmarshal(b, &ab.v)
}

func (ab *AtomicBool) MarshalJson() ([]byte, error) {
	return json.Marshal(ab.Load())
}
