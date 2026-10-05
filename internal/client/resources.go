package client

import (
	"context"
	"net/http"
	"net/url"
)

// listEnvelope is the shape of every /v1 list response: { data: [...] }.
type listEnvelope[T any] struct {
	Data    []T  `json:"data"`
	HasMore bool `json:"has_more"`
}

// ───────────────────────────── OAuth clients ─────────────────────────────

// OAuthClient mirrors the bapi-oauth-clients projection. ClientSecret is only
// populated on Create / RotateSecret; a read never returns it.
type OAuthClient struct {
	ID                      string   `json:"id"`
	ClientID                string   `json:"client_id"`
	Name                    string   `json:"name"`
	LogoURL                 *string  `json:"logo_url"`
	RedirectURIs            []string `json:"redirect_uris"`
	AllowedScopes           []string `json:"allowed_scopes"`
	GrantTypes              []string `json:"grant_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	SecretPrefix            *string  `json:"secret_prefix"`
	IsPublic                bool     `json:"is_public"`
	FirstParty              bool     `json:"first_party"`
	CreatedAt               int64    `json:"created_at"`
	UpdatedAt               int64    `json:"updated_at"`
	// Revealed exactly once on create/rotate; absent for a public client.
	ClientSecret *string `json:"client_secret"`
}

// OAuthClientWrite is the create/update payload. Pointer fields let an update
// send only the keys that changed (a PATCH), matching the API's merge semantics.
type OAuthClientWrite struct {
	Name                    *string  `json:"name,omitempty"`
	RedirectURIs            []string `json:"redirect_uris,omitempty"`
	AllowedScopes           []string `json:"allowed_scopes,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	TokenEndpointAuthMethod *string  `json:"token_endpoint_auth_method,omitempty"`
	LogoURL                 *string  `json:"logo_url,omitempty"`
	FirstParty              *bool    `json:"first_party,omitempty"`
}

func (c *Client) CreateOAuthClient(ctx context.Context, body OAuthClientWrite) (*OAuthClient, error) {
	var out OAuthClient
	if err := c.do(ctx, http.MethodPost, "/v1/oauth_clients", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetOAuthClient(ctx context.Context, id string) (*OAuthClient, error) {
	var out OAuthClient
	if err := c.do(ctx, http.MethodGet, "/v1/oauth_clients/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateOAuthClient(ctx context.Context, id string, body OAuthClientWrite) (*OAuthClient, error) {
	var out OAuthClient
	if err := c.do(ctx, http.MethodPatch, "/v1/oauth_clients/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteOAuthClient(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/oauth_clients/"+url.PathEscape(id), nil, nil)
}

// ─────────────────────────── SSO connections ───────────────────────────

// ClaimRoleMapping maps an IdP claim value onto an Atlas role key.
type ClaimRoleMapping struct {
	Claim   string `json:"claim"`
	Value   string `json:"value"`
	RoleKey string `json:"roleKey"`
}

// SsoConnection mirrors the bapi-sso-connections projection. Secrets are never
// returned; only has_* booleans report their presence.
type SsoConnection struct {
	ID                     string             `json:"id"`
	OrganizationID         *string            `json:"organization_id"`
	Type                   string             `json:"type"`
	Status                 string             `json:"status"`
	OidcIssuer             *string            `json:"oidc_issuer"`
	OidcClientID           *string            `json:"oidc_client_id"`
	HasSecret              bool               `json:"has_secret"`
	SamlIdpEntityID        *string            `json:"saml_idp_entity_id"`
	SamlIdpSsoURL          *string            `json:"saml_idp_sso_url"`
	SamlSpEntityID         *string            `json:"saml_sp_entity_id"`
	HasSamlCertificate     bool               `json:"has_saml_certificate"`
	SamlAllowIdpInitiated  bool               `json:"saml_allow_idp_initiated"`
	SamlSignAuthnRequests  bool               `json:"saml_sign_authn_requests"`
	SamlWantResponseSigned bool               `json:"saml_want_response_signed"`
	HasDiscourseSecret     bool               `json:"has_discourse_secret"`
	DiscourseProviderURL   *string            `json:"discourse_provider_url"`
	AllowedDomains         []string           `json:"allowed_domains"`
	ClaimRoleMappings      []ClaimRoleMapping `json:"claim_role_mappings"`
	DefaultRoleID          *string            `json:"default_role_id"`
	CreatedAt              int64              `json:"created_at"`
	UpdatedAt              int64              `json:"updated_at"`
}

// SsoConnectionWrite is the create/update payload. Pointer/omitempty fields mean
// an update only touches supplied keys.
type SsoConnectionWrite struct {
	OrganizationID         *string            `json:"organization_id,omitempty"`
	Type                   *string            `json:"type,omitempty"`
	Status                 *string            `json:"status,omitempty"`
	OidcIssuer             *string            `json:"oidc_issuer,omitempty"`
	OidcClientID           *string            `json:"oidc_client_id,omitempty"`
	OidcClientSecret       *string            `json:"oidc_client_secret,omitempty"`
	SamlIdpEntityID        *string            `json:"saml_idp_entity_id,omitempty"`
	SamlIdpSsoURL          *string            `json:"saml_idp_sso_url,omitempty"`
	SamlIdpCertificate     *string            `json:"saml_idp_certificate,omitempty"`
	SamlSpEntityID         *string            `json:"saml_sp_entity_id,omitempty"`
	SamlAllowIdpInitiated  *bool              `json:"saml_allow_idp_initiated,omitempty"`
	SamlSignAuthnRequests  *bool              `json:"saml_sign_authn_requests,omitempty"`
	SamlWantResponseSigned *bool              `json:"saml_want_response_signed,omitempty"`
	DiscourseSecret        *string            `json:"discourse_secret,omitempty"`
	DiscourseProviderURL   *string            `json:"discourse_provider_url,omitempty"`
	AllowedDomains         []string           `json:"allowed_domains,omitempty"`
	ClaimRoleMappings      []ClaimRoleMapping `json:"claim_role_mappings,omitempty"`
	DefaultRoleID          *string            `json:"default_role_id,omitempty"`
}

func (c *Client) CreateSsoConnection(ctx context.Context, body SsoConnectionWrite) (*SsoConnection, error) {
	var out SsoConnection
	if err := c.do(ctx, http.MethodPost, "/v1/sso_connections", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetSsoConnection(ctx context.Context, id string) (*SsoConnection, error) {
	var out SsoConnection
	if err := c.do(ctx, http.MethodGet, "/v1/sso_connections/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateSsoConnection(ctx context.Context, id string, body SsoConnectionWrite) (*SsoConnection, error) {
	var out SsoConnection
	if err := c.do(ctx, http.MethodPatch, "/v1/sso_connections/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSsoConnection(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/sso_connections/"+url.PathEscape(id), nil, nil)
}

// ─────────────────────────── Resource servers ───────────────────────────

// ResourceServerScope is one {value, description} entry.
type ResourceServerScope struct {
	Value       string  `json:"value"`
	Description *string `json:"description,omitempty"`
}

// ResourceServer mirrors the bapi-resource-servers projection.
type ResourceServer struct {
	ID              string                `json:"id"`
	Identifier      string                `json:"identifier"`
	Name            string                `json:"name"`
	Scopes          []ResourceServerScope `json:"scopes"`
	TokenTTLSeconds int64                 `json:"token_ttl_seconds"`
	SigningAlg      string                `json:"signing_alg"`
	CreatedAt       int64                 `json:"created_at"`
	UpdatedAt       int64                 `json:"updated_at"`
}

// ResourceServerWrite is the create/update payload.
type ResourceServerWrite struct {
	Identifier      *string               `json:"identifier,omitempty"`
	Name            *string               `json:"name,omitempty"`
	Scopes          []ResourceServerScope `json:"scopes,omitempty"`
	TokenTTLSeconds *int64                `json:"token_ttl_seconds,omitempty"`
	SigningAlg      *string               `json:"signing_alg,omitempty"`
}

func (c *Client) CreateResourceServer(ctx context.Context, body ResourceServerWrite) (*ResourceServer, error) {
	var out ResourceServer
	if err := c.do(ctx, http.MethodPost, "/v1/resource_servers", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetResourceServer(ctx context.Context, id string) (*ResourceServer, error) {
	var out ResourceServer
	if err := c.do(ctx, http.MethodGet, "/v1/resource_servers/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateResourceServer(ctx context.Context, id string, body ResourceServerWrite) (*ResourceServer, error) {
	var out ResourceServer
	if err := c.do(ctx, http.MethodPatch, "/v1/resource_servers/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteResourceServer(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/resource_servers/"+url.PathEscape(id), nil, nil)
}

// ─────────────────────────── JWT templates ───────────────────────────

// JwtTemplate is keyed by Name; claims is a flat map of claim → value template.
type JwtTemplate struct {
	Name   string            `json:"name"`
	Claims map[string]string `json:"claims"`
}

type jwtTemplateCreate struct {
	Name   string            `json:"name"`
	Claims map[string]string `json:"claims"`
}

type jwtTemplateUpdate struct {
	Claims map[string]string `json:"claims"`
}

func (c *Client) CreateJwtTemplate(ctx context.Context, name string, claims map[string]string) (*JwtTemplate, error) {
	var out JwtTemplate
	if err := c.do(ctx, http.MethodPost, "/v1/jwt_templates", jwtTemplateCreate{Name: name, Claims: claims}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetJwtTemplate(ctx context.Context, name string) (*JwtTemplate, error) {
	var out JwtTemplate
	if err := c.do(ctx, http.MethodGet, "/v1/jwt_templates/"+url.PathEscape(name), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateJwtTemplate(ctx context.Context, name string, claims map[string]string) (*JwtTemplate, error) {
	var out JwtTemplate
	if err := c.do(ctx, http.MethodPatch, "/v1/jwt_templates/"+url.PathEscape(name), jwtTemplateUpdate{Claims: claims}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteJwtTemplate(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/v1/jwt_templates/"+url.PathEscape(name), nil, nil)
}

// ─────────────────────────── Webhook endpoints ───────────────────────────

// WebhookEndpoint mirrors the bapi-platform projection. The signing Secret is
// revealed exactly once, on create.
type WebhookEndpoint struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	EnabledEvents []string `json:"enabled_events"`
	Active        bool     `json:"active"`
	DisabledAt    *int64   `json:"disabled_at"`
	CreatedAt     int64    `json:"created_at"`
	Secret        *string  `json:"secret"`
}

type webhookEndpointCreate struct {
	URL           string   `json:"url"`
	EnabledEvents []string `json:"enabled_events,omitempty"`
}

func (c *Client) CreateWebhookEndpoint(ctx context.Context, url string, events []string) (*WebhookEndpoint, error) {
	var out WebhookEndpoint
	if err := c.do(ctx, http.MethodPost, "/v1/webhook_endpoints", webhookEndpointCreate{URL: url, EnabledEvents: events}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWebhookEndpoint reads by id. There is no single-GET route, so it lists and
// filters — a disabled ("deleted") endpoint is treated as absent.
func (c *Client) GetWebhookEndpoint(ctx context.Context, id string) (*WebhookEndpoint, error) {
	var out listEnvelope[WebhookEndpoint]
	if err := c.do(ctx, http.MethodGet, "/v1/webhook_endpoints", nil, &out); err != nil {
		return nil, err
	}
	for i := range out.Data {
		if out.Data[i].ID == id && out.Data[i].Active {
			return &out.Data[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Code: "NOT_FOUND", Message: "Unknown webhook endpoint."}}}
}

// WebhookEndpointUpdate is the PATCH /v1/webhook_endpoints/:id body. Every field
// is a pointer/slice so only the fields the caller supplies are serialized (a nil
// field is omitted and left untouched server-side). The signing secret is NOT a
// field here and is never re-issued by an update.
type WebhookEndpointUpdate struct {
	URL           *string  `json:"url,omitempty"`
	EnabledEvents []string `json:"enabled_events,omitempty"`
	Active        *bool    `json:"active,omitempty"`
}

// UpdateWebhookEndpoint patches an endpoint IN PLACE (url, enabled_events and/or
// active) without rotating the signing secret. PATCH /v1/webhook_endpoints/:id —
// the 200 body is the updated projection with NO secret field (the secret is
// preserved, so the caller must carry it forward from prior state).
func (c *Client) UpdateWebhookEndpoint(ctx context.Context, id string, patch WebhookEndpointUpdate) (*WebhookEndpoint, error) {
	var out WebhookEndpoint
	if err := c.do(ctx, http.MethodPatch, "/v1/webhook_endpoints/"+url.PathEscape(id), patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteWebhookEndpoint(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/webhook_endpoints/"+url.PathEscape(id), nil, nil)
}

// ─────────────────────────── Roles ───────────────────────────

// Role mirrors the bapi-roles projection.
type Role struct {
	ID          string   `json:"id"`
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	IsSystem    bool     `json:"is_system"`
	Permissions []string `json:"permissions"`
	MemberCount int64    `json:"member_count"`
	CreatedAt   int64    `json:"created_at"`
}

type roleCreate struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type roleUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (c *Client) CreateRole(ctx context.Context, key, name string, description *string, permissions []string) (*Role, error) {
	var out Role
	body := roleCreate{Key: key, Name: name, Description: description, Permissions: permissions}
	if err := c.do(ctx, http.MethodPost, "/v1/roles", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRole reads by id. Roles have no single-GET route, so it lists and filters.
func (c *Client) GetRole(ctx context.Context, id string) (*Role, error) {
	var out listEnvelope[Role]
	if err := c.do(ctx, http.MethodGet, "/v1/roles", nil, &out); err != nil {
		return nil, err
	}
	for i := range out.Data {
		if out.Data[i].ID == id {
			return &out.Data[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Code: "NOT_FOUND", Message: "Unknown role."}}}
}

// UpdateRoleLabel PATCHes the name/description (the key is immutable).
func (c *Client) UpdateRoleLabel(ctx context.Context, id string, name, description *string) (*Role, error) {
	var out Role
	if err := c.do(ctx, http.MethodPatch, "/v1/roles/"+url.PathEscape(id), roleUpdate{Name: name, Description: description}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetRolePermissions PUTs the full permission set. Undefined permissions are
// silently dropped by the API; the applied set is returned in `permissions`.
func (c *Client) SetRolePermissions(ctx context.Context, id string, permissions []string) ([]string, error) {
	if permissions == nil {
		permissions = []string{}
	}
	var out struct {
		Permissions []string `json:"permissions"`
	}
	body := map[string][]string{"permissions": permissions}
	if err := c.do(ctx, http.MethodPut, "/v1/roles/"+url.PathEscape(id)+"/permissions", body, &out); err != nil {
		return nil, err
	}
	return out.Permissions, nil
}

func (c *Client) DeleteRole(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/roles/"+url.PathEscape(id), nil, nil)
}

// ─────────────────────────── Organizations ───────────────────────────

// Organization mirrors the bapi-organizations projection.
type Organization struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	Slug                  string  `json:"slug"`
	ImageURL              *string `json:"image_url"`
	MaxAllowedMemberships *int64  `json:"max_allowed_memberships"`
	CreatedBy             *string `json:"created_by"`
	CreatedAt             int64   `json:"created_at"`
	UpdatedAt             int64   `json:"updated_at"`
}

type organizationCreate struct {
	Name                  string `json:"name"`
	Slug                  string `json:"slug"`
	CreatedBy             string `json:"created_by"`
	MaxAllowedMemberships *int64 `json:"max_allowed_memberships,omitempty"`
}

type organizationUpdate struct {
	Name                  *string `json:"name,omitempty"`
	Slug                  *string `json:"slug,omitempty"`
	MaxAllowedMemberships *int64  `json:"max_allowed_memberships,omitempty"`
}

func (c *Client) CreateOrganization(ctx context.Context, name, slug, createdBy string, maxMemberships *int64) (*Organization, error) {
	var out Organization
	body := organizationCreate{Name: name, Slug: slug, CreatedBy: createdBy, MaxAllowedMemberships: maxMemberships}
	if err := c.do(ctx, http.MethodPost, "/v1/organizations", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetOrganization(ctx context.Context, id string) (*Organization, error) {
	var out Organization
	if err := c.do(ctx, http.MethodGet, "/v1/organizations/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateOrganization(ctx context.Context, id string, body organizationUpdate) (*Organization, error) {
	var out Organization
	if err := c.do(ctx, http.MethodPatch, "/v1/organizations/"+url.PathEscape(id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// OrganizationUpdate is the exported update payload for the organization resource.
type OrganizationUpdate = organizationUpdate

func (c *Client) DeleteOrganization(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/organizations/"+url.PathEscape(id), nil, nil)
}

// ─────────────────────────── Custom domains ───────────────────────────

// Domain mirrors the bapi-domains projection (instance custom domains).
type Domain struct {
	ID                 string  `json:"id"`
	Role               string  `json:"role"`
	Host               string  `json:"host"`
	Status             string  `json:"status"`
	Action             *string `json:"action"`
	Live               bool    `json:"live"`
	CnameTarget        string  `json:"cname_target"`
	LastCheckedAt      *int64  `json:"last_checked_at"`
	LastObservedTarget *string `json:"last_observed_target"`
	FailureReason      *string `json:"failure_reason"`
	CertificateExpires *int64  `json:"certificate_expires_at"`
	CookieDomain       *string `json:"cookie_domain"`
}

type domainCreate struct {
	Role string `json:"role"`
	Host string `json:"host"`
}

func (c *Client) CreateDomain(ctx context.Context, role, host string) (*Domain, error) {
	var out Domain
	if err := c.do(ctx, http.MethodPost, "/v1/domains", domainCreate{Role: role, Host: host}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDomain reads by id. There is no single-GET route, so it lists and filters.
func (c *Client) GetDomain(ctx context.Context, id string) (*Domain, error) {
	var out listEnvelope[Domain]
	if err := c.do(ctx, http.MethodGet, "/v1/domains", nil, &out); err != nil {
		return nil, err
	}
	for i := range out.Data {
		if out.Data[i].ID == id {
			return &out.Data[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Code: "NOT_FOUND", Message: "Unknown domain."}}}
}

func (c *Client) DeleteDomain(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/domains/"+url.PathEscape(id), nil, nil)
}

// ─────────────────────────── OAuth providers (social login) ───────────────────────────

// OAuthProvider mirrors the bapi-oauth-providers projection (the §11.1 secret-key
// mirror of the dashboard "OAuth providers" screen). It is the whole catalog
// entry plus whatever this instance has configured. The client SECRET is
// write-only and never returned; only HasSecret reports its presence. The
// redirect URI is DERIVED from the instance host, never accepted from a caller.
type OAuthProvider struct {
	Provider      string   `json:"provider"`
	DisplayName   string   `json:"display_name"`
	Category      string   `json:"category"`
	Tier          string   `json:"tier"`
	RedirectURI   string   `json:"redirect_uri"`
	DefaultScopes []string `json:"default_scopes"`
	Configured    bool     `json:"configured"`
	ClientID      *string  `json:"client_id"`
	HasSecret     bool     `json:"has_secret"`
	Enabled       bool     `json:"enabled"`
	AllowSignIn   bool     `json:"allow_sign_in"`
	AllowSignUp   bool     `json:"allow_sign_up"`
	Scopes        []string `json:"scopes"`
	UpdatedAt     *int64   `json:"updated_at"`
}

// OAuthProviderWrite is the idempotent PUT payload. Values is keyed by the
// provider's own credential-field keys (typically client_id + client_secret);
// the API distributes them to the id column, the encrypted secret column and the
// config blob per the provider's real credential shape. A nil Scopes lets the
// API fall back to the provider's default scopes.
type OAuthProviderWrite struct {
	Values   map[string]string `json:"values,omitempty"`
	Settings map[string]string `json:"settings,omitempty"`
	Scopes   []string          `json:"scopes,omitempty"`
}

// GetOAuthProvider reads one provider by catalog key.
// GET /v1/oauth_providers/:provider — 404 only for an UNKNOWN key; a known but
// unconfigured provider is returned with Configured=false.
func (c *Client) GetOAuthProvider(ctx context.Context, provider string) (*OAuthProvider, error) {
	var out OAuthProvider
	if err := c.do(ctx, http.MethodGet, "/v1/oauth_providers/"+url.PathEscape(provider), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PutOAuthProvider sets a provider's credentials (+ scopes/settings). Idempotent.
// PUT /v1/oauth_providers/:provider — returns only { provider, configured }, so
// callers GET afterwards for the full projection.
func (c *Client) PutOAuthProvider(ctx context.Context, provider string, body OAuthProviderWrite) error {
	return c.do(ctx, http.MethodPut, "/v1/oauth_providers/"+url.PathEscape(provider), body, nil)
}

// SetOAuthProviderEnabled toggles a provider on or off.
// POST /v1/oauth_providers/:provider/enabled — enabling one with no usable
// credentials is refused by the API.
func (c *Client) SetOAuthProviderEnabled(ctx context.Context, provider string, enabled bool) error {
	return c.do(ctx, http.MethodPost, "/v1/oauth_providers/"+url.PathEscape(provider)+"/enabled", map[string]bool{"enabled": enabled}, nil)
}

// SetOAuthProviderScope scopes a configured provider to sign-in, sign-up or both.
// POST /v1/oauth_providers/:provider/scope — the API refuses "neither" (disable
// the provider instead) and requires it to be configured first.
func (c *Client) SetOAuthProviderScope(ctx context.Context, provider string, allowSignIn, allowSignUp bool) error {
	body := map[string]bool{"allow_sign_in": allowSignIn, "allow_sign_up": allowSignUp}
	return c.do(ctx, http.MethodPost, "/v1/oauth_providers/"+url.PathEscape(provider)+"/scope", body, nil)
}

// DeleteOAuthProvider removes a provider's credentials. Linked accounts survive.
// DELETE /v1/oauth_providers/:provider — 404 when it was not configured.
func (c *Client) DeleteOAuthProvider(ctx context.Context, provider string) error {
	return c.do(ctx, http.MethodDelete, "/v1/oauth_providers/"+url.PathEscape(provider), nil, nil)
}

// ptr is a tiny helper used by resource code to take the address of a literal.
func ptr[T any](v T) *T { return &v }

// Ptr exposes ptr for the provider package.
func Ptr[T any](v T) *T { return ptr(v) }
