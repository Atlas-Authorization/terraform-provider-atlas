package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &oauthClientResource{}
	_ resource.ResourceWithImportState = &oauthClientResource{}
	_ resource.ResourceWithConfigure   = &oauthClientResource{}
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
	m.SecretPrefix = stringPtrToValue(c.SecretPrefix)
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
