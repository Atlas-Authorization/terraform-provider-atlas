package client

import (
	"context"
	"encoding/json"
	"net/http"
)

// ───────────────────────── Better Auth session bridge ─────────────────────────

const betterAuthBridgePath = "/v1/migrations/better_auth/bridge"

// BetterAuthBridge mirrors GET /v1/migrations/better_auth/bridge: during a
// migration window, a returning user's Better Auth session is exchanged once
// for an Atlas session. CreatedAt is nil when the bridge was never configured
// (the API then answers with its defaults). The old app's secret is
// write-only: only LiveFallback.HasSecret is ever returned.
type BetterAuthBridge struct {
	Enabled      bool                         `json:"enabled"`
	EndsAt       *int64                       `json:"ends_at"`
	CookiePrefix string                       `json:"cookie_prefix"`
	LiveFallback BetterAuthBridgeLiveFallback `json:"live_fallback"`
	ExchangePath string                       `json:"exchange_path"`
	CreatedAt    *int64                       `json:"created_at"`
	UpdatedAt    *int64                       `json:"updated_at"`
}

// BetterAuthBridgeLiveFallback is the server-to-server check of a session
// missing from the import against the old app.
type BetterAuthBridgeLiveFallback struct {
	Enabled   bool    `json:"enabled"`
	BaseURL   *string `json:"base_url"`
	HasSecret bool    `json:"has_secret"`
}

// BetterAuthBridgeWrite is the PUT body. The API merges it onto the stored
// settings, so the provider always sends every field it manages: EndsAt and
// BaseURL nil are sent as null (no window end / no base URL).
type BetterAuthBridgeWrite struct {
	Enabled      bool
	EndsAt       *string
	CookiePrefix string
	LiveFallback BetterAuthBridgeLiveFallbackWrite
}

// BetterAuthBridgeLiveFallbackWrite is the live_fallback of a PUT body. The
// secret is tri-state: SendSecret false omits it (the stored one is kept);
// SendSecret true sends Secret, where nil clears the stored secret.
type BetterAuthBridgeLiveFallbackWrite struct {
	Enabled    bool
	BaseURL    *string
	SendSecret bool
	Secret     *string
}

// MarshalJSON renders the body with explicit nulls and the optional secret.
func (w BetterAuthBridgeWrite) MarshalJSON() ([]byte, error) {
	live := map[string]interface{}{
		"enabled":  w.LiveFallback.Enabled,
		"base_url": w.LiveFallback.BaseURL,
	}
	if w.LiveFallback.SendSecret {
		live["secret"] = w.LiveFallback.Secret
	}
	return json.Marshal(map[string]interface{}{
		"enabled":       w.Enabled,
		"ends_at":       w.EndsAt,
		"cookie_prefix": w.CookiePrefix,
		"live_fallback": live,
	})
}

// GetBetterAuthBridge reads the settings. GET /v1/migrations/better_auth/bridge.
func (c *Client) GetBetterAuthBridge(ctx context.Context) (*BetterAuthBridge, error) {
	var out BetterAuthBridge
	if err := c.do(ctx, http.MethodGet, betterAuthBridgePath, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PutBetterAuthBridge creates or updates the settings. PUT /v1/migrations/better_auth/bridge.
func (c *Client) PutBetterAuthBridge(ctx context.Context, body BetterAuthBridgeWrite) (*BetterAuthBridge, error) {
	var out BetterAuthBridge
	if err := c.do(ctx, http.MethodPut, betterAuthBridgePath, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBetterAuthBridge removes the settings: the bridge is off and the stored
// secret is deleted. DELETE /v1/migrations/better_auth/bridge.
func (c *Client) DeleteBetterAuthBridge(ctx context.Context) error {
	return c.do(ctx, http.MethodDelete, betterAuthBridgePath, nil, nil)
}
