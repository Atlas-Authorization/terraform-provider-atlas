package client

import (
	"context"
	"net/http"
	"net/url"
)

// ───────────────────────────── Trusted-JWT connections ─────────────────────────────

// JWTConnection mirrors /v1/jwt_connections/:id: a connection that signs a
// person in from a short-lived JWT signed by the customer's own system.
type JWTConnection struct {
	ID             string  `json:"id"`
	OrganizationID *string `json:"organization_id"`
	Status         string  `json:"status"`
	JWKSURI        *string `json:"jwks_uri"`
	Issuer         *string `json:"issuer"`
	Audience       *string `json:"audience"`
	EmailClaim     string  `json:"email_claim"`
	MaxAgeSeconds  int64   `json:"max_age_seconds"`
	DefaultRoleID  *string `json:"default_role_id"`
	LoginURL       string  `json:"login_url"`
	CreatedAt      int64   `json:"created_at"`
	UpdatedAt      int64   `json:"updated_at"`
}

// JWTConnectionWrite is the create / PATCH body: the keys to send. A key whose
// value is nil is sent as JSON null (organization_id and default_role_id clear).
type JWTConnectionWrite map[string]interface{}

func (c *Client) CreateJWTConnection(ctx context.Context, body JWTConnectionWrite) (*JWTConnection, error) {
	var out JWTConnection
	if err := c.do(ctx, http.MethodPost, "/v1/jwt_connections", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetJWTConnection(ctx context.Context, id string) (*JWTConnection, error) {
	var out JWTConnection
	if err := c.do(ctx, http.MethodGet, "/v1/jwt_connections/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateJWTConnection(ctx context.Context, id string, body JWTConnectionWrite) (*JWTConnection, error) {
	var out JWTConnection
	if err := c.do(ctx, http.MethodPatch, "/v1/jwt_connections/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteJWTConnection(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/jwt_connections/"+url.PathEscape(id), nil, nil)
}
