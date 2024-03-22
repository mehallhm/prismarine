package monument

import (
	"context"
	"net/http"
	"prismarine/shard/remote"
)

var _ remote.Client = (*client)(nil)

type client struct {
	httpClient  *http.Client
	baseUrl     string
	tokenId     string
	maxAttempts int
}

func (c *client) GetServers(ctx context.Context, limit int) ([]remote.RawServerData, error) {
	return nil, nil
}
