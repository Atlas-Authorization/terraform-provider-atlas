package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// ───────────────────────────── Design ─────────────────────────────

// Design mirrors the design read/write response for an instance
// (`object: "design"`) or an organization (`object: "organization_design"`):
// the stored overrides, the fully-resolved design every page renders, and
// contrast warnings.
type Design struct {
	Object         string          `json:"object"`
	OrganizationID *string         `json:"organization_id"`
	Design         json.RawMessage `json:"design"`
	Resolved       json.RawMessage `json:"resolved"`
	Warnings       json.RawMessage `json:"warnings"`
	Legacy         bool            `json:"legacy"`
}

type designBody struct {
	Design json.RawMessage `json:"design"`
}

// GetInstanceDesign reads the instance design. GET /v1/instance/design.
func (c *Client) GetInstanceDesign(ctx context.Context) (*Design, error) {
	var out Design
	if err := c.do(ctx, http.MethodGet, "/v1/instance/design", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PutInstanceDesign replaces the instance design ({ design }); an empty object
// removes it. PUT /v1/instance/design.
func (c *Client) PutInstanceDesign(ctx context.Context, design json.RawMessage) (*Design, error) {
	if len(design) == 0 {
		design = json.RawMessage(`{}`)
	}
	var out Design
	if err := c.do(ctx, http.MethodPut, "/v1/instance/design", designBody{Design: design}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOrganizationDesign reads an organization's design override.
// GET /v1/organizations/:id/design.
func (c *Client) GetOrganizationDesign(ctx context.Context, orgID string) (*Design, error) {
	var out Design
	if err := c.do(ctx, http.MethodGet, "/v1/organizations/"+url.PathEscape(orgID)+"/design", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PutOrganizationDesign replaces an organization's design override; an empty
// object removes it. PUT /v1/organizations/:id/design.
func (c *Client) PutOrganizationDesign(ctx context.Context, orgID string, design json.RawMessage) (*Design, error) {
	if len(design) == 0 {
		design = json.RawMessage(`{}`)
	}
	var out Design
	if err := c.do(ctx, http.MethodPut, "/v1/organizations/"+url.PathEscape(orgID)+"/design", designBody{Design: design}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ───────────────────────────── Cookie consent ─────────────────────────────

// CookieConsent mirrors /v1/instance/cookie_consent: the consent-management
// platform injected into the hosted pages. Enabled is false (and every other
// field absent) when none is set.
type CookieConsent struct {
	Enabled      bool     `json:"enabled"`
	ScriptURL    *string  `json:"script_url"`
	ScriptID     *string  `json:"script_id"`
	SiteID       *string  `json:"site_id"`
	BlockingMode *string  `json:"blocking_mode"`
	ExtraOrigins []string `json:"extra_origins"`
}

// CookieConsentWrite is the PUT body; it replaces the whole setting.
type CookieConsentWrite struct {
	ScriptURL    string   `json:"script_url"`
	ScriptID     *string  `json:"script_id"`
	SiteID       *string  `json:"site_id"`
	BlockingMode *string  `json:"blocking_mode"`
	ExtraOrigins []string `json:"extra_origins"`
}

// GetCookieConsent reads the setting. GET /v1/instance/cookie_consent.
func (c *Client) GetCookieConsent(ctx context.Context) (*CookieConsent, error) {
	var out CookieConsent
	if err := c.do(ctx, http.MethodGet, "/v1/instance/cookie_consent", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PutCookieConsent sets (replaces) the setting. PUT /v1/instance/cookie_consent.
func (c *Client) PutCookieConsent(ctx context.Context, body CookieConsentWrite) (*CookieConsent, error) {
	if body.ExtraOrigins == nil {
		body.ExtraOrigins = []string{}
	}
	var out CookieConsent
	if err := c.do(ctx, http.MethodPut, "/v1/instance/cookie_consent", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCookieConsent turns the CMP off. DELETE /v1/instance/cookie_consent.
func (c *Client) DeleteCookieConsent(ctx context.Context) error {
	return c.do(ctx, http.MethodDelete, "/v1/instance/cookie_consent", nil, nil)
}

// ───────────────────────────── Preflight ─────────────────────────────

// PreflightFinding is one setup problem the server can see.
type PreflightFinding struct {
	Severity string            `json:"severity"`
	Code     string            `json:"code"`
	Message  string            `json:"message"`
	Fix      string            `json:"fix"`
	Subject  map[string]string `json:"subject"`
}

// Preflight mirrors GET /v1/instance/preflight.
type Preflight struct {
	OK        bool               `json:"ok"`
	CheckedAt int64              `json:"checked_at"`
	Findings  []PreflightFinding `json:"findings"`
}

// GetPreflight runs the setup check. GET /v1/instance/preflight.
func (c *Client) GetPreflight(ctx context.Context) (*Preflight, error) {
	var out Preflight
	if err := c.do(ctx, http.MethodGet, "/v1/instance/preflight", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ───────────────────────────── Entitlement overrides ─────────────────────────────

// EntitlementOverride mirrors a sales-led entitlement override (no Stripe):
// a plan / features / limits pinned onto a user or org, optionally expiring.
type EntitlementOverride struct {
	ID          string                 `json:"id"`
	SubjectType string                 `json:"subject_type"`
	SubjectID   string                 `json:"subject_id"`
	Plan        *string                `json:"plan"`
	Features    []string               `json:"features"`
	Limits      map[string]interface{} `json:"limits"`
	ExpiresAt   *int64                 `json:"expires_at"`
	CreatedBy   *string                `json:"created_by"`
	CreatedAt   int64                  `json:"created_at"`
}

// EntitlementOverrideWrite is the POST body (grant-only; there is no edit).
type EntitlementOverrideWrite struct {
	SubjectType string              `json:"subject_type"`
	SubjectID   string              `json:"subject_id"`
	Plan        *string             `json:"plan,omitempty"`
	Features    []string            `json:"features,omitempty"`
	Limits      *map[string]float64 `json:"limits,omitempty"`
	ExpiresAt   *int64              `json:"expires_at,omitempty"`
}

// CreateEntitlementOverride grants an override. POST /v1/entitlement_overrides.
func (c *Client) CreateEntitlementOverride(ctx context.Context, body EntitlementOverrideWrite) (*EntitlementOverride, error) {
	var out EntitlementOverride
	if err := c.do(ctx, http.MethodPost, "/v1/entitlement_overrides", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEntitlementOverride finds one override by id. There is no single-GET
// route, so it lists (filtered by subject when known) and matches the id.
func (c *Client) GetEntitlementOverride(ctx context.Context, id, subjectType, subjectID string) (*EntitlementOverride, error) {
	q := url.Values{}
	if subjectType != "" {
		q.Set("subject_type", subjectType)
	}
	if subjectID != "" {
		q.Set("subject_id", subjectID)
	}
	path := "/v1/entitlement_overrides"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var out listEnvelope[EntitlementOverride]
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	for i := range out.Data {
		if out.Data[i].ID == id {
			return &out.Data[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Code: "NOT_FOUND", Message: "Unknown entitlement override."}}}
}

// DeleteEntitlementOverride revokes an override. DELETE /v1/entitlement_overrides/:id.
func (c *Client) DeleteEntitlementOverride(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/entitlement_overrides/"+url.PathEscape(id), nil, nil)
}

// ───────────────────────────── Messaging providers ─────────────────────────────

// MessagingProvider mirrors one channel of GET /v1/messaging_providers. Config
// carries the NON-secret settings only; SecretsSet names which secret keys are
// stored, never their values.
type MessagingProvider struct {
	Channel     string            `json:"channel"`
	Transport   string            `json:"transport"`
	FromAddress *string           `json:"from_address"`
	Status      string            `json:"status"`
	Config      map[string]string `json:"config"`
	SecretsSet  map[string]bool   `json:"secrets_set"`
	UpdatedAt   int64             `json:"updated_at"`
}

type messagingProviders struct {
	Email *MessagingProvider `json:"email"`
	SMS   *MessagingProvider `json:"sms"`
}

// MessagingProviderWrite is the PUT body for one channel. An omitted or empty
// secret keeps the stored one.
type MessagingProviderWrite struct {
	Transport   string            `json:"transport"`
	FromAddress *string           `json:"from_address,omitempty"`
	Config      map[string]string `json:"config"`
}

// GetMessagingProvider reads one channel ("email" or "sms"); a NOT_FOUND error
// when the channel has no provider configured. GET /v1/messaging_providers.
func (c *Client) GetMessagingProvider(ctx context.Context, channel string) (*MessagingProvider, error) {
	var out messagingProviders
	if err := c.do(ctx, http.MethodGet, "/v1/messaging_providers", nil, &out); err != nil {
		return nil, err
	}
	p := out.Email
	if channel == "sms" {
		p = out.SMS
	}
	if p == nil {
		return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Code: "NOT_FOUND", Message: "No " + channel + " provider is configured."}}}
	}
	return p, nil
}

// PutMessagingProvider saves one channel's provider.
// PUT /v1/messaging_providers/email | /v1/messaging_providers/sms.
func (c *Client) PutMessagingProvider(ctx context.Context, channel string, body MessagingProviderWrite) (*MessagingProvider, error) {
	if body.Config == nil {
		body.Config = map[string]string{}
	}
	var out MessagingProvider
	if err := c.do(ctx, http.MethodPut, "/v1/messaging_providers/"+url.PathEscape(channel), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMessagingProvider removes one channel's provider (back to the Atlas
// default sender). DELETE /v1/messaging_providers/:channel.
func (c *Client) DeleteMessagingProvider(ctx context.Context, channel string) error {
	return c.do(ctx, http.MethodDelete, "/v1/messaging_providers/"+url.PathEscape(channel), nil, nil)
}
