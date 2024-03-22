package monument

import (
	"bytes"
	"context"
	"net/http"
)

func (c *client) Get(ctx context.Context, path string, query q) (*Response, error) {
	return nil, nil
}

type Response struct {
	*http.Response
}

type q map[string]string

func (c *client) request(ctx context.Context, method, path string, body *bytes.Buffer, opts ...func(r *http.Request)) (*Response, error) {
	return nil, nil
}
