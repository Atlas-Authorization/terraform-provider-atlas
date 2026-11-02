package provider

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &oauthClientResource{}
	_ resource.ResourceWithImportState    = &oauthClientResource{}
	_ resource.ResourceWithConfigure      = &oauthClientResource{}
	_ resource.ResourceWithValidateConfig = &oauthClientResource{}
	_ resource.ResourceWithModifyPlan     = &oauthClientResource{}
)

// NewOAuthClientResource is the factory registered with the provider.
func NewOAuthClientResource() resource.Resource { return &oauthClientResource{} }

type oauthClientResource struct {
	client *client.Client
}

// oauthClientModel is the Terraform state/plan shape for atlas_oauth_client.
type oauthClientModel struct {
	ID                      types.String `tfsdk:"id"`
	ClientID                types.String `tfsdk:"client_id"`
	Name                    types.String `tfsdk:"name"`
	RedirectURIs            types.List   `tfsdk:"redirect_uris"`
	AllowedScopes           types.List   `tfsdk:"allowed_scopes"`
	GrantTypes              types.List   `tfsdk:"grant_types"`
	TokenEndpointAuthMethod types.String `tfsdk:"token_endpoint_auth_method"`
	LogoURI                 types.String `tfsdk:"logo_uri"`
	FirstParty              types.Bool   `tfsdk:"first_party"`
	SkipLogoutConfirmation  types.Bool   `tfsdk:"skip_logout_confirmation"`
	OIDCMetadata            types.String `tfsdk:"oidc_metadata"`
	SecretPrefix            types.String `tfsdk:"secret_prefix"`
	IsPublic                types.Bool   `tfsdk:"is_public"`
	ClientSecret            types.String `tfsdk:"client_secret"`
	CreatedAt               types.Int64  `tfsdk:"created_at"`
	UpdatedAt               types.Int64  `tfsdk:"updated_at"`
}

func (r *oauthClientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_client"
}

func (r *oauthClientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *oauthClientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A \"Sign in with Atlas\" relying party (OAuth/OIDC client). A confidential client's " +
			"secret is revealed once, on create, and stored (sensitive) in state; a public (PKCE) client has none.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The public OAuth client_id used at the authorize/token endpoints.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable name shown on the consent screen.",
				Required:            true,
			},
			"redirect_uris": schema.ListAttribute{
				MarkdownDescription: "Allowed redirect URIs. Absolute https (or http://localhost for development).",
				Required:            true,
				ElementType:         types.StringType,
			},
			"allowed_scopes": schema.ListAttribute{
				MarkdownDescription: "OIDC scopes the client may request. Defaults to [openid, profile, email]; openid is required.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"grant_types": schema.ListAttribute{
				MarkdownDescription: "Allowed grant types. Defaults to [authorization_code, refresh_token].",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"token_endpoint_auth_method": schema.StringAttribute{
				MarkdownDescription: "One of client_secret_basic, client_secret_post or none (public/PKCE). Defaults to client_secret_basic.",
				Optional:            true,
				Computed:            true,
			},
			"logo_uri": schema.StringAttribute{
				MarkdownDescription: "Optional logo URL shown on the consent screen.",
				Optional:            true,
			},
			"first_party": schema.BoolAttribute{
				MarkdownDescription: "When true the client is first-party and may skip the consent screen. Defaults to false.",
				Optional:            true,
				Computed:            true,
			},
			"skip_logout_confirmation": schema.BoolAttribute{
				MarkdownDescription: "When true, an RP-initiated logout (`/oauth2/logout`) from this client with an exactly " +
					"registered `post_logout_redirect_uri` ends the session without the \"Sign out?\" page and redirects " +
					"straight back. **First-party clients only** (`first_party = true`); Atlas refuses it on a third-party " +
					"client. Defaults to false.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"oidc_metadata": schema.StringAttribute{
				MarkdownDescription: "OpenID metadata for the client as a JSON object (use `jsonencode`). Keys: `jwks`, `jwks_uri`, " +
					"`tls_client_auth_subject_dn`, `id_token_signed_response_alg`, `userinfo_signed_response_alg`, " +
					"`authorization_signed_response_alg`, `introspection_signed_response_alg`, the matching `*_encrypted_response_alg` / " +
					"`*_encrypted_response_enc`, `request_object_signing_alg`, `request_object_encryption_alg` / `_enc`, " +
					"`require_signed_request_object`, `request_uris`, `response_types`, `backchannel_token_delivery_mode` " +
					"(`poll`, `ping`, `push`), `backchannel_client_notification_endpoint`, `backchannel_user_code_parameter`, " +
					"`backchannel_authentication_request_signing_alg`, `org_selection` (`none`, `ask`, `require`), `min_app_version`, " +
					"`native_sso_group`, `wimse_workload_id`, `tls_client_auth_san_*`, `dpop_bound_access_tokens` and " +
					"`txn_token_unsigned_subject`. Set a key to `null` to clear it. Only the keys you list are managed; others set " +
					"elsewhere are left alone. Not read back after an import.",
				Optional: true,
			},
			"secret_prefix": schema.StringAttribute{
				MarkdownDescription: "Non-secret prefix of the live client secret, for recognising which secret is deployed.",
				Computed:            true,
			},
			"is_public": schema.BoolAttribute{
				MarkdownDescription: "True when the client authenticates with PKCE and holds no secret.",
				Computed:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The confidential client secret, revealed only on create. Empty for a public client. " +
					"Use the rotate flow out of band; a read never returns it.",
				Computed:      true,
				Sensitive:     true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch milliseconds).",
				Computed:            true,
			},
			"updated_at": schema.Int64Attribute{
				MarkdownDescription: "Last update time (epoch milliseconds).",
				Computed:            true,
			},
		},
	}
}

func (r *oauthClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan oauthClientModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := client.OAuthClientWrite{
		Name:         client.Ptr(plan.Name.ValueString()),
		RedirectURIs: listToStringSlice(ctx, plan.RedirectURIs, &resp.Diagnostics),
	}
	if scopes := listToStringSlice(ctx, plan.AllowedScopes, &resp.Diagnostics); scopes != nil {
		body.AllowedScopes = scopes
	}
	if grants := listToStringSlice(ctx, plan.GrantTypes, &resp.Diagnostics); grants != nil {
		body.GrantTypes = grants
	}
	if m := optionalString(plan.TokenEndpointAuthMethod); m != nil {
		body.TokenEndpointAuthMethod = m
	}
	if l := optionalString(plan.LogoURI); l != nil {
		body.LogoURL = l
	}
	if !plan.FirstParty.IsNull() && !plan.FirstParty.IsUnknown() {
		body.FirstParty = client.Ptr(plan.FirstParty.ValueBool())
	}
	body.SkipLogoutConfirmation = optionalBool(plan.SkipLogoutConfirmation)
	body.Extra = oidcMetadataBody(plan.OIDCMetadata, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateOAuthClient(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create OAuth client", err.Error())
		return
	}

	r.mapToState(ctx, created, &plan, &resp.Diagnostics, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oauthClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state oauthClientModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fetched, err := r.client.GetOAuthClient(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read OAuth client", err.Error())
		return
	}

	// A read never returns the secret; keep whatever create stored in state.
	r.mapToState(ctx, fetched, &state, &resp.Diagnostics, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *oauthClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state oauthClientModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := client.OAuthClientWrite{
		Name:         client.Ptr(plan.Name.ValueString()),
		RedirectURIs: listToStringSlice(ctx, plan.RedirectURIs, &resp.Diagnostics),
	}
	if scopes := listToStringSlice(ctx, plan.AllowedScopes, &resp.Diagnostics); scopes != nil {
		body.AllowedScopes = scopes
	}
	if m := optionalString(plan.TokenEndpointAuthMethod); m != nil {
		body.TokenEndpointAuthMethod = m
	}
	// An explicit null logo_uri clears it; the API treats "" as clear.
	if plan.LogoURI.IsNull() {
		body.LogoURL = client.Ptr("")
	} else {
		body.LogoURL = optionalString(plan.LogoURI)
	}
	if !plan.FirstParty.IsNull() && !plan.FirstParty.IsUnknown() {
		body.FirstParty = client.Ptr(plan.FirstParty.ValueBool())
	}
	body.SkipLogoutConfirmation = optionalBool(plan.SkipLogoutConfirmation)
	body.Extra = oidcMetadataBody(plan.OIDCMetadata, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateOAuthClient(ctx, plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update OAuth client", err.Error())
		return
	}

	// Carry the create-time secret forward; update does not re-issue it.
	plan.ClientSecret = state.ClientSecret
	r.mapToState(ctx, updated, &plan, &resp.Diagnostics, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oauthClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state oauthClientModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOAuthClient(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete OAuth client", err.Error())
	}
}

// ValidateConfig catches skip_logout_confirmation on a client configured as
// third-party (first_party = false) at plan time; Atlas would answer 422.
func (r *oauthClientResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var skip, firstParty types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("skip_logout_confirmation"), &skip)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("first_party"), &firstParty)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if skip.IsUnknown() || !skip.ValueBool() || firstParty.IsUnknown() || firstParty.IsNull() || firstParty.ValueBool() {
		return
	}
	addSkipLogoutConfirmationError(&resp.Diagnostics)
}

// ModifyPlan covers first_party left unset: a new client defaults to
// third-party, and an existing one keeps its current (state) value. Either
// way, a third-party client cannot skip the logout confirmation.
func (r *oauthClientResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return // destroy
	}
	var skip, configFirstParty types.Bool
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("skip_logout_confirmation"), &skip)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("first_party"), &configFirstParty)...)
	if resp.Diagnostics.HasError() || skip.IsUnknown() || !skip.ValueBool() || !configFirstParty.IsNull() {
		return
	}
	firstParty := false
	if !req.State.Raw.IsNull() {
		var prior types.Bool
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("first_party"), &prior)...)
		firstParty = prior.ValueBool()
	}
	if !firstParty {
		addSkipLogoutConfirmationError(&resp.Diagnostics)
	}
}

func addSkipLogoutConfirmationError(diags *diag.Diagnostics) {
	diags.AddAttributeError(path.Root("skip_logout_confirmation"),
		"skip_logout_confirmation needs a first-party client",
		"Only first-party clients may skip the logout confirmation: Atlas refuses it (422) on a third-party "+
			"client. Set first_party = true on this client, or remove skip_logout_confirmation.")
}

func (r *oauthClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapToState projects an API object onto the Terraform model. withSecret is true
// only right after a create/rotate, when the one-time secret is present.
func (r *oauthClientResource) mapToState(ctx context.Context, c *client.OAuthClient, m *oauthClientModel, diags *diag.Diagnostics, withSecret bool) {
	m.ID = types.StringValue(c.ID)
	m.ClientID = types.StringValue(c.ClientID)
	m.Name = types.StringValue(c.Name)
	m.RedirectURIs = stringSliceToList(ctx, c.RedirectURIs, diags)
	m.AllowedScopes = stringSliceToList(ctx, c.AllowedScopes, diags)
	m.GrantTypes = stringSliceToList(ctx, c.GrantTypes, diags)
	m.TokenEndpointAuthMethod = types.StringValue(c.TokenEndpointAuthMethod)
	m.LogoURI = stringPtrToValue(c.LogoURL)
	m.FirstParty = types.BoolValue(c.FirstParty)
	m.SkipLogoutConfirmation = types.BoolValue(c.SkipLogoutConfirmation)
	m.SecretPrefix = stringPtrToValue(c.SecretPrefix)
	m.OIDCMetadata = oidcMetadataFromAPI(m.OIDCMetadata, c.Raw)
	m.IsPublic = types.BoolValue(c.IsPublic)
	m.CreatedAt = types.Int64Value(c.CreatedAt)
	m.UpdatedAt = types.Int64Value(c.UpdatedAt)
	if withSecret {
		if c.ClientSecret != nil {
			m.ClientSecret = types.StringValue(*c.ClientSecret)
		} else {
			// A public client has none — record the empty string rather than
			// leaving the computed attribute unknown after apply.
			m.ClientSecret = types.StringValue("")
		}
	}
}

// oidcMetadataBody parses the oidc_metadata JSON object into top-level request keys.
func oidcMetadataBody(v types.String, diags *diag.Diagnostics) map[string]json.RawMessage {
	if v.IsNull() || v.IsUnknown() || strings.TrimSpace(v.ValueString()) == "" {
		return nil
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal([]byte(v.ValueString()), &out); err != nil {
		diags.AddAttributeError(path.Root("oidc_metadata"), "Invalid oidc_metadata", "oidc_metadata must be a JSON object: "+err.Error())
		return nil
	}
	return out
}

// oidcMetadataFromAPI returns the configured document while the server still holds
// every key in it, and otherwise the server's values for those keys (drift). Keys
// that were never configured are not tracked, and an unset attribute stays unset.
func oidcMetadataFromAPI(prior types.String, raw map[string]json.RawMessage) types.String {
	if prior.IsNull() || prior.IsUnknown() || raw == nil {
		return prior
	}
	var cfg map[string]any
	if json.Unmarshal([]byte(prior.ValueString()), &cfg) != nil {
		return prior
	}
	got := map[string]any{}
	for k := range cfg {
		if v, ok := raw[k]; ok {
			var x any
			if json.Unmarshal(v, &x) == nil {
				got[k] = x
			}
		}
	}
	if jsonCovers(cfg, got) {
		return prior
	}
	out, err := json.Marshal(got)
	if err != nil {
		return prior
	}
	return types.StringValue(string(out))
}
