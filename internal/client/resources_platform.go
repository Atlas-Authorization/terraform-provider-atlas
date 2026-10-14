package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// ───────────────────── Sign in with Atlas apps (platform key) ─────────────────────

// AtlasIDApp mirrors the /v1/platform/atlas_id_apps projection: an OpenID
// Connect client on the shared Sign in with Atlas issuer, owned by the
// platform's account. The client secret is returned only on create and on
// rotate_secret.
type AtlasIDApp struct {
	ID                                string          `json:"id"`
	ClientID                          string          `json:"client_id"`
	Name                              string          `json:"name"`
	LogoURL                           *string         `json:"logo_url"`
	HomepageURL                       *string         `json:"homepage_url"`
	Branding                          json.RawMessage `json:"branding"`
	RedirectURIs                      []string        `json:"redirect_uris"`
	PostLogoutRedirectURIs            []string        `json:"post_logout_redirect_uris"`
	BackchannelLogoutURI              *string         `json:"backchannel_logout_uri"`
	BackchannelLogoutSessionRequired  bool            `json:"backchannel_logout_session_required"`
	FrontchannelLogoutURI             *string         `json:"frontchannel_logout_uri"`
	FrontchannelLogoutSessionRequired bool            `json:"frontchannel_logout_session_required"`
	AllowedScopes                     []string        `json:"allowed_scopes"`
	Type                              string          `json:"type"`
	SecretPrefix                      *string         `json:"secret_prefix"`
	ManagedByInstanceID               *string         `json:"managed_by_instance_id"`
	UserCount                         *int64          `json:"user_count"`
	CreatedAt                         int64           `json:"created_at"`
	UpdatedAt                         int64           `json:"updated_at"`
	// ClientSecret is present only on create and rotate_secret responses.
	ClientSecret *string `json:"client_secret"`
}

// AtlasIDAppWrite is the create/PATCH body. It is a map so that an explicit
// JSON null (which CLEARS a field server-side: logout URIs, logo_url,
// homepage_url, branding keys) can be sent distinctly from an omitted field.
type AtlasIDAppWrite map[string]interface{}

// CreateAtlasIDApp registers an app. POST /v1/platform/atlas_id_apps.
func (c *Client) CreateAtlasIDApp(ctx context.Context, body AtlasIDAppWrite) (*AtlasIDApp, error) {
	var out AtlasIDApp
	if err := c.doPlatform(ctx, http.MethodPost, "/v1/platform/atlas_id_apps", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAtlasIDApp reads one app. GET /v1/platform/atlas_id_apps/:id (404 when
// another account's or deleted).
func (c *Client) GetAtlasIDApp(ctx context.Context, id string) (*AtlasIDApp, error) {
	var out AtlasIDApp
	if err := c.doPlatform(ctx, http.MethodGet, "/v1/platform/atlas_id_apps/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAtlasIDApp patches an app. PATCH /v1/platform/atlas_id_apps/:id.
// branding merges per field; a null clears.
func (c *Client) UpdateAtlasIDApp(ctx context.Context, id string, body AtlasIDAppWrite) (*AtlasIDApp, error) {
	var out AtlasIDApp
	if err := c.doPlatform(ctx, http.MethodPatch, "/v1/platform/atlas_id_apps/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RotateAtlasIDAppSecret mints a new client secret (returned once; the old
// one keeps working for 24 hours). POST /v1/platform/atlas_id_apps/:id/rotate_secret.
func (c *Client) RotateAtlasIDAppSecret(ctx context.Context, id string) (*AtlasIDApp, error) {
	var out AtlasIDApp
	if err := c.doPlatform(ctx, http.MethodPost, "/v1/platform/atlas_id_apps/"+url.PathEscape(id)+"/rotate_secret", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAtlasIDApp deletes an app. DELETE /v1/platform/atlas_id_apps/:id.
func (c *Client) DeleteAtlasIDApp(ctx context.Context, id string) error {
	return c.doPlatform(ctx, http.MethodDelete, "/v1/platform/atlas_id_apps/"+url.PathEscape(id), nil, nil)
}
