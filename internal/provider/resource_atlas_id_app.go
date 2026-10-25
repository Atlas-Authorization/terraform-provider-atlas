package provider

import (
	"context"
	"encoding/json"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &atlasIDAppResource{}
	_ resource.ResourceWithImportState = &atlasIDAppResource{}
	_ resource.ResourceWithConfigure   = &atlasIDAppResource{}
	_ resource.ResourceWithModifyPlan  = &atlasIDAppResource{}
)

// NewAtlasIDAppResource is the factory registered with the provider.
func NewAtlasIDAppResource() resource.Resource { return &atlasIDAppResource{} }

type atlasIDAppResource struct {
	client *client.Client
}

type atlasIDAppModel struct {
	ID                                types.String    `tfsdk:"id"`
	ClientID                          types.String    `tfsdk:"client_id"`
	Name                              types.String    `tfsdk:"name"`
	Type                              types.String    `tfsdk:"type"`
	RedirectURIs                      types.List      `tfsdk:"redirect_uris"`
	PostLogoutRedirectURIs            types.List      `tfsdk:"post_logout_redirect_uris"`
	BackchannelLogoutURI              types.String    `tfsdk:"backchannel_logout_uri"`
	BackchannelLogoutSessionRequired  types.Bool      `tfsdk:"backchannel_logout_session_required"`
	FrontchannelLogoutURI             types.String    `tfsdk:"frontchannel_logout_uri"`
	FrontchannelLogoutSessionRequired types.Bool      `tfsdk:"frontchannel_logout_session_required"`
	AllowedScopes                     types.List      `tfsdk:"allowed_scopes"`
	LogoURL                           types.String    `tfsdk:"logo_url"`
	HomepageURL                       types.String    `tfsdk:"homepage_url"`
	Branding                          jsonObjectValue `tfsdk:"branding"`
	RotateSecretTrigger               types.String    `tfsdk:"rotate_secret_trigger"`
	ClientSecret                      types.String    `tfsdk:"client_secret"`
	SecretPrefix                      types.String    `tfsdk:"secret_prefix"`
	ManagedByInstanceID               types.String    `tfsdk:"managed_by_instance_id"`
	GrantsManagedByInstanceID         types.String    `tfsdk:"grants_managed_by_instance_id"`
	ManagedVia                        types.String    `tfsdk:"managed_via"`
	TestMode                          types.Bool      `tfsdk:"test_mode"`
	ProvisionedVerifiedDomainsOnly    types.Bool      `tfsdk:"provisioned_verified_domains_only"`
	CreatedAt                         types.Int64     `tfsdk:"created_at"`
	UpdatedAt                         types.Int64     `tfsdk:"updated_at"`
}

func (r *atlasIDAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_id_app"
}

func (r *atlasIDAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *atlasIDAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A **Sign in with Atlas** app: an OpenID Connect client on the shared Atlas issuer " +
			"(`https://atlasauth.net`), registered to your platform account. Needs the provider's `platform_key` " +
			"(`psk_...`). A confidential app's `client_secret` is returned once, on create (and on rotation), and kept " +
			"(sensitive) in state. `branding` is a JSON object compared semantically, so the server's key order, " +
			"formatting and null members never cause drift.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The OAuth client id your app sends to the Atlas issuer.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "App name shown on the sign-in and consent screens (at most 100 characters).",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "`confidential` (default; gets a client secret) or `public` (PKCE, no secret). " +
					"Cannot change after creation — changing it forces replacement.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"redirect_uris": schema.ListAttribute{
				MarkdownDescription: "Exact-match redirect URIs (https, or http on localhost).",
				Required:            true,
				ElementType:         types.StringType,
			},
			"post_logout_redirect_uris": schema.ListAttribute{
				MarkdownDescription: "Where `/oauth2/logout` may send the user after sign-out. Exact match; same URL rule " +
					"as `redirect_uris`. Defaults to none.",
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"backchannel_logout_uri": schema.StringAttribute{
				MarkdownDescription: "OIDC Back-Channel Logout: a signed `logout_token` is POSTed here when the user signs out.",
				Optional:            true,
			},
			"backchannel_logout_session_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the app needs `sid` in the logout token. Default false.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"frontchannel_logout_uri": schema.StringAttribute{
				MarkdownDescription: "OIDC Front-Channel Logout: loaded in a hidden iframe on the Atlas sign-out page.",
				Optional:            true,
			},
			"frontchannel_logout_session_required": schema.BoolAttribute{
				MarkdownDescription: "Add `sid` (and `iss`) to the front-channel logout URL. Default false.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"allowed_scopes": schema.ListAttribute{
				MarkdownDescription: "Scopes the app may request: `openid` (required), `profile`, `email`, `offline_access` " +
					"(refresh tokens). Defaults to `openid profile email`.",
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"logo_url": schema.StringAttribute{
				MarkdownDescription: "https URL of the app logo.",
				Optional:            true,
			},
			"homepage_url": schema.StringAttribute{
				MarkdownDescription: "https URL of the app homepage.",
				Optional:            true,
			},
			"branding": schema.StringAttribute{
				CustomType: jsonObjectType{},
				MarkdownDescription: "How the Sign in with Atlas page and consent screen look for this app, as a JSON " +
					"object (use `jsonencode`): `theme` (`auto`|`light`|`dark`), `accent_color`, `background_color` " +
					"(`#RRGGBB`), `privacy_policy_url`, `terms_of_service_url`, `copy` (`sign_in_title`, " +
					"`sign_in_subtitle`, `sign_up_title`, `sign_up_subtitle`, `consent_message`, `support_url`, " +
					"`support_label`) and `design` (sign-in page design settings). Compared semantically; a key " +
					"removed from the configuration is cleared on the server.",
				Optional: true,
			},
			"rotate_secret_trigger": schema.StringAttribute{
				MarkdownDescription: "Any value; changing it rotates the client secret in place (the new secret lands " +
					"in `client_secret`; the previous one keeps working for 24 hours). Confidential apps only.",
				Optional: true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The client secret (confidential apps), revealed on create and on rotation only. " +
					"Not recoverable on import.",
				Computed:      true,
				Sensitive:     true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"secret_prefix": schema.StringAttribute{
				MarkdownDescription: "The first characters of the current secret, for identification.",
				Computed:            true,
			},
			"managed_by_instance_id": schema.StringAttribute{
				MarkdownDescription: "Set when the app is managed by an instance's Sign in with Atlas toggle (read-only there).",
				Computed:            true,
			},
			"grants_managed_by_instance_id": schema.StringAttribute{
				MarkdownDescription: "Adopt the app into an instance of the same account: that instance's scoped key can " +
					"then list and revoke the app's grants. The app stays editable here. Removing the attribute releases " +
					"the app. Not for an app managed by an instance's Sign in with Atlas toggle.",
				Optional: true,
			},
			"managed_via": schema.StringAttribute{
				MarkdownDescription: "How the app's grants are managed: `instance_toggle`, `adopted`, or unset.",
				Computed:            true,
			},
			"test_mode": schema.BoolAttribute{
				MarkdownDescription: "Test sign-in mode (default false): an address with `+atlas_test` on a domain your " +
					"account has verified signs in to **this app only** with the fixed code `424242`, like a development " +
					"instance, and gets a test account confined to this app. Turning it on needs a verified domain on the " +
					"account, otherwise the API refuses with `DOMAIN_NOT_VERIFIED`.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"provisioned_verified_domains_only": schema.BoolAttribute{
				MarkdownDescription: "Require this app's test and review accounts (`atlas_id_provisioned_account`) to use " +
					"an email address on a domain your account has verified. Default false.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time (epoch ms).",
				PlanModifiers: nil},
			"updated_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

// writeBody builds the create / PATCH body. On update, nullable fields the
// configuration no longer sets are sent as explicit nulls (which clear them),
// and branding is sent as a merge-patch from the prior state to the plan.
func (r *atlasIDAppResource) writeBody(ctx context.Context, plan, prior *atlasIDAppModel, diags *diag.Diagnostics) client.AtlasIDAppWrite {
	body := client.AtlasIDAppWrite{
		"name":          plan.Name.ValueString(),
		"redirect_uris": listToStringSlice(ctx, plan.RedirectURIs, diags),
	}
	update := prior != nil
	if update && body["redirect_uris"] == nil {
		body["redirect_uris"] = []string{}
	}
	if !plan.PostLogoutRedirectURIs.IsUnknown() && !plan.PostLogoutRedirectURIs.IsNull() {
		v := listToStringSlice(ctx, plan.PostLogoutRedirectURIs, diags)
		if v == nil {
			v = []string{}
		}
		body["post_logout_redirect_uris"] = v
	}
	if !plan.AllowedScopes.IsUnknown() && !plan.AllowedScopes.IsNull() {
		body["allowed_scopes"] = listToStringSlice(ctx, plan.AllowedScopes, diags)
	}
	for key, b := range map[string]types.Bool{
		"backchannel_logout_session_required":  plan.BackchannelLogoutSessionRequired,
		"frontchannel_logout_session_required": plan.FrontchannelLogoutSessionRequired,
		"test_mode":                            plan.TestMode,
		"provisioned_verified_domains_only":    plan.ProvisionedVerifiedDomainsOnly,
	} {
		if !b.IsNull() && !b.IsUnknown() {
			body[key] = b.ValueBool()
		}
	}
	for key, s := range map[string]types.String{
		"backchannel_logout_uri":  plan.BackchannelLogoutURI,
		"frontchannel_logout_uri": plan.FrontchannelLogoutURI,
		"logo_url":                plan.LogoURL,
		"homepage_url":            plan.HomepageURL,
	} {
		switch {
		case !s.IsNull() && !s.IsUnknown():
			body[key] = s.ValueString()
		case update && s.IsNull():
			body[key] = nil
		}
	}
	if !update {
		if t := plan.Type; !t.IsNull() && !t.IsUnknown() {
			body["type"] = t.ValueString()
		}
	}

	desired := jsonObjectToMap(plan.Branding, path.Root("branding"), diags)
	switch {
	case !update:
		if desired != nil {
			body["branding"] = desired
		}
	case desired == nil && !prior.Branding.IsNull():
		body["branding"] = nil
	case desired != nil:
		prev := jsonObjectToMap(prior.Branding, path.Root("branding"), diags)
		if prev == nil {
			prev = map[string]interface{}{}
		}
		body["branding"] = mergePatchFor(prev, desired)
	}
	return body
}

func (r *atlasIDAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan atlasIDAppModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := r.writeBody(ctx, &plan, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateAtlasIDApp(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sign in with Atlas app", err.Error())
		return
	}
	adoptInto := plan.GrantsManagedByInstanceID
	mapAtlasIDApp(ctx, created, &plan, &resp.Diagnostics)
	plan.ClientSecret = stringPtrToValue(created.ClientSecret)
	if !adoptInto.IsNull() && !adoptInto.IsUnknown() {
		target := adoptInto.ValueString()
		adopted, err := r.client.SetAtlasIDAppManagedBy(ctx, created.ID, &target)
		if err != nil {
			// The app exists: keep it in state (tainted) so it is not orphaned.
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			resp.Diagnostics.AddError("Unable to adopt Sign in with Atlas app", err.Error())
			return
		}
		mapAtlasIDApp(ctx, adopted, &plan, &resp.Diagnostics)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *atlasIDAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state atlasIDAppModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	app, err := r.client.GetAtlasIDApp(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read Sign in with Atlas app", err.Error())
		return
	}
	mapAtlasIDApp(ctx, app, &state, &resp.Diagnostics)
	if state.ClientSecret.IsUnknown() {
		state.ClientSecret = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *atlasIDAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state atlasIDAppModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := r.writeBody(ctx, &plan, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateAtlasIDApp(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sign in with Atlas app", err.Error())
		return
	}
	plan.ClientSecret = state.ClientSecret
	if !plan.RotateSecretTrigger.Equal(state.RotateSecretTrigger) {
		rotated, err := r.client.RotateAtlasIDAppSecret(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to rotate Sign in with Atlas app secret", err.Error())
			return
		}
		updated = rotated
		plan.ClientSecret = stringPtrToValue(rotated.ClientSecret)
	}
	if !plan.GrantsManagedByInstanceID.Equal(state.GrantsManagedByInstanceID) {
		var target *string
		if !plan.GrantsManagedByInstanceID.IsNull() && !plan.GrantsManagedByInstanceID.IsUnknown() {
			v := plan.GrantsManagedByInstanceID.ValueString()
			target = &v
		}
		adopted, err := r.client.SetAtlasIDAppManagedBy(ctx, state.ID.ValueString(), target)
		if err != nil {
			resp.Diagnostics.AddError("Unable to change Sign in with Atlas app adoption", err.Error())
			return
		}
		// The managed_by response has no secret fields; keep the rotated one.
		adopted.ClientSecret = nil
		updated = adopted
	}
	mapAtlasIDApp(ctx, updated, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *atlasIDAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state atlasIDAppModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAtlasIDApp(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sign in with Atlas app", err.Error())
	}
}

// ImportState imports by app id. The client secret cannot be recovered; rotate
// it (rotate_secret_trigger) if you need it in state.
func (r *atlasIDAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapAtlasIDApp projects the API object onto the model. Branding is projected
// relative to what the model already holds (see projectAppBranding).
func mapAtlasIDApp(ctx context.Context, a *client.AtlasIDApp, m *atlasIDAppModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(a.ID)
	m.ClientID = types.StringValue(a.ClientID)
	m.Name = types.StringValue(a.Name)
	m.Type = types.StringValue(a.Type)
	m.RedirectURIs = stringSliceToList(ctx, a.RedirectURIs, diags)
	m.PostLogoutRedirectURIs = stringSliceToList(ctx, a.PostLogoutRedirectURIs, diags)
	m.BackchannelLogoutURI = stringPtrToValue(nonEmpty(a.BackchannelLogoutURI))
	m.BackchannelLogoutSessionRequired = types.BoolValue(a.BackchannelLogoutSessionRequired)
	m.FrontchannelLogoutURI = stringPtrToValue(nonEmpty(a.FrontchannelLogoutURI))
	m.FrontchannelLogoutSessionRequired = types.BoolValue(a.FrontchannelLogoutSessionRequired)
	m.AllowedScopes = stringSliceToList(ctx, a.AllowedScopes, diags)
	m.LogoURL = stringPtrToValue(nonEmpty(a.LogoURL))
	m.HomepageURL = stringPtrToValue(nonEmpty(a.HomepageURL))
	m.Branding = projectAppBranding(a.Branding, m.Branding)
	m.SecretPrefix = stringPtrToValue(a.SecretPrefix)
	m.ManagedByInstanceID = stringPtrToValue(a.ManagedByInstanceID)
	m.ManagedVia = stringPtrToValue(a.ManagedVia)
	// Only an adopted app maps back to the configurable attribute; a toggle app
	// reports its instance through managed_by_instance_id alone.
	if a.ManagedVia != nil && *a.ManagedVia == "adopted" {
		m.GrantsManagedByInstanceID = stringPtrToValue(a.ManagedByInstanceID)
	} else {
		m.GrantsManagedByInstanceID = types.StringNull()
	}
	m.TestMode = types.BoolValue(a.TestMode)
	m.ProvisionedVerifiedDomainsOnly = types.BoolValue(a.ProvisionedVerifiedDomainsOnly)
	m.CreatedAt = types.Int64Value(a.CreatedAt)
	m.UpdatedAt = types.Int64Value(a.UpdatedAt)
}

// projectAppBranding turns the API's branding (which always carries every
// field, nulls for unset ones, a default theme and the computed
// resolved_design) into the configuration-shaped JSON: resolved_design is
// dropped, nulls are pruned, and the default theme ("auto") is dropped unless
// the configuration itself sets a theme.
func projectAppBranding(raw json.RawMessage, prior jsonObjectValue) jsonObjectValue {
	if len(raw) == 0 || string(raw) == "null" {
		return newJSONObjectNull()
	}
	var b map[string]interface{}
	if err := json.Unmarshal(raw, &b); err != nil {
		return newJSONObjectValue(string(raw))
	}
	delete(b, "resolved_design")
	if b["theme"] == "auto" {
		var p map[string]interface{}
		if !prior.IsNull() && !prior.IsUnknown() {
			_ = json.Unmarshal([]byte(prior.ValueString()), &p)
		}
		if _, set := p["theme"]; !set {
			delete(b, "theme")
		}
	}
	enc, _ := json.Marshal(b)
	return jsonObjectFromAPI(enc)
}

// nonEmpty maps an empty string to nil (the API clears a URI to "" or null).
func nonEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

// ModifyPlan plans a new client_secret when rotate_secret_trigger changes.
func (r *atlasIDAppResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	planSecretRotation(ctx, req, resp, "client_secret")
}
