package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// ─────────────────────────── Instance configuration ───────────────────────────

// Instance mirrors the §9.3 GET /v1/instance projection. AuthConfig is the RAW
// stored patch (often `{}` = all defaults) and round-trips back into PATCH;
// AuthConfigResolved is the EFFECTIVE config with every default expanded, so a
// caller can inspect the live auth behaviour. Both are kept as json.RawMessage
// because the shape is freeform and deeply nested — the provider surfaces them
// as JSON strings rather than modelling every sub-section.
type Instance struct {
	ID                 string          `json:"id"`
	Environment        string          `json:"environment"`
	PublishableKey     string          `json:"publishable_key"`
	FrontendAPIHost    string          `json:"frontend_api_host"`
	AllowedOrigins     []string        `json:"allowed_origins"`
	AuthConfig         json.RawMessage `json:"auth_config"`
	AuthConfigResolved json.RawMessage `json:"auth_config_resolved"`
	CreatedAt          int64           `json:"created_at"`
}

// InstanceUpdate is the PATCH /v1/instance body. Both fields are optional; a nil
// field is omitted so a PATCH touches only what the caller supplied. AuthConfig
// is a PARTIAL patch merged server-side into the current config (never a
// wholesale replacement), matching the dashboard's merge semantics.
type InstanceUpdate struct {
	AllowedOrigins []string        `json:"allowed_origins,omitempty"`
	AuthConfig     json.RawMessage `json:"auth_config,omitempty"`
}

// GetInstance reads the singleton instance config. GET /v1/instance.
func (c *Client) GetInstance(ctx context.Context) (*Instance, error) {
	var out Instance
	if err := c.do(ctx, http.MethodGet, "/v1/instance", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateInstance applies a partial PATCH to the instance config. PATCH
// /v1/instance returns only { id, allowed_origins, auth_config,
// auth_config_resolved }, so the instance_config resource re-reads via
// GetInstance to populate the remaining computed fields.
func (c *Client) UpdateInstance(ctx context.Context, body InstanceUpdate) (*Instance, error) {
	var out Instance
	if err := c.do(ctx, http.MethodPatch, "/v1/instance", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ─────────────────────────── Redirect URLs ───────────────────────────

// RedirectURL mirrors the §9.3 redirect-URL allowlist projection. The API
// NORMALISES the stored url on create (bare scheme://host[:port] with the path
// dropped when it is just "/"), so the value read back may differ from the
// string supplied.
type RedirectURL struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	CreatedAt int64  `json:"created_at"`
}

type redirectURLCreate struct {
	URL string `json:"url"`
}

// CreateRedirectURL adds an allowlisted redirect URL. POST /v1/redirect_urls.
func (c *Client) CreateRedirectURL(ctx context.Context, u string) (*RedirectURL, error) {
	var out RedirectURL
	if err := c.do(ctx, http.MethodPost, "/v1/redirect_urls", redirectURLCreate{URL: u}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRedirectURL reads one by id. GET /v1/redirect_urls/:id.
func (c *Client) GetRedirectURL(ctx context.Context, id string) (*RedirectURL, error) {
	var out RedirectURL
	if err := c.do(ctx, http.MethodGet, "/v1/redirect_urls/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteRedirectURL removes one by id. DELETE /v1/redirect_urls/:id. There is no
// update route — the url is immutable, so a change replaces the resource.
func (c *Client) DeleteRedirectURL(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/redirect_urls/"+url.PathEscape(id), nil, nil)
}

// ─────────────────────────── Billing plans (end-user) ───────────────────────────

// BillingPlan mirrors the §billing bapi-billing-plans projection — the plans a
// tenant defines for their app's users. StripePriceID is null for the free tier
// (then Free is true). Amount is the smallest currency unit (cents) kept as a
// string display cache of the Stripe price.
type BillingPlan struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Slug          string   `json:"slug"`
	StripePriceID *string  `json:"stripe_price_id"`
	Free          bool     `json:"free"`
	Interval      string   `json:"interval"`
	Amount        *string  `json:"amount"`
	Currency      string   `json:"currency"`
	Features      []string `json:"features"`
	Audience      string   `json:"audience"`
	Active        bool     `json:"active"`
	PricingModel  string   `json:"pricing_model"`
	TrialDays     *int64   `json:"trial_days"`
	StripeMeterID *string  `json:"stripe_meter_id"`
	UsageUnit     *string  `json:"usage_unit"`
	CreatedAt     int64    `json:"created_at"`
	UpdatedAt     int64    `json:"updated_at"`
}

// BillingPlanWrite is the create/update payload. Pointer/omitempty fields let an
// update PATCH only the keys that changed.
type BillingPlanWrite struct {
	Name          *string  `json:"name,omitempty"`
	Slug          *string  `json:"slug,omitempty"`
	StripePriceID *string  `json:"stripe_price_id,omitempty"`
	Interval      *string  `json:"interval,omitempty"`
	Amount        *string  `json:"amount,omitempty"`
	Currency      *string  `json:"currency,omitempty"`
	Features      []string `json:"features,omitempty"`
	Audience      *string  `json:"audience,omitempty"`
	Active        *bool    `json:"active,omitempty"`
	PricingModel  *string  `json:"pricing_model,omitempty"`
	TrialDays     *int64   `json:"trial_days,omitempty"`
	StripeMeterID *string  `json:"stripe_meter_id,omitempty"`
	UsageUnit     *string  `json:"usage_unit,omitempty"`
}

func (c *Client) CreateBillingPlan(ctx context.Context, body BillingPlanWrite) (*BillingPlan, error) {
	var out BillingPlan
	if err := c.do(ctx, http.MethodPost, "/v1/billing/plans", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetBillingPlan(ctx context.Context, id string) (*BillingPlan, error) {
	var out BillingPlan
	if err := c.do(ctx, http.MethodGet, "/v1/billing/plans/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateBillingPlan(ctx context.Context, id string, body BillingPlanWrite) (*BillingPlan, error) {
	var out BillingPlan
	if err := c.do(ctx, http.MethodPatch, "/v1/billing/plans/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteBillingPlan(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/billing/plans/"+url.PathEscape(id), nil, nil)
}

// ─────────────────────────── Organization policy ───────────────────────────

// OrgPolicy mirrors the §4.4 org-scoped security policy (the resolved view the
// API returns under `policy`). Every field is an OPTIONAL tightening on top of
// the instance policy; an absent field inherits instance behaviour.
type OrgPolicy struct {
	RequireMfa            *bool    `json:"requireMfa,omitempty"`
	SsoRequired           *bool    `json:"ssoRequired,omitempty"`
	SessionIdleOverrideMs *int64   `json:"sessionIdleOverrideMs,omitempty"`
	AllowedSignInMethods  []string `json:"allowedSignInMethods,omitempty"`
	IPAllowlist           []string `json:"ipAllowlist,omitempty"`
	MaxSessionAgeSeconds  *int64   `json:"maxSessionAgeSeconds,omitempty"`
}

// OrgPolicyPatch is the PATCH body. A field left nil is omitted (untouched); the
// resource sends an explicit value to set and the API treats a JSON null as a
// CLEAR. The provider models set-or-inherit, so it only ever sends values or
// omits — it does not emit nulls.
type OrgPolicyPatch struct {
	RequireMfa            *bool    `json:"requireMfa,omitempty"`
	SsoRequired           *bool    `json:"ssoRequired,omitempty"`
	SessionIdleOverrideMs *int64   `json:"sessionIdleOverrideMs,omitempty"`
	AllowedSignInMethods  []string `json:"allowedSignInMethods,omitempty"`
	IPAllowlist           []string `json:"ipAllowlist,omitempty"`
	MaxSessionAgeSeconds  *int64   `json:"maxSessionAgeSeconds,omitempty"`
}

// orgPolicyResponse is the { organization_id, policy } envelope PATCH returns.
type orgPolicyResponse struct {
	OrganizationID string    `json:"organization_id"`
	Policy         OrgPolicy `json:"policy"`
}

// SetOrgPolicy merges a policy patch into an organization's policy bag.
// PATCH /v1/organizations/:id/policy — returns the resolved policy.
func (c *Client) SetOrgPolicy(ctx context.Context, orgID string, body OrgPolicyPatch) (*OrgPolicy, error) {
	var out orgPolicyResponse
	if err := c.do(ctx, http.MethodPatch, "/v1/organizations/"+url.PathEscape(orgID)+"/policy", body, &out); err != nil {
		return nil, err
	}
	return &out.Policy, nil
}

// GetOrgPolicy reads an organization's resolved policy. Atlas exposes no
// read-only GET for org policy (the org projection omits it), so the current
// policy is read via an idempotent empty PATCH — an empty patch validates
// cleanly and merges nothing, returning the policy unchanged. A 404 for the
// organization propagates as NOT_FOUND so a deleted org drops from state.
func (c *Client) GetOrgPolicy(ctx context.Context, orgID string) (*OrgPolicy, error) {
	return c.SetOrgPolicy(ctx, orgID, OrgPolicyPatch{})
}
