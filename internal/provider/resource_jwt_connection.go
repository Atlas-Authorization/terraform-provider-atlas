package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &jwtConnectionResource{}
	_ resource.ResourceWithImportState = &jwtConnectionResource{}
	_ resource.ResourceWithConfigure   = &jwtConnectionResource{}
)

// NewJWTConnectionResource is the factory for atlas_jwt_connection.
func NewJWTConnectionResource() resource.Resource { return &jwtConnectionResource{} }

type jwtConnectionResource struct {
	client *client.Client
}

type jwtConnectionModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Status         types.String `tfsdk:"status"`
	JWKSURI        types.String `tfsdk:"jwks_uri"`
	Issuer         types.String `tfsdk:"issuer"`
	Audience       types.String `tfsdk:"audience"`
	EmailClaim     types.String `tfsdk:"email_claim"`
	MaxAgeSeconds  types.Int64  `tfsdk:"max_age_seconds"`
	DefaultRoleID  types.String `tfsdk:"default_role_id"`
	LoginURL       types.String `tfsdk:"login_url"`
	CreatedAt      types.Int64  `tfsdk:"created_at"`
	UpdatedAt      types.Int64  `tfsdk:"updated_at"`
}

func (r *jwtConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jwt_connection"
}

func (r *jwtConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *jwtConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A trusted-JWT connection: your own system signs a short-lived JWT for a person and sends " +
			"them to `login_url?token=<jwt>&redirect_url=<your app URL>`. Atlas checks the signature against `jwks_uri`, " +
			"the `iss`, `aud`, `exp`, the max age and that the `jti` was never used before, then signs them in and seats " +
			"them in `organization_id`. An `active` connection needs `jwks_uri`, `issuer` and `audience`. Needs a secret " +
			"key with `sso_connections:write`. Import with the connection id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The connection id.",
				Computed:            true,
				PlanModifiers:       keep,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization people who sign in this way are seated in. Removing it clears it.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "`draft` (default), `active` or `disabled`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       keep,
			},
			"jwks_uri": schema.StringAttribute{
				MarkdownDescription: "An https URL serving the signer's JSON Web Key Set (public RS / PS / ES / EdDSA keys; HS keys are ignored).",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       keep,
			},
			"issuer": schema.StringAttribute{
				MarkdownDescription: "The exact `iss` the tokens carry.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       keep,
			},
			"audience": schema.StringAttribute{
				MarkdownDescription: "An `aud` value the tokens must carry.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       keep,
			},
			"email_claim": schema.StringAttribute{
				MarkdownDescription: "The claim carrying the email address (default `email`). `email_verified: true` lets " +
					"the sign-in link to an existing account with that verified email.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: keep,
			},
			"max_age_seconds": schema.Int64Attribute{
				MarkdownDescription: "The longest a token may live (`exp` - `iat`), 30-3600 seconds (default 300).",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"default_role_id": schema.StringAttribute{
				MarkdownDescription: "The organization role a newly seated member gets (else `org:member`). Removing it clears it.",
				Optional:            true,
			},
			"login_url": schema.StringAttribute{
				MarkdownDescription: "Send the person here with `?token=<jwt>&redirect_url=<your app URL>`.",
				Computed:            true,
				PlanModifiers:       keep,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch ms).",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.Int64Attribute{
				MarkdownDescription: "Last update time (epoch ms).",
				Computed:            true,
			},
		},
	}
}

// jwtConnectionBody is the create / PATCH body. organization_id and
// default_role_id are always sent (null clears them); the rest only when known,
// since Atlas keeps (and defaults) them when left out.
func jwtConnectionBody(m *jwtConnectionModel) client.JWTConnectionWrite {
	body := client.JWTConnectionWrite{
		"organization_id": optionalString(m.OrganizationID),
		"default_role_id": optionalString(m.DefaultRoleID),
	}
	for key, v := range map[string]types.String{
		"status": m.Status, "jwks_uri": m.JWKSURI, "issuer": m.Issuer, "audience": m.Audience, "email_claim": m.EmailClaim,
	} {
		if !v.IsNull() && !v.IsUnknown() {
			body[key] = v.ValueString()
		}
	}
	if !m.MaxAgeSeconds.IsNull() && !m.MaxAgeSeconds.IsUnknown() {
		body["max_age_seconds"] = m.MaxAgeSeconds.ValueInt64()
	}
	return body
}

func (r *jwtConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan jwtConnectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateJWTConnection(ctx, jwtConnectionBody(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create JWT connection", err.Error())
		return
	}
	mapJWTConnection(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jwtConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state jwtConnectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetJWTConnection(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read JWT connection", err.Error())
		return
	}
	mapJWTConnection(fetched, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *jwtConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state jwtConnectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateJWTConnection(ctx, state.ID.ValueString(), jwtConnectionBody(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update JWT connection", err.Error())
		return
	}
	mapJWTConnection(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jwtConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state jwtConnectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteJWTConnection(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete JWT connection", err.Error())
	}
}

func (r *jwtConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapJWTConnection(c *client.JWTConnection, m *jwtConnectionModel) {
	m.ID = types.StringValue(c.ID)
	m.OrganizationID = stringPtrToValue(c.OrganizationID)
	m.Status = types.StringValue(c.Status)
	m.JWKSURI = stringPtrToValue(c.JWKSURI)
	m.Issuer = stringPtrToValue(c.Issuer)
	m.Audience = stringPtrToValue(c.Audience)
	m.EmailClaim = types.StringValue(c.EmailClaim)
	m.MaxAgeSeconds = types.Int64Value(c.MaxAgeSeconds)
	m.DefaultRoleID = stringPtrToValue(c.DefaultRoleID)
	m.LoginURL = types.StringValue(c.LoginURL)
	m.CreatedAt = types.Int64Value(c.CreatedAt)
	m.UpdatedAt = types.Int64Value(c.UpdatedAt)
}
