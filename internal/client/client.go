// Package client is a minimal typed Go client for the Atlas Backend API (the
// instance-scoped /v1/* endpoints an `sk_` secret key can call).
//
// It deliberately depends on nothing but net/http and encoding/json: the
// Terraform provider only ever manages declarative config objects, so a small,
// dependency-free client keeps the build fast and the surface auditable. Every
// request is authenticated with `Authorization: Bearer <secret_key>`, sends and
// accepts JSON, and maps the §9.1 error envelope
//
//	{ "errors": [ { "code", "message", "param?", "meta?" } ] }
//
// onto an *APIError that callers (and the resources) can branch on.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultAPIURL is the Atlas Backend API origin, overridable per provider.
const DefaultAPIURL = "https://api.atlas.dev"

// Client is the shared HTTP core every resource method calls.
type Client struct {
	// APIURL is the base URL of the instance's Backend API, e.g.
	// https://api.atlas.dev. The /v1/... path is appended by each method;
	// trailing slashes are tolerated.
	APIURL string
	// SecretKey is the instance secret key (sk_...). Sent as a bearer token on
	// every request and never placed in a URL or logged.
	SecretKey string
	// HTTP is the underlying client; a caller may inject one for tests.
	HTTP *http.Client
	// UserAgent identifies the provider build to the API.
	UserAgent string
}

// New builds a Client with sane defaults.
func New(apiURL, secretKey string) *Client {
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}
	return &Client{
		APIURL:    strings.TrimRight(apiURL, "/"),
		SecretKey: secretKey,
		HTTP:      http.DefaultClient,
		UserAgent: "terraform-provider-atlas",
	}
}

// ErrorItem is one entry in the §9.1 error envelope. `code` is the stable,
// machine-readable part of the contract that callers branch on.
type ErrorItem struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Param   string                 `json:"param,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// APIError is returned for every non-2xx response. It keeps the HTTP status and
// the full envelope so a caller can inspect param/meta when needed.
type APIError struct {
	Status int
	Errors []ErrorItem
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		first := e.Errors[0]
		if first.Param != "" {
			return fmt.Sprintf("%s: %s (%s) [param=%s]", first.Code, first.Message, statusText(e.Status), first.Param)
		}
		return fmt.Sprintf("%s: %s (%s)", first.Code, first.Message, statusText(e.Status))
	}
	return fmt.Sprintf("atlas API request failed with status %d", e.Status)
}

// IsNotFound reports whether the error is a 404 / NOT_FOUND. Read paths use it
// to drop a resource from state when the object has been deleted out of band.
func (e *APIError) IsNotFound() bool {
	if e.Status == http.StatusNotFound {
		return true
	}
	return e.HasCode("NOT_FOUND")
}

// HasCode reports whether any error in the envelope carries the given code.
func (e *APIError) HasCode(code string) bool {
	for _, item := range e.Errors {
		if item.Code == code {
			return true
		}
	}
	return false
}

func statusText(status int) string {
	if t := http.StatusText(status); t != "" {
		return fmt.Sprintf("HTTP %d %s", status, t)
	}
	return fmt.Sprintf("HTTP %d", status)
}

// IsNotFound is a convenience for `errors.As`-free call sites: true when err is
// an *APIError representing a missing object.
func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.IsNotFound()
}

// do performs a single request. body is JSON-encoded when non-nil; the decoded
// response is written into out when out is non-nil and the body is non-empty.
func (c *Client) do(ctx context.Context, method, path string, body, out interface{}) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}

	url := c.APIURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.SecretKey)
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp.StatusCode, raw)
	}

	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decoding response body: %w", err)
	}
	return nil
}

// parseError turns a non-2xx body into an *APIError, tolerating a non-JSON body
// (e.g. an upstream proxy) by keeping it as the message.
func parseError(status int, raw []byte) error {
	apiErr := &APIError{Status: status}
	if len(raw) > 0 {
		var envelope struct {
			Errors []ErrorItem `json:"errors"`
		}
		if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Errors) > 0 {
			apiErr.Errors = envelope.Errors
			return apiErr
		}
		msg := string(raw)
		if len(msg) > 500 {
			msg = msg[:500]
		}
		apiErr.Errors = []ErrorItem{{Code: "UNKNOWN", Message: msg}}
		return apiErr
	}
	apiErr.Errors = []ErrorItem{{
		Code:    "UNKNOWN",
		Message: fmt.Sprintf("atlas API request failed with status %d", status),
	}}
	return apiErr
}
