package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// ───────────────────────── Policy acceptance ─────────────────────────

// Policy mirrors GET /v1/policies/{id}: a policy (terms of service, privacy
// policy, any key) users accept. Versions are content and are not part of this
// type; publish them with the Backend API, the dashboard, or a policy source.
type Policy struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Title     string `json:"title"`
	Display   string `json:"display"`
	Required  bool   `json:"required"`
	AppliesTo string `json:"applies_to"`
	Reaccept  string `json:"reaccept"`
	GraceDays *int64 `json:"grace_days"`
	Source    string `json:"source"`
	Position  int64  `json:"position"`
	Status    string `json:"status"`
}

// PolicyWrite is the body of POST /v1/policies and PATCH /v1/policies/{id}.
// The provider always sends every field it manages, GraceDays nil as an
// explicit null (clears the grace window), so a PATCH converges the policy on
// the configuration. Key is sent on create only; Status is sent on update only.
type PolicyWrite struct {
	Key       string
	Title     string
	Display   string
	Required  bool
	AppliesTo string
	Reaccept  string
	GraceDays *int64
	Status    string
}

// MarshalJSON renders the body with the nullable grace_days.
func (w PolicyWrite) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"title":      w.Title,
		"display":    w.Display,
		"required":   w.Required,
		"applies_to": w.AppliesTo,
		"reaccept":   w.Reaccept,
		"grace_days": w.GraceDays,
	}
	if w.Key != "" {
		m["key"] = w.Key
	}
	if w.Status != "" {
		m["status"] = w.Status
	}
	return json.Marshal(m)
}

func policyPath(idOrKey string) string { return "/v1/policies/" + url.PathEscape(idOrKey) }

// CreatePolicy creates a policy. POST /v1/policies.
func (c *Client) CreatePolicy(ctx context.Context, body PolicyWrite) (*Policy, error) {
	var out Policy
	if err := c.do(ctx, http.MethodPost, "/v1/policies", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPolicy reads a policy by id or key. GET /v1/policies/{id}.
func (c *Client) GetPolicy(ctx context.Context, idOrKey string) (*Policy, error) {
	var out Policy
	if err := c.do(ctx, http.MethodGet, policyPath(idOrKey), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdatePolicy patches a policy. PATCH /v1/policies/{id}.
func (c *Client) UpdatePolicy(ctx context.Context, id string, body PolicyWrite) (*Policy, error) {
	var out Policy
	if err := c.do(ctx, http.MethodPatch, policyPath(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeletePolicy deletes a never-published policy; one with a published version
// is retired instead (its acceptances stay). DELETE /v1/policies/{id}.
func (c *Client) DeletePolicy(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, policyPath(id), nil, nil)
}

// IsConflict reports whether err is a 409 / conflict (a duplicate key).
func IsConflict(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && (apiErr.Status == http.StatusConflict || apiErr.HasCode("CONFLICT"))
}

// PolicySettings mirrors GET /v1/instance/policy_settings.
type PolicySettings struct {
	SessionClaim bool `json:"session_claim"`
}

const policySettingsPath = "/v1/instance/policy_settings"

// GetPolicySettings reads the settings. GET /v1/instance/policy_settings.
func (c *Client) GetPolicySettings(ctx context.Context) (*PolicySettings, error) {
	var out PolicySettings
	if err := c.do(ctx, http.MethodGet, policySettingsPath, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetPolicySettings sets whether the pol claim rides on session tokens.
// PATCH /v1/instance/policy_settings.
func (c *Client) SetPolicySettings(ctx context.Context, sessionClaim bool) (*PolicySettings, error) {
	var out PolicySettings
	if err := c.do(ctx, http.MethodPatch, policySettingsPath, map[string]interface{}{"session_claim": sessionClaim}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PolicySource mirrors GET /v1/instance/policy_source: the external system that
// serves the instance's policy documents. The signing secret is write-only:
// only SigningSecretSet is ever returned.
type PolicySource struct {
	Enabled             bool    `json:"enabled"`
	EndpointURL         *string `json:"endpoint_url"`
	SigningSecretSet    bool    `json:"signing_secret_set"`
	Sync                string  `json:"sync"`
	PollIntervalMinutes int64   `json:"poll_interval_minutes"`
	WebhookURL          *string `json:"webhook_url"`
	ProtocolVersion     string  `json:"protocol_version"`
	LastSyncedAt        *int64  `json:"last_synced_at"`
	LastSyncOK          *bool   `json:"last_sync_ok"`
	LastSyncError       *string `json:"last_sync_error"`
}

// PolicySourceWrite is the PUT body. The API merges it onto the stored source,
// so the provider sends every field it manages. The secret is tri-state:
// SendSecret false omits it (the stored one is kept); SendSecret true sends
// Secret, where nil clears the stored secret. PollIntervalMinutes nil is omitted.
type PolicySourceWrite struct {
	Enabled             bool
	EndpointURL         string
	Sync                string
	PollIntervalMinutes *int64
	SendSecret          bool
	Secret              *string
}

// MarshalJSON renders the body with the optional secret.
func (w PolicySourceWrite) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"enabled":      w.Enabled,
		"endpoint_url": w.EndpointURL,
		"sync":         w.Sync,
	}
	if w.PollIntervalMinutes != nil {
		m["poll_interval_minutes"] = *w.PollIntervalMinutes
	}
	if w.SendSecret {
		m["signing_secret"] = w.Secret
	}
	return json.Marshal(m)
}

const policySourcePath = "/v1/instance/policy_source"

// GetPolicySource reads the source. GET /v1/instance/policy_source.
func (c *Client) GetPolicySource(ctx context.Context) (*PolicySource, error) {
	var out PolicySource
	if err := c.do(ctx, http.MethodGet, policySourcePath, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PutPolicySource creates or updates the source. PUT /v1/instance/policy_source.
func (c *Client) PutPolicySource(ctx context.Context, body PolicySourceWrite) (*PolicySource, error) {
	var out PolicySource
	if err := c.do(ctx, http.MethodPut, policySourcePath, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeletePolicySource disconnects the source; its policies and history stay.
// DELETE /v1/instance/policy_source.
func (c *Client) DeletePolicySource(ctx context.Context) error {
	return c.do(ctx, http.MethodDelete, policySourcePath, nil, nil)
}
