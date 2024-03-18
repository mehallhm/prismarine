package events

import (
	"sync"
	"time"
)

type SinkName string

const (
	// LogSink handles console output for game servers, including messages being
	// sent via Shard to the console instance
	LogSink SinkName = "log"
	// InstallSink handles installation output for a server
	InstallSink SinkName = "install"
)

// SinkPool represents a pool of sinks
type SinkPool struct {
	sync.RWMutex
	sinks []chan []byte
}

// NewSinkPool returns a new empty SinkPool, which generally lives with an
// instance for its full lifetime
func NewSinkPool() *SinkPool {
	return &SinkPool{}
}

// On adds a channel to the sink pool instance
func (p *SinkPool) On(c chan []byte) {
	p.Lock()
	defer p.Unlock()
	p.sinks = append(p.sinks, c)
}

// Off removes a given channel from the pool. If no matching channel is found
// this function is a no-op
func (p *SinkPool) Off(c chan []byte) {
	p.Lock()
	defer p.Unlock()

	sinks := p.sinks
	for i, sink := range sinks {
		if c != sink {
			continue
		}

		// Need to maintain order, so shift everything to the left
		copy(sinks[i:], sinks[i+1:])
		sinks[len(sinks)-1] = nil
		sinks = sinks[:len(sinks)-1]
		p.sinks = sinks

		// Avoid a panic if the sink channel is nil
		if c != nil {
			close(c)
		}

		return
	}
}

// Destroy closes and removes all channels
func (p *SinkPool) Destroy() {
	p.Lock()
	defer p.Unlock()

	for _, c := range p.sinks {
		if c != nil {
			close(c)
		}
	}

	p.sinks = nil
}

// Push sends a given message to each of the channels registered in the pool.
// This will use a Ring Buffer channel in order to avoid blocking the channel
// sends, and attempt to push though the most recent messages in the queue in
// favor of the oldest messages.
//
// If the channel becomes full and isn't being drained fast enough, this
// function will remove the oldest message in the channel, and then push the
// message that it got onto the end, effectively making the channel a rolling
// buffer.
//
// There is a potential for data to be lost when passing it through this
// function, but only in instances where the channel buffer is full and the
// channel is not drained fast enough, in which case dropping messages is most
// likely the best option anyways. This uses waitgroups to allow every channel
// to attempt its send concurrently thus making the total blocking time of this
// function "O(1)" instead of "O(n)".
func (p *SinkPool) Push(data []byte) {
	p.Lock()
	defer p.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(p.sinks))
	for _, c := range p.sinks {
		go func(c chan []byte) {
			defer wg.Done()
			select {
			case c <- data:
			case <-time.After(time.Millisecond * 10):
				if len(c) == 0 {
					break
				}
				<-c
				c <- data
			}
		}(c)
	}
	wg.Wait()
}
