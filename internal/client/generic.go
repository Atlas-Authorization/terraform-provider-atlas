package client

import "context"

// Call performs an arbitrary Backend API request with the instance secret key.
// body is JSON-encoded when non-nil; the response is decoded into out when out
// is non-nil. It backs the table-driven resources (generic_resource.go), whose
// request and response shapes come straight from the OpenAPI document, so they
// need no per-resource Go types.
func (c *Client) Call(ctx context.Context, method, path string, body, out interface{}) error {
	return c.do(ctx, method, path, body, out)
}
