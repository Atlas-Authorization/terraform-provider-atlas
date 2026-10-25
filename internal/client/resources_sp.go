package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Provisioning tags the integrator spec added to several resources:
// external_ref (an idempotent-create / adoption key) and metadata (filterable
// tags). Metadata values are strings, numbers or booleans on the wire.

// ─────────────────────────────── Whoami ───────────────────────────────

// Whoami is GET /v1/whoami: what the configured secret key is.
type Whoami struct {
	Key struct {
		ID              string   `json:"id"`
		Type            string   `json:"type"`
		Prefix          string   `json:"prefix"`
		Environment     *string  `json:"environment"`
		Scopes          []string `json:"scopes"`
		EffectiveScopes []string `json:"effective_scopes"`
		Owner           *string  `json:"owner"`
		AllowedIPs      []string `json:"allowed_ips"`
		CreatedAt       *int64   `json:"created_at"`
		ExpiresAt       *int64   `json:"expires_at"`
		LastUsedAt      *int64   `json:"last_used_at"`
		LastUsedIP      *string  `json:"last_used_ip"`
	} `json:"key"`
	Instance struct {
		ID          string  `json:"id"`
		Environment *string `json:"environment"`
	} `json:"instance"`
	Account struct {
		ID *string `json:"id"`
	} `json:"account"`
}

// GetWhoami reads GET /v1/whoami.
func (c *Client) GetWhoami(ctx context.Context) (*Whoami, error) {
	var out Whoami
	if err := c.do(ctx, http.MethodGet, "/v1/whoami", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ───────────────────────────── Secret keys ─────────────────────────────

// SecretKey mirrors the /v1/secret_keys projection. Secret is only present on
// create (and rotate); a read never returns it.
type SecretKey struct {
	ID             string                 `json:"id"`
	Name           *string                `json:"name"`
	Owner          *string                `json:"owner"`
	Prefix         string                 `json:"prefix"`
	Format         string                 `json:"format"`
	Environment    *string                `json:"environment"`
	Scopes         []string               `json:"scopes"`
	Status         string                 `json:"status"`
	ExpiresAt      *int64                 `json:"expires_at"`
	AllowedIPs     []string               `json:"allowed_ips"`
	AllowedOrigins []string               `json:"allowed_origins"`
	LastUsedAt     *int64                 `json:"last_used_at"`
	LastUsedIP     *string                `json:"last_used_ip"`
	CreatedAt      int64                  `json:"created_at"`
	RevokedAt      *int64                 `json:"revoked_at"`
	ExternalRef    *string                `json:"external_ref"`
	Metadata       map[string]interface{} `json:"metadata"`
	Secret         *string                `json:"secret"`
}

// SecretKeyCreate is the POST /v1/secret_keys body.
type SecretKeyCreate struct {
	Name           *string                `json:"name,omitempty"`
	Scopes         []string               `json:"scopes,omitempty"`
	ExpiresAt      *string                `json:"expires_at,omitempty"`
	AllowedIPs     []string               `json:"allowed_ips,omitempty"`
	AllowedOrigins []string               `json:"allowed_origins,omitempty"`
	Owner          *string                `json:"owner,omitempty"`
	ExternalRef    *string                `json:"external_ref,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// SecretKeyUpdate is the PATCH /v1/secret_keys/:id body. The lists are sent as
// empty arrays (not omitted) when non-nil, so a list can be cleared.
type SecretKeyUpdate struct {
	Name           *string                `json:"name,omitempty"`
	Owner          *string                `json:"owner,omitempty"`
	ExpiresAt      *string                `json:"expires_at,omitempty"`
	AllowedIPs     *[]string              `json:"allowed_ips,omitempty"`
	AllowedOrigins *[]string              `json:"allowed_origins,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

func (c *Client) CreateSecretKey(ctx context.Context, body SecretKeyCreate) (*SecretKey, error) {
	var out SecretKey
	if err := c.do(ctx, http.MethodPost, "/v1/secret_keys", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetSecretKey(ctx context.Context, id string) (*SecretKey, error) {
	var out SecretKey
	if err := c.do(ctx, http.MethodGet, "/v1/secret_keys/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateSecretKey(ctx context.Context, id string, body SecretKeyUpdate) (*SecretKey, error) {
	var out SecretKey
	if err := c.do(ctx, http.MethodPatch, "/v1/secret_keys/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RevokeSecretKey revokes the key (POST /v1/secret_keys/:id/revoke). A key
// that is already revoked answers 409, which the caller treats as done.
func (c *Client) RevokeSecretKey(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, "/v1/secret_keys/"+url.PathEscape(id)+"/revoke", nil, nil)
}

// ───────────────────────────── Partner links ─────────────────────────────

// PartnerLink mirrors the inbound /v1/partner_links projection. ClientSecret is
// only present on create and rotate.
type PartnerLink struct {
	ID               string   `json:"id"`
	ClientID         string   `json:"client_id"`
	Partner          string   `json:"partner"`
	Name             string   `json:"name"`
	Preset           bool     `json:"preset"`
	Resource         *string  `json:"resource"`
	Scopes           []string `json:"scopes"`
	SecretPrefix     *string  `json:"client_secret_prefix"`
	TokenEndpoint    string   `json:"token_endpoint"`
	CertFingerprints []string `json:"cert_fingerprints"`
	AuthMethods      []string `json:"auth_methods"`
	RequestID        *string  `json:"request_id"`
	Status           string   `json:"status"`
	ExpiresAt        *int64   `json:"expires_at"`
	RevokedAt        *int64   `json:"revoked_at"`
	LastUsedAt       *int64   `json:"last_used_at"`
	Calls30d         int64    `json:"calls_30d"`
	CreatedAt        int64    `json:"created_at"`
	ClientSecret     *string  `json:"client_secret"`
}

// PartnerLinkCreate is the POST /v1/partner_links body.
type PartnerLinkCreate struct {
	Partner          string   `json:"partner"`
	Name             *string  `json:"name,omitempty"`
	Scopes           []string `json:"scopes"`
	ExpiresAt        *string  `json:"expires_at,omitempty"`
	CertFingerprints []string `json:"cert_fingerprints,omitempty"`
}

func (c *Client) CreatePartnerLink(ctx context.Context, body PartnerLinkCreate) (*PartnerLink, error) {
	var out PartnerLink
	if err := c.do(ctx, http.MethodPost, "/v1/partner_links", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetPartnerLink(ctx context.Context, id string) (*PartnerLink, error) {
	var out PartnerLink
	if err := c.do(ctx, http.MethodGet, "/v1/partner_links/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetPartnerLinkCertFingerprints locks the link to client certificates (mTLS);
// an empty slice removes the lock. PATCH /v1/partner_links/:id.
func (c *Client) SetPartnerLinkCertFingerprints(ctx context.Context, id string, fingerprints []string) (*PartnerLink, error) {
	if fingerprints == nil {
		fingerprints = []string{}
	}
	var out PartnerLink
	body := map[string]interface{}{"cert_fingerprints": fingerprints}
	if err := c.do(ctx, http.MethodPatch, "/v1/partner_links/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RotatePartnerLinkSecret mints a new client secret (shown once); the old one
// stops working at once. The link keeps its id.
func (c *Client) RotatePartnerLinkSecret(ctx context.Context, id string) (*PartnerLink, error) {
	var out PartnerLink
	if err := c.do(ctx, http.MethodPost, "/v1/partner_links/"+url.PathEscape(id)+"/rotate_secret", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RevokePartnerLink revokes the link (DELETE); tokens minted from it stop at once.
func (c *Client) RevokePartnerLink(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/partner_links/"+url.PathEscape(id), nil, nil)
}

// OutboundPartnerLink is a link a partner granted Atlas, so Atlas can call it.
type OutboundPartnerLink struct {
	ID         string   `json:"id"`
	Partner    string   `json:"partner"`
	Name       string   `json:"name"`
	Preset     bool     `json:"preset"`
	Resource   *string  `json:"resource"`
	Scopes     []string `json:"scopes"`
	BaseURL    *string  `json:"base_url"`
	TokenURL   *string  `json:"token_url"`
	ClientID   string   `json:"client_id"`
	AuthMethod string   `json:"auth_method"`
	Source     string   `json:"source"`
	Status     string   `json:"status"`
	ExpiresAt  *int64   `json:"expires_at"`
	// Cmp is the partner's consent manager; CmpDefault is true when it is
	// derived (a Cookie Munch link with a site:<id> resource), not set.
	Cmp        map[string]interface{} `json:"cmp"`
	CmpDefault bool                   `json:"cmp_default"`
	LastError  *string                `json:"last_error"`
	CreatedAt  int64                  `json:"created_at"`
	UpdatedAt  int64                  `json:"updated_at"`
}

// OutboundPartnerLinkWrite is the PUT /v1/partner_links/outbound/:partner body.
type OutboundPartnerLinkWrite struct {
	Name         *string  `json:"name,omitempty"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
	Resource     *string  `json:"resource,omitempty"`
	BaseURL      *string  `json:"base_url,omitempty"`
	TokenURL     *string  `json:"token_url,omitempty"`
	AuthMethod   *string  `json:"auth_method,omitempty"`
	ExpiresAt    *string  `json:"expires_at,omitempty"`
	// Cmp is raw JSON: nil omits it (the API keeps the stored one), `null`
	// clears it, an object sets it.
	Cmp json.RawMessage `json:"cmp,omitempty"`
}

func (c *Client) PutOutboundPartnerLink(ctx context.Context, partner string, body OutboundPartnerLinkWrite) (*OutboundPartnerLink, error) {
	var out OutboundPartnerLink
	if err := c.do(ctx, http.MethodPut, "/v1/partner_links/outbound/"+url.PathEscape(partner), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOutboundPartnerLink lists the outbound links and picks one by partner id.
func (c *Client) GetOutboundPartnerLink(ctx context.Context, partner string) (*OutboundPartnerLink, error) {
	var out listEnvelope[OutboundPartnerLink]
	if err := c.do(ctx, http.MethodGet, "/v1/partner_links/outbound", nil, &out); err != nil {
		return nil, err
	}
	for i := range out.Data {
		if out.Data[i].Partner == partner {
			return &out.Data[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Code: "NOT_FOUND", Message: "Unknown outbound partner link."}}}
}

func (c *Client) DeleteOutboundPartnerLink(ctx context.Context, partner string) error {
	return c.do(ctx, http.MethodDelete, "/v1/partner_links/outbound/"+url.PathEscape(partner), nil, nil)
}

// ───────────────────────────── Integrations ─────────────────────────────

// Integration mirrors GET /v1/instance/integrations/:id. An integration that is
// not connected answers { id, enabled: false, connected: false } only.
type Integration struct {
	ID                      string                 `json:"id"`
	Preset                  *string                `json:"preset"`
	Name                    *string                `json:"name"`
	Role                    *string                `json:"role"`
	Enabled                 bool                   `json:"enabled"`
	Connected               bool                   `json:"connected"`
	Audience                *string                `json:"audience"`
	AudienceFormat          *string                `json:"audience_format"`
	PrincipalClaim          *string                `json:"principal_claim"`
	ExtClaim                *string                `json:"ext_claim"`
	TTLSeconds              *int64                 `json:"ttl_seconds"`
	Issuer                  *string                `json:"issuer"`
	JWKSURL                 *string                `json:"jwks_url"`
	EventsURL               *string                `json:"events_url"`
	BackchannelLogoutURL    *string                `json:"backchannel_logout_url"`
	BackchannelLogoutSource *string                `json:"backchannel_logout_source"`
	VerifyErrorReportURL    *string                `json:"verify_error_report_url"`
	OAuthClientID           *string                `json:"oauth_client_id"`
	AtlasIDUsers            *bool                  `json:"atlas_id_users"`
	ExternalRef             *string                `json:"external_ref"`
	Metadata                map[string]interface{} `json:"metadata"`
	// Cmp is the consent manager (script + consent_api adapter) the hosted
	// pages load; nil when none. For the Cookie Munch preset it may be the
	// preset's derived default.
	Cmp map[string]interface{} `json:"cmp"`
	// PerCallAudience is { param, template, allowed, any_matching, default };
	// nil when off.
	PerCallAudience map[string]interface{} `json:"per_call_audience"`
	EnabledAt       *int64                 `json:"enabled_at"`
	UpdatedAt       *int64                 `json:"updated_at"`
}

// PutIntegration is PUT /v1/instance/integrations/:id. The body is the whole
// connection (the API replaces the stored record), built by the caller as a
// map so a field can be sent as JSON null to switch a default off.
func (c *Client) PutIntegration(ctx context.Context, id string, body map[string]interface{}) (*Integration, error) {
	var out Integration
	if err := c.do(ctx, http.MethodPut, "/v1/instance/integrations/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetIntegration(ctx context.Context, id string) (*Integration, error) {
	var out Integration
	if err := c.do(ctx, http.MethodGet, "/v1/instance/integrations/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteIntegration(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/instance/integrations/"+url.PathEscape(id), nil, nil)
}

// ───────────────────────────── Webhook endpoints (full) ─────────────────────────────

// WebhookEndpointFull is the complete /v1/webhook_endpoints/:id projection,
// including the 2026-10 fields. WebhookEndpoint (resources.go) keeps the
// original subset.
type WebhookEndpointFull struct {
	ID                  string                 `json:"id"`
	URL                 *string                `json:"url"`
	Description         *string                `json:"description"`
	EnabledEvents       []string               `json:"enabled_events"`
	EventNaming         string                 `json:"event_naming"`
	Active              bool                   `json:"active"`
	External            bool                   `json:"external"`
	SignatureVersions   []string               `json:"signature_versions"`
	Destination         map[string]interface{} `json:"destination"`
	FailingSince        *int64                 `json:"failing_since"`
	DisabledAt          *int64                 `json:"disabled_at"`
	LastFailingSince    *int64                 `json:"last_failing_since"`
	Health              string                 `json:"health"`
	LastSuccessAt       *int64                 `json:"last_success_at"`
	ConsecutiveFailures int64                  `json:"consecutive_failures"`
	AlertContacts       []string               `json:"alert_contacts"`
	CreatedAt           int64                  `json:"created_at"`
	UpdatedAt           int64                  `json:"updated_at"`
	ExternalRef         *string                `json:"external_ref"`
	Metadata            map[string]interface{} `json:"metadata"`
	Secret              *string                `json:"secret"`
}

// WebhookEndpointWrite is the create / PATCH body with every field. Nil
// pointers and slices are omitted.
type WebhookEndpointWrite struct {
	URL                    *string                `json:"url,omitempty"`
	Description            *string                `json:"description,omitempty"`
	EnabledEvents          []string               `json:"enabled_events,omitempty"`
	EventNaming            *string                `json:"event_naming,omitempty"`
	Active                 *bool                  `json:"active,omitempty"`
	External               *bool                  `json:"external,omitempty"`
	SignatureVersions      []string               `json:"signature_versions,omitempty"`
	AlertContacts          *[]string              `json:"alert_contacts,omitempty"`
	Destination            map[string]interface{} `json:"destination,omitempty"`
	DestinationCredentials map[string]interface{} `json:"destination_credentials,omitempty"`
	ExternalRef            *string                `json:"external_ref,omitempty"`
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
}

func (c *Client) CreateWebhookEndpointFull(ctx context.Context, body WebhookEndpointWrite) (*WebhookEndpointFull, error) {
	var out WebhookEndpointFull
	if err := c.do(ctx, http.MethodPost, "/v1/webhook_endpoints", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWebhookEndpointFull reads GET /v1/webhook_endpoints/:id.
func (c *Client) GetWebhookEndpointFull(ctx context.Context, id string) (*WebhookEndpointFull, error) {
	var out WebhookEndpointFull
	if err := c.do(ctx, http.MethodGet, "/v1/webhook_endpoints/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateWebhookEndpointFull(ctx context.Context, id string, body WebhookEndpointWrite) (*WebhookEndpointFull, error) {
	var out WebhookEndpointFull
	if err := c.do(ctx, http.MethodPatch, "/v1/webhook_endpoints/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ───────────────────────────── Audit exports ─────────────────────────────

// AuditExport mirrors /v1/audit_exports/:id. Destination secrets come back as
// has_<field> booleans; the webhook signing Secret is only present on create
// and rotate.
type AuditExport struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Type                string                 `json:"type"`
	Enabled             bool                   `json:"enabled"`
	Status              string                 `json:"status"`
	ActionFilter        []string               `json:"action_filter"`
	Destination         map[string]interface{} `json:"destination"`
	ConsecutiveFailures int64                  `json:"consecutive_failures"`
	LastError           *string                `json:"last_error"`
	LastDeliveredAt     *int64                 `json:"last_delivered_at"`
	ExportedCount       int64                  `json:"exported_count"`
	CreatedAt           int64                  `json:"created_at"`
	UpdatedAt           int64                  `json:"updated_at"`
	Secret              *string                `json:"secret"`
}

// AuditExportWrite is the create / PATCH body.
type AuditExportWrite struct {
	Name          *string           `json:"name,omitempty"`
	Type          *string           `json:"type,omitempty"`
	Destination   map[string]string `json:"destination,omitempty"`
	ActionFilter  interface{}       `json:"action_filter,omitempty"`
	Enabled       *bool             `json:"enabled,omitempty"`
	BackfillHours *int64            `json:"backfill_hours,omitempty"`
}

func (c *Client) CreateAuditExport(ctx context.Context, body AuditExportWrite) (*AuditExport, error) {
	var out AuditExport
	if err := c.do(ctx, http.MethodPost, "/v1/audit_exports", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetAuditExport(ctx context.Context, id string) (*AuditExport, error) {
	var out AuditExport
	if err := c.do(ctx, http.MethodGet, "/v1/audit_exports/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateAuditExport(ctx context.Context, id string, body AuditExportWrite) (*AuditExport, error) {
	var out AuditExport
	if err := c.do(ctx, http.MethodPatch, "/v1/audit_exports/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RotateAuditExportSecret(ctx context.Context, id string) (*AuditExport, error) {
	var out AuditExport
	if err := c.do(ctx, http.MethodPost, "/v1/audit_exports/"+url.PathEscape(id)+"/rotate_secret", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteAuditExport(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/audit_exports/"+url.PathEscape(id), nil, nil)
}

// ───────────────────────────── Status ─────────────────────────────

// StatusSubscription mirrors /v1/instance/status_subscription (a singleton).
type StatusSubscription struct {
	Subscribed         bool     `json:"subscribed"`
	Enabled            bool     `json:"enabled"`
	Regions            []string `json:"regions"`
	ComponentIDs       []string `json:"component_ids"`
	MinImpact          string   `json:"min_impact"`
	IncludeMaintenance bool     `json:"include_maintenance"`
	EmailOwners        bool     `json:"email_owners"`
	Webhook            bool     `json:"webhook"`
	RSSURL             *string  `json:"rss_url"`
	AtomURL            *string  `json:"atom_url"`
	UpdatedAt          *int64   `json:"updated_at"`
}

// StatusSubscriptionWrite is the PUT body; nil fields are left as they are.
type StatusSubscriptionWrite struct {
	Enabled            *bool     `json:"enabled,omitempty"`
	Regions            *[]string `json:"regions,omitempty"`
	ComponentIDs       *[]string `json:"component_ids,omitempty"`
	MinImpact          *string   `json:"min_impact,omitempty"`
	IncludeMaintenance *bool     `json:"include_maintenance,omitempty"`
	EmailOwners        *bool     `json:"email_owners,omitempty"`
	Webhook            *bool     `json:"webhook,omitempty"`
}

func (c *Client) GetStatusSubscription(ctx context.Context) (*StatusSubscription, error) {
	var out StatusSubscription
	if err := c.do(ctx, http.MethodGet, "/v1/instance/status_subscription", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) PutStatusSubscription(ctx context.Context, body StatusSubscriptionWrite) (*StatusSubscription, error) {
	var out StatusSubscription
	if err := c.do(ctx, http.MethodPut, "/v1/instance/status_subscription", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DisableStatusSubscription switches the subscription off (DELETE keeps the row).
func (c *Client) DisableStatusSubscription(ctx context.Context) error {
	return c.do(ctx, http.MethodDelete, "/v1/instance/status_subscription", nil, nil)
}

// StatusDependency mirrors /v1/instance/status_dependencies/:id.
type StatusDependency struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	URL             string   `json:"url"`
	ComponentFilter []string `json:"component_filter"`
	Enabled         bool     `json:"enabled"`
	State           string   `json:"state"`
	Indicator       *string  `json:"indicator"`
	Description     *string  `json:"description"`
	LastCheckedAt   *int64   `json:"last_checked_at"`
	LastSuccessAt   *int64   `json:"last_success_at"`
	LastError       *string  `json:"last_error"`
	Stale           bool     `json:"stale"`
	CreatedAt       int64    `json:"created_at"`
}

// StatusDependencyWrite is the create / PATCH body (url is create-only).
type StatusDependencyWrite struct {
	Name            *string   `json:"name,omitempty"`
	URL             *string   `json:"url,omitempty"`
	ComponentFilter *[]string `json:"component_filter,omitempty"`
	Enabled         *bool     `json:"enabled,omitempty"`
}

func (c *Client) CreateStatusDependency(ctx context.Context, body StatusDependencyWrite) (*StatusDependency, error) {
	var out StatusDependency
	if err := c.do(ctx, http.MethodPost, "/v1/instance/status_dependencies", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetStatusDependency(ctx context.Context, id string) (*StatusDependency, error) {
	var out StatusDependency
	if err := c.do(ctx, http.MethodGet, "/v1/instance/status_dependencies/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateStatusDependency(ctx context.Context, id string, body StatusDependencyWrite) (*StatusDependency, error) {
	var out StatusDependency
	if err := c.do(ctx, http.MethodPatch, "/v1/instance/status_dependencies/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteStatusDependency(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/instance/status_dependencies/"+url.PathEscape(id), nil, nil)
}

// ───────────────────────────── Partner platforms ─────────────────────────────

// PartnerPlatform is this instance as a partner platform of another product
// (/v1/partner_platforms/:partner). It never carries the partner token: only
// whether one is stored and its short prefix.
type PartnerPlatform struct {
	ID                  string   `json:"id"`
	Partner             string   `json:"partner"`
	Name                string   `json:"name"`
	Preset              bool     `json:"preset"`
	PlatformID          string   `json:"platform_id"`
	Enabled             bool     `json:"enabled"`
	JWKSURL             string   `json:"jwks_url"`
	SignatureHeader     string   `json:"signature_header"`
	SignatureScheme     string   `json:"signature_scheme"`
	ToleranceSeconds    int64    `json:"tolerance_seconds"`
	AllowedSources      []string `json:"allowed_sources"`
	RecordConsents      bool     `json:"record_consents"`
	ReceiverURL         string   `json:"receiver_url"`
	ReceiverURLTemplate string   `json:"receiver_url_template"`
	Registration        struct {
		Status       string `json:"status"`
		RegisteredAt *int64 `json:"registered_at"`
	} `json:"registration"`
	TokenSet          bool    `json:"token_set"`
	TokenPrefix       *string `json:"token_prefix"`
	LastDeliveryAt    *int64  `json:"last_delivery_at"`
	LastDeliveryError *string `json:"last_delivery_error"`
}

func partnerPlatformPath(partner string) string {
	return "/v1/partner_platforms/" + url.PathEscape(partner)
}

// GetPartnerPlatform is GET /v1/partner_platforms/:partner.
func (c *Client) GetPartnerPlatform(ctx context.Context, partner string) (*PartnerPlatform, error) {
	var out PartnerPlatform
	if err := c.do(ctx, http.MethodGet, partnerPlatformPath(partner), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PutPartnerPlatform is PUT /v1/partner_platforms/:partner. The body is partial:
// a key left out is kept (a preset fills the rest on create).
func (c *Client) PutPartnerPlatform(ctx context.Context, partner string, body map[string]interface{}) (*PartnerPlatform, error) {
	var out PartnerPlatform
	if err := c.do(ctx, http.MethodPut, partnerPlatformPath(partner), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeletePartnerPlatform is DELETE /v1/partner_platforms/:partner (and its stored token).
func (c *Client) DeletePartnerPlatform(ctx context.Context, partner string) error {
	return c.do(ctx, http.MethodDelete, partnerPlatformPath(partner), nil, nil)
}

// StorePartnerPlatformToken is POST /v1/partner_platforms/:partner/token. The
// token goes out in the request body only; the response is the platform view,
// which never contains it.
func (c *Client) StorePartnerPlatformToken(ctx context.Context, partner, token string, rotate bool, apiBaseURL string) (*PartnerPlatform, error) {
	body := map[string]interface{}{"token": token, "rotate": rotate}
	if apiBaseURL != "" {
		body["api_base_url"] = apiBaseURL
	}
	var out PartnerPlatform
	if err := c.do(ctx, http.MethodPost, partnerPlatformPath(partner)+"/token", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ───────────────────────────── Privacy orchestrator ─────────────────────────────

// PrivacyOrchestrator mirrors /v1/instance/privacy_orchestrator (a singleton).
// SigningSecret and CompletionWebhookSecret are returned once, on the write
// that created them.
type PrivacyOrchestrator struct {
	Configured       bool    `json:"configured"`
	Enabled          bool    `json:"enabled"`
	Provider         *string `json:"provider"`
	Name             *string `json:"name"`
	CaseURL          *string `json:"case_url"`
	CancelURL        *string `json:"cancel_url"`
	AuthHeader       string  `json:"auth_header"`
	AuthScheme       string  `json:"auth_scheme"`
	APIKeySet        bool    `json:"api_key_set"`
	SigningSecretSet bool    `json:"signing_secret_set"`
	UsePartnerLink   bool    `json:"use_partner_link"`
	PartnerLinkID    *string `json:"partner_link_id"`
	SiteID           *string `json:"site_id"`
	// CancelDeletionOnSignIn is the instance setting
	// compliance.retention.cancelDeletionOnSignIn.
	CancelDeletionOnSignIn *bool `json:"cancel_deletion_on_sign_in"`
	OpenClosureCases       bool  `json:"open_closure_cases"`
	CompletionWebhook      *struct {
		WebhookEndpointID string `json:"webhook_endpoint_id"`
		URL               string `json:"url"`
	} `json:"completion_webhook"`
	LastCaseAt              *int64  `json:"last_case_at"`
	LastError               *string `json:"last_error"`
	SigningSecret           *string `json:"signing_secret"`
	CompletionWebhookSecret *string `json:"completion_webhook_secret"`
}

func (c *Client) GetPrivacyOrchestrator(ctx context.Context) (*PrivacyOrchestrator, error) {
	var out PrivacyOrchestrator
	if err := c.do(ctx, http.MethodGet, "/v1/instance/privacy_orchestrator", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PutPrivacyOrchestrator is PUT /v1/instance/privacy_orchestrator. The body is
// built by the caller as a map: a key left out is kept, a key set to nil is sent
// as JSON null (clears a nullable setting).
func (c *Client) PutPrivacyOrchestrator(ctx context.Context, body map[string]interface{}) (*PrivacyOrchestrator, error) {
	var out PrivacyOrchestrator
	if err := c.do(ctx, http.MethodPut, "/v1/instance/privacy_orchestrator", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeletePrivacyOrchestrator(ctx context.Context) error {
	return c.do(ctx, http.MethodDelete, "/v1/instance/privacy_orchestrator", nil, nil)
}

// ───────────────────────────── Referral settings ─────────────────────────────

// ReferralSettings mirrors GET /v1/instance/referrals (a singleton).
type ReferralSettings struct {
	Enabled                 bool     `json:"enabled"`
	Provider                string   `json:"provider"`
	EndpointURL             *string  `json:"endpoint_url"`
	ProgramID               *string  `json:"program_id"`
	ConnectionID            *string  `json:"connection_id"`
	SigningSecretSet        bool     `json:"signing_secret_set"`
	APIKeySet               bool     `json:"api_key_set"`
	URLParams               []string `json:"url_params"`
	ContextCookies          []string `json:"context_cookies"`
	ConsentCategory         *string  `json:"consent_category"`
	ConsentFallbackCategory *string  `json:"consent_fallback_category"`
	AutoEnrollUsers         bool     `json:"auto_enroll_users"`
	LinkTemplate            *string  `json:"link_template"`
	TimeoutMs               *int64   `json:"timeout_ms"`
	CodeLookupURL           *string  `json:"code_lookup_url"`
	SignupCodeField         bool     `json:"signup_code_field"`
	BrowserSources          []string `json:"browser_sources"`
	ConsentProvider         *string  `json:"consent_provider"`
	ForwardLifecycle        bool     `json:"forward_lifecycle"`
	MirrorToPublicMetadata  bool     `json:"mirror_to_public_metadata"`
	IncludeReferralClaims   bool     `json:"include_referral_claims"`
	Security                struct {
		ChallengeReferredSignups    string `json:"challenge_referred_signups"`
		OnHighRisk                  string `json:"on_high_risk"`
		HoldUntilEmailVerified      bool   `json:"hold_until_email_verified"`
		MaxReferredSignupsPerIPHour int64  `json:"max_referred_signups_per_ip_hour"`
	} `json:"security"`
	Missing   []string `json:"missing"`
	UpdatedAt *int64   `json:"updated_at"`
}

// ReferralSecurity is the nested security patch.
type ReferralSecurity struct {
	ChallengeReferredSignups    *string `json:"challenge_referred_signups,omitempty"`
	OnHighRisk                  *string `json:"on_high_risk,omitempty"`
	HoldUntilEmailVerified      *bool   `json:"hold_until_email_verified,omitempty"`
	MaxReferredSignupsPerIPHour *int64  `json:"max_referred_signups_per_ip_hour,omitempty"`
}

// ReferralSettingsWrite is the PATCH body: the keys to send. A key whose value
// is nil is sent as JSON null (clearing a nullable setting); absent keys are
// left as they are.
type ReferralSettingsWrite struct {
	Fields map[string]interface{}
}

// MarshalJSON writes the field map as-is (a key set to nil is sent as null).
func (w ReferralSettingsWrite) MarshalJSON() ([]byte, error) {
	return marshalMap(w.Fields)
}

func (c *Client) GetReferralSettings(ctx context.Context) (*ReferralSettings, error) {
	var out ReferralSettings
	if err := c.do(ctx, http.MethodGet, "/v1/instance/referrals", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) PatchReferralSettings(ctx context.Context, body ReferralSettingsWrite) (*ReferralSettings, error) {
	var out ReferralSettings
	if err := c.do(ctx, http.MethodPatch, "/v1/instance/referrals", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func marshalMap(m map[string]interface{}) ([]byte, error) {
	if m == nil {
		m = map[string]interface{}{}
	}
	return json.Marshal(m)
}

// RotateStatusFeed mints a new private RSS/Atom feed token (the old URLs stop
// working). POST /v1/instance/status_subscription/rotate_feed.
func (c *Client) RotateStatusFeed(ctx context.Context) (*StatusSubscription, error) {
	var out StatusSubscription
	if err := c.do(ctx, http.MethodPost, "/v1/instance/status_subscription/rotate_feed", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
