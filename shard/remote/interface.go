package remote

import "context"

type Client interface {
	// GetServers returns all the servers that are present in the Monument
	GetServers(context context.Context, perPage int) ([]RawServerData, error)
}
