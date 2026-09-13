package provider

import (
	"context"

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
	_ resource.Resource                = &oauthProviderResource{}
	_ resource.ResourceWithImportState = &oauthProviderResource{}
	_ resource.ResourceWithConfigure   = &oauthProviderResource{}
)

// NewOAuthProviderResource is the factory registered with the provider.
func NewOAuthProviderResource() resource.Resource { return &oauthProviderResource{} }

type oauthProviderResource struct {
	client *client.Client
}

// oauthProviderModel maps a configured SOCIAL sign-in provider. `credentials`
// (the client id/secret keyed by the provider's field keys) is write-only: the
// API stores the secret encrypted and never returns it, so presence is tracked
// via has_secret and by whatever configuration sets. `settings` is likewise not
// read back — both survive a Read untouched.
type oauthProviderModel struct {
	Provider    types.String `tfsdk:"provider_key"`
	Credentials types.Map    `tfsdk:"credentials"`
	Settings    types.Map    `tfsdk:"settings"`
	Scopes      types.List   `tfsdk:"scopes"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	AllowSignIn types.Bool   `tfsdk:"allow_sign_in"`
	AllowSignUp types.Bool   `tfsdk:"allow_sign_up"`
	ID          types.String `tfsdk:"id"`
	ClientID    types.String `tfsdk:"client_id"`
	HasSecret   types.Bool   `tfsdk:"has_secret"`
	Configured  types.Bool   `tfsdk:"configured"`
	RedirectURI types.String `tfsdk:"redirect_uri"`
	DisplayName types.String `tfsdk:"display_name"`
	UpdatedAt   types.Int64  `tfsdk:"updated_at"`
}

func (r *oauthProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_provider"
}

func (r *oauthProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *oauthProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A configured SOCIAL sign-in provider (Google, GitHub, ... — a key from the " +
			"Atlas provider catalog). Maps the secret-key Backend API's oauth_providers routes. The client " +
			"secret is write-only: the API never returns it, so it is tracked only via `has_secret` and by " +
			"whatever `credentials` sets. The redirect URI is derived from your instance host and returned " +
			"read-only — register that exact string at the provider console.",
		Attributes: map[string]schema.Attribute{
			"provider_key": schema.StringAttribute{
				MarkdownDescription: "Catalog key of the social provider (e.g. `google`, `github`). Immutable — forces " +
					"replacement. (Named `provider_key` because `provider` is a reserved Terraform meta-argument.)",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"credentials": schema.MapAttribute{
				MarkdownDescription: "The provider credentials, keyed by the provider's own field keys — typically " +
					"`client_id` and `client_secret`. The secret is write-only; never returned by the API.",
				Required:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"settings": schema.MapAttribute{
				MarkdownDescription: "Optional provider-specific settings (e.g. a team/tenant domain), keyed by setting key. Write-only; not read back.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"scopes": schema.ListAttribute{
				MarkdownDescription: "OAuth scopes to request. Defaults to the provider's catalog default scopes when unset.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the provider is enabled for sign-in. Enabling one with no usable credentials is refused.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"allow_sign_in": schema.BoolAttribute{
				MarkdownDescription: "Whether existing users may sign in with this provider. Defaults to true. `allow_sign_in` and `allow_sign_up` cannot both be false — disable the provider instead.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"allow_sign_up": schema.BoolAttribute{
				MarkdownDescription: "Whether new users may sign up with this provider (JIT account creation). Defaults to true.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource id (equals the provider key).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"client_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "The configured client id (not a secret)."},
			"has_secret":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether a client secret is stored."},
			"configured":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the provider has credentials configured."},
			"redirect_uri": schema.StringAttribute{Computed: true, MarkdownDescription: "The redirect/callback URI derived from your instance host. Register this exact string at the provider."},
			"display_name": schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable provider name from the catalog."},
			"updated_at":   schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

func (r *oauthProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan oauthProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oauthProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state oauthProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetOAuthProvider(ctx, state.Provider.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read OAuth provider", err.Error())
		return
	}
	// Credentials removed out of band (or never configured) means the resource is
	// gone as far as Terraform is concerned — the GET still returns the catalog
	// entry, so distinguish on `configured`.
	if !fetched.Configured {
		resp.State.RemoveResource(ctx)
		return
	}
	mapOAuthProviderToState(ctx, fetched, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *oauthProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan oauthProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oauthProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state oauthProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOAuthProvider(ctx, state.Provider.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete OAuth provider", err.Error())
	}
}

// ImportState imports by provider key: `terraform import atlas_oauth_provider.google google`.
// Write-only credentials/settings are not recovered — a following apply re-sends them.
func (r *oauthProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("provider_key"), req, resp)
}

// apply drives a provider to the planned state: PUT the credentials (idempotent),
// then set the sign-in/sign-up scope and the enabled toggle, then GET the final
// projection. The PUT runs first so a missing client id fails before we ever try
// to enable an unusable provider. Credentials and settings are left on the model
// untouched by the read-back, since the API never returns them.
func (r *oauthProviderResource) apply(ctx context.Context, m *oauthProviderModel, diags *diag.Diagnostics) {
	providerKey := m.Provider.ValueString()

	body := client.OAuthProviderWrite{
		Values:   mapToStringMap(ctx, m.Credentials, diags),
		Settings: mapToStringMap(ctx, m.Settings, diags),
		Scopes:   listToStringSlice(ctx, m.Scopes, diags),
	}
	if diags.HasError() {
		return
	}
	if err := r.client.PutOAuthProvider(ctx, providerKey, body); err != nil {
		diags.AddError("Unable to configure OAuth provider", err.Error())
		return
	}

	// Both scope flags default to true (does sign-in and sign-up). The API rejects
	// "neither"; that error surfaces to the operator unchanged.
	if err := r.client.SetOAuthProviderScope(ctx, providerKey,
		boolOrDefault(m.AllowSignIn, true), boolOrDefault(m.AllowSignUp, true)); err != nil {
		diags.AddError("Unable to set OAuth provider scope", err.Error())
		return
	}

	if err := r.client.SetOAuthProviderEnabled(ctx, providerKey, boolOrDefault(m.Enabled, false)); err != nil {
		diags.AddError("Unable to set OAuth provider enabled state", err.Error())
		return
	}

	fetched, err := r.client.GetOAuthProvider(ctx, providerKey)
	if err != nil {
		diags.AddError("Unable to read back OAuth provider", err.Error())
		return
	}
	mapOAuthProviderToState(ctx, fetched, m, diags)
}

// boolOrDefault resolves an Optional+Computed bool: an unset (null/unknown) value
// falls back to def so an omitted attribute has a deterministic applied value.
func boolOrDefault(v types.Bool, def bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return def
	}
	return v.ValueBool()
}

// mapOAuthProviderToState projects the API object onto the model. Write-only
// credentials and settings are deliberately left untouched (never returned), so
// their configured values survive a read; every readable field picks up drift.
func mapOAuthProviderToState(ctx context.Context, c *client.OAuthProvider, m *oauthProviderModel, diags *diag.Diagnostics) {
	m.Provider = types.StringValue(c.Provider)
	m.ID = types.StringValue(c.Provider)
	m.ClientID = stringPtrToValue(c.ClientID)
	m.HasSecret = types.BoolValue(c.HasSecret)
	m.Configured = types.BoolValue(c.Configured)
	m.RedirectURI = types.StringValue(c.RedirectURI)
	m.DisplayName = types.StringValue(c.DisplayName)
	m.Enabled = types.BoolValue(c.Enabled)
	m.AllowSignIn = types.BoolValue(c.AllowSignIn)
	m.AllowSignUp = types.BoolValue(c.AllowSignUp)
	m.Scopes = stringSliceToList(ctx, c.Scopes, diags)
	m.UpdatedAt = int64PtrToValue(c.UpdatedAt)
}
