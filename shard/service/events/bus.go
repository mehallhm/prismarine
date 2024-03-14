package events

import "strings"

type Event struct {
	topic string
	Data  interface{}
}

type Bus struct {
	*SinkPool
}

func NewBus() *Bus {
	return &Bus{
		NewSinkPool(),
	}
}

func (b *Bus) Publish(topic string, data interface{}) {
	// Some of our actions for the socket support passing a more specific namespace,
	// such as "backup completed:1234" to indicate which specific backup was completed.
	//
	// In these cases, we still need to send the event using the standard listener
	// name of "backup completed".
	if strings.Contains(topic, ":") {
		parts := strings.SplitN(topic, ":", 2)
		if len(parts) == 2 {
			topic = parts[0]
		}
	}
}
