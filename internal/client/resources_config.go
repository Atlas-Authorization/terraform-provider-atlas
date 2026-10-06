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
// (then Free is true). Amount is a free-form display string stored verbatim by
// Atlas — not parsed, validated or converted; the tenant chooses the units
// (e.g. "12.00" for $12, or "1200" for cents).
type BillingPlan struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Slug          string             `json:"slug"`
	StripePriceID *string            `json:"stripe_price_id"`
	Free          bool               `json:"free"`
	Interval      string             `json:"interval"`
	Amount        *string            `json:"amount"`
	Currency      string             `json:"currency"`
	Features      []string           `json:"features"`
	Limits        map[string]float64 `json:"limits"`
	Audience      string             `json:"audience"`
	Active        bool               `json:"active"`
	PricingModel  string             `json:"pricing_model"`
	TrialDays     *int64             `json:"trial_days"`
	StripeMeterID *string            `json:"stripe_meter_id"`
	UsageUnit     *string            `json:"usage_unit"`
	CreatedAt     int64              `json:"created_at"`
	UpdatedAt     int64              `json:"updated_at"`
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
	// Limits is a pointer so an explicit empty map (clear all limits) is sent as {}.
	Limits        *map[string]float64 `json:"limits,omitempty"`
	Audience      *string             `json:"audience,omitempty"`
	Active        *bool               `json:"active,omitempty"`
	PricingModel  *string             `json:"pricing_model,omitempty"`
	TrialDays     *int64              `json:"trial_days,omitempty"`
	StripeMeterID *string             `json:"stripe_meter_id,omitempty"`
	UsageUnit     *string             `json:"usage_unit,omitempty"`
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

// ─────────────────────── Org settings schema (P1-5) ───────────────────────

// OrgSettingsSchema is the per-instance JSON Schema registry entry. Schema is
// the raw document; Version is bumped by the server on every (re)registration.
type OrgSettingsSchema struct {
	Schema    json.RawMessage `json:"schema"`
	Version   int64           `json:"version"`
	UpdatedAt int64           `json:"updated_at"`
}

// PutOrgSettingsSchema registers (or replaces) the instance's schema. PUT is the
// idempotent form.
func (c *Client) PutOrgSettingsSchema(ctx context.Context, schema json.RawMessage) (*OrgSettingsSchema, error) {
	var out OrgSettingsSchema
	body := map[string]json.RawMessage{"schema": schema}
	if err := c.do(ctx, http.MethodPut, "/v1/organization_settings_schema", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetOrgSettingsSchema(ctx context.Context) (*OrgSettingsSchema, error) {
	var out OrgSettingsSchema
	if err := c.do(ctx, http.MethodGet, "/v1/organization_settings_schema", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ───────────────────────── Notification templates (P2-12) ─────────────────────────

// NotificationTemplate is a tenant-authored notification template, keyed by name.
type NotificationTemplate struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Category  string `json:"category"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// NotificationTemplateWrite is the create/update payload (Name is create-only).
type NotificationTemplateWrite struct {
	Name     string `json:"name,omitempty"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	Category string `json:"category"`
}

func (c *Client) CreateNotificationTemplate(ctx context.Context, body NotificationTemplateWrite) (*NotificationTemplate, error) {
	var out NotificationTemplate
	if err := c.do(ctx, http.MethodPost, "/v1/notification_templates", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetNotificationTemplate(ctx context.Context, name string) (*NotificationTemplate, error) {
	var out NotificationTemplate
	if err := c.do(ctx, http.MethodGet, "/v1/notification_templates/"+url.PathEscape(name), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateNotificationTemplate(ctx context.Context, name string, body NotificationTemplateWrite) (*NotificationTemplate, error) {
	body.Name = "" // the name is the URL key and immutable
	var out NotificationTemplate
	if err := c.do(ctx, http.MethodPut, "/v1/notification_templates/"+url.PathEscape(name), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteNotificationTemplate(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/v1/notification_templates/"+url.PathEscape(name), nil, nil)
}

// ─────────────────────── Notification categories (round-7 #9) ───────────────────────

// NotificationCategory is a tenant-defined end-user notification category,
// keyed by `key`, with a user-facing `label` shown on the preference toggle.
// Tenant categories are always optional (they never gate a built-in).
type NotificationCategory struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Label     string `json:"label"`
	Optional  bool   `json:"optional"`
	CreatedAt int64  `json:"created_at"`
}

// NotificationCategoryWrite is the create payload (`key` is create-only; the
// update route takes `label` alone).
type NotificationCategoryWrite struct {
	Key   string `json:"key,omitempty"`
	Label string `json:"label"`
}

type notificationCategoryList struct {
	Data []NotificationCategory `json:"data"`
}

func (c *Client) CreateNotificationCategory(ctx context.Context, body NotificationCategoryWrite) (*NotificationCategory, error) {
	var out NotificationCategory
	if err := c.do(ctx, http.MethodPost, "/v1/notification_categories", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNotificationCategory reads one category via the list endpoint (there is no
// get-by-key route). An unknown key returns a synthetic 404.
func (c *Client) GetNotificationCategory(ctx context.Context, key string) (*NotificationCategory, error) {
	var out notificationCategoryList
	if err := c.do(ctx, http.MethodGet, "/v1/notification_categories", nil, &out); err != nil {
		return nil, err
	}
	for i := range out.Data {
		if out.Data[i].Key == key {
			return &out.Data[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Code: "NOT_FOUND", Message: "Unknown notification category."}}}
}

func (c *Client) UpdateNotificationCategory(ctx context.Context, key string, body NotificationCategoryWrite) (*NotificationCategory, error) {
	body.Key = "" // the key is the URL path and immutable
	var out NotificationCategory
	if err := c.do(ctx, http.MethodPut, "/v1/notification_categories/"+url.PathEscape(key), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteNotificationCategory removes the category. The API returns 409 if a
// notification template still references it.
func (c *Client) DeleteNotificationCategory(ctx context.Context, key string) error {
	return c.do(ctx, http.MethodDelete, "/v1/notification_categories/"+url.PathEscape(key), nil, nil)
}

// ───────────────────────────── Email templates (§11.1) ─────────────────────────────

// EmailTemplateOverride is the stored copy override for one email template.
type EmailTemplateOverride struct {
	Subject *string `json:"subject,omitempty"`
	Text    *string `json:"text,omitempty"`
}

// EmailTemplate is one entry of the email-template list: the built-in default,
// the instance override (nil when not customised) and the customised flag.
type EmailTemplate struct {
	Name       string                 `json:"name"`
	Override   *EmailTemplateOverride `json:"override"`
	Customised bool                   `json:"customised"`
}

type emailTemplateList struct {
	Data []EmailTemplate `json:"data"`
}

// GetEmailTemplate reads one template via the list endpoint (there is no
// single-item GET). An unknown name returns a synthetic 404.
func (c *Client) GetEmailTemplate(ctx context.Context, name string) (*EmailTemplate, error) {
	var out emailTemplateList
	if err := c.do(ctx, http.MethodGet, "/v1/email_templates", nil, &out); err != nil {
		return nil, err
	}
	for i := range out.Data {
		if out.Data[i].Name == name {
			return &out.Data[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Code: "NOT_FOUND", Message: "Unknown email template."}}}
}

// PutEmailTemplate saves the override; the server validates {{placeholders}}.
func (c *Client) PutEmailTemplate(ctx context.Context, name string, body EmailTemplateOverride) (*EmailTemplate, error) {
	var out EmailTemplate
	if err := c.do(ctx, http.MethodPut, "/v1/email_templates/"+url.PathEscape(name), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteEmailTemplate reverts the template to the built-in copy.
func (c *Client) DeleteEmailTemplate(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/v1/email_templates/"+url.PathEscape(name), nil, nil)
}
