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
	_ resource.Resource                = &ssoConnectionResource{}
	_ resource.ResourceWithImportState = &ssoConnectionResource{}
	_ resource.ResourceWithConfigure   = &ssoConnectionResource{}
)

// NewSsoConnectionResource is the factory registered with the provider.
func NewSsoConnectionResource() resource.Resource { return &ssoConnectionResource{} }

type ssoConnectionResource struct {
	client *client.Client
}

type claimRoleMappingModel struct {
	Claim   types.String `tfsdk:"claim"`
	Value   types.String `tfsdk:"value"`
	RoleKey types.String `tfsdk:"role_key"`
}

type ssoConnectionModel struct {
	ID                     types.String            `tfsdk:"id"`
	OrganizationID         types.String            `tfsdk:"organization_id"`
	Type                   types.String            `tfsdk:"type"`
	Status                 types.String            `tfsdk:"status"`
	OidcIssuer             types.String            `tfsdk:"oidc_issuer"`
	OidcClientID           types.String            `tfsdk:"oidc_client_id"`
	OidcClientSecret       types.String            `tfsdk:"oidc_client_secret"`
	SamlIdpEntityID        types.String            `tfsdk:"saml_idp_entity_id"`
	SamlIdpSsoURL          types.String            `tfsdk:"saml_idp_sso_url"`
	SamlIdpCertificate     types.String            `tfsdk:"saml_idp_certificate"`
	SamlSpEntityID         types.String            `tfsdk:"saml_sp_entity_id"`
	SamlAllowIdpInitiated  types.Bool              `tfsdk:"saml_allow_idp_initiated"`
	SamlSignAuthnRequests  types.Bool              `tfsdk:"saml_sign_authn_requests"`
	SamlWantResponseSigned types.Bool              `tfsdk:"saml_want_response_signed"`
	DiscourseSecret        types.String            `tfsdk:"discourse_secret"`
	DiscourseProviderURL   types.String            `tfsdk:"discourse_provider_url"`
	AllowedDomains         types.List              `tfsdk:"allowed_domains"`
	ClaimRoleMappings      []claimRoleMappingModel `tfsdk:"claim_role_mappings"`
	DefaultRoleID          types.String            `tfsdk:"default_role_id"`
	HasSecret              types.Bool              `tfsdk:"has_secret"`
	HasSamlCertificate     types.Bool              `tfsdk:"has_saml_certificate"`
	HasDiscourseSecret     types.Bool              `tfsdk:"has_discourse_secret"`
	CreatedAt              types.Int64             `tfsdk:"created_at"`
	UpdatedAt              types.Int64             `tfsdk:"updated_at"`
}

func (r *ssoConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sso_connection"
}

func (r *ssoConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *ssoConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An enterprise SSO connection (oidc, saml or discourse). Secrets (OIDC client secret, " +
			"SAML IdP certificate, Discourse shared secret) are write-only: the API never returns them, so they are " +
			"tracked only via the has_* booleans and by whatever is set in configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Connection protocol: oidc, saml or discourse. Immutable — forces replacement.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "draft, active or disabled. Activation requires the protocol's mandatory fields.",
				Optional:            true,
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Optional organization this connection is scoped to.",
				Optional:            true,
			},
			"oidc_issuer": schema.StringAttribute{
				MarkdownDescription: "OIDC issuer URL.",
				Optional:            true,
			},
			"oidc_client_id": schema.StringAttribute{
				MarkdownDescription: "OIDC client id at the identity provider.",
				Optional:            true,
			},
			"oidc_client_secret": schema.StringAttribute{
				MarkdownDescription: "OIDC client secret. Write-only; never returned by the API.",
				Optional:            true,
				Sensitive:           true,
			},
			"saml_idp_entity_id": schema.StringAttribute{
				MarkdownDescription: "SAML IdP entity id.",
				Optional:            true,
			},
			"saml_idp_sso_url": schema.StringAttribute{
				MarkdownDescription: "SAML IdP single sign-on URL.",
				Optional:            true,
			},
			"saml_idp_certificate": schema.StringAttribute{
				MarkdownDescription: "SAML IdP signing certificate (PEM). Write-only; never returned by the API.",
				Optional:            true,
				Sensitive:           true,
			},
			"saml_sp_entity_id": schema.StringAttribute{
				MarkdownDescription: "SAML service-provider entity id.",
				Optional:            true,
			},
			"saml_allow_idp_initiated": schema.BoolAttribute{
				MarkdownDescription: "Allow IdP-initiated SAML sign-in.",
				Optional:            true,
				Computed:            true,
			},
			"saml_sign_authn_requests": schema.BoolAttribute{
				MarkdownDescription: "Sign outgoing SAML AuthnRequests.",
				Optional:            true,
				Computed:            true,
			},
			"saml_want_response_signed": schema.BoolAttribute{
				MarkdownDescription: "Require the SAML response itself to be signed.",
				Optional:            true,
				Computed:            true,
			},
			"discourse_secret": schema.StringAttribute{
				MarkdownDescription: "Discourse SSO shared secret. Write-only; never returned by the API.",
				Optional:            true,
				Sensitive:           true,
			},
			"discourse_provider_url": schema.StringAttribute{
				MarkdownDescription: "Discourse provider URL.",
				Optional:            true,
			},
			"allowed_domains": schema.ListAttribute{
				MarkdownDescription: "Email domains this connection applies to.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"default_role_id": schema.StringAttribute{
				MarkdownDescription: "Role granted to users who sign in through this connection when no claim mapping matches.",
				Optional:            true,
			},
			"claim_role_mappings": schema.ListNestedAttribute{
				MarkdownDescription: "Map an IdP claim value onto an Atlas role key.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"claim":    schema.StringAttribute{Required: true, MarkdownDescription: "The claim name."},
						"value":    schema.StringAttribute{Required: true, MarkdownDescription: "The claim value to match."},
						"role_key": schema.StringAttribute{Required: true, MarkdownDescription: "The Atlas role key to grant."},
					},
				},
			},
			"has_secret": schema.BoolAttribute{
				MarkdownDescription: "Whether an OIDC client secret is stored.",
				Computed:            true,
			},
			"has_saml_certificate": schema.BoolAttribute{
				MarkdownDescription: "Whether a SAML IdP certificate is stored.",
				Computed:            true,
			},
			"has_discourse_secret": schema.BoolAttribute{
				MarkdownDescription: "Whether a Discourse shared secret is stored.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time (epoch ms)."},
			"updated_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

func (r *ssoConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ssoConnectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := client.SsoConnectionWrite{Type: client.Ptr(plan.Type.ValueString())}
	applyWritable(ctx, &plan, &body, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateSsoConnection(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create SSO connection", err.Error())
		return
	}
	mapSsoToState(ctx, created, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ssoConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ssoConnectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetSsoConnection(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read SSO connection", err.Error())
		return
	}
	mapSsoToState(ctx, fetched, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ssoConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ssoConnectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.SsoConnectionWrite{}
	applyWritable(ctx, &plan, &body, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateSsoConnection(ctx, plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update SSO connection", err.Error())
		return
	}
	mapSsoToState(ctx, updated, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ssoConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ssoConnectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSsoConnection(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete SSO connection", err.Error())
	}
}

func (r *ssoConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// applyWritable copies every writable planned attribute onto the API payload.
// Optional string/bool fields are sent only when set, matching the API's PATCH
// merge semantics so an unrelated save never clobbers a sibling field.
func applyWritable(ctx context.Context, m *ssoConnectionModel, body *client.SsoConnectionWrite, diags *diag.Diagnostics) {
	body.OrganizationID = optionalString(m.OrganizationID)
	body.Status = optionalString(m.Status)
	body.OidcIssuer = optionalString(m.OidcIssuer)
	body.OidcClientID = optionalString(m.OidcClientID)
	body.OidcClientSecret = optionalString(m.OidcClientSecret)
	body.SamlIdpEntityID = optionalString(m.SamlIdpEntityID)
	body.SamlIdpSsoURL = optionalString(m.SamlIdpSsoURL)
	body.SamlIdpCertificate = optionalString(m.SamlIdpCertificate)
	body.SamlSpEntityID = optionalString(m.SamlSpEntityID)
	body.DiscourseSecret = optionalString(m.DiscourseSecret)
	body.DiscourseProviderURL = optionalString(m.DiscourseProviderURL)
	body.DefaultRoleID = optionalString(m.DefaultRoleID)
	if !m.SamlAllowIdpInitiated.IsNull() && !m.SamlAllowIdpInitiated.IsUnknown() {
		body.SamlAllowIdpInitiated = client.Ptr(m.SamlAllowIdpInitiated.ValueBool())
	}
	if !m.SamlSignAuthnRequests.IsNull() && !m.SamlSignAuthnRequests.IsUnknown() {
		body.SamlSignAuthnRequests = client.Ptr(m.SamlSignAuthnRequests.ValueBool())
	}
	if !m.SamlWantResponseSigned.IsNull() && !m.SamlWantResponseSigned.IsUnknown() {
		body.SamlWantResponseSigned = client.Ptr(m.SamlWantResponseSigned.ValueBool())
	}
	if domains := listToStringSlice(ctx, m.AllowedDomains, diags); domains != nil {
		body.AllowedDomains = domains
	}
	if len(m.ClaimRoleMappings) > 0 {
		mappings := make([]client.ClaimRoleMapping, 0, len(m.ClaimRoleMappings))
		for _, cm := range m.ClaimRoleMappings {
			mappings = append(mappings, client.ClaimRoleMapping{
				Claim:   cm.Claim.ValueString(),
				Value:   cm.Value.ValueString(),
				RoleKey: cm.RoleKey.ValueString(),
			})
		}
		body.ClaimRoleMappings = mappings
	}
}

// mapSsoToState projects the API object onto the model. Write-only secrets are
// deliberately left untouched (they are never returned), so their configured
// values survive a read; drift on every non-secret field is picked up here.
func mapSsoToState(ctx context.Context, c *client.SsoConnection, m *ssoConnectionModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(c.ID)
	m.Type = types.StringValue(c.Type)
	m.Status = types.StringValue(c.Status)
	m.OrganizationID = stringPtrToValue(c.OrganizationID)
	m.OidcIssuer = stringPtrToValue(c.OidcIssuer)
	m.OidcClientID = stringPtrToValue(c.OidcClientID)
	m.SamlIdpEntityID = stringPtrToValue(c.SamlIdpEntityID)
	m.SamlIdpSsoURL = stringPtrToValue(c.SamlIdpSsoURL)
	m.SamlSpEntityID = stringPtrToValue(c.SamlSpEntityID)
	m.SamlAllowIdpInitiated = types.BoolValue(c.SamlAllowIdpInitiated)
	m.SamlSignAuthnRequests = types.BoolValue(c.SamlSignAuthnRequests)
	m.SamlWantResponseSigned = types.BoolValue(c.SamlWantResponseSigned)
	m.DiscourseProviderURL = stringPtrToValue(c.DiscourseProviderURL)
	m.DefaultRoleID = stringPtrToValue(c.DefaultRoleID)
	m.HasSecret = types.BoolValue(c.HasSecret)
	m.HasSamlCertificate = types.BoolValue(c.HasSamlCertificate)
	m.HasDiscourseSecret = types.BoolValue(c.HasDiscourseSecret)
	m.CreatedAt = types.Int64Value(c.CreatedAt)
	m.UpdatedAt = types.Int64Value(c.UpdatedAt)

	// allowed_domains is Optional (not Computed): only reflect it when the model
	// already tracks it, to avoid a null->[] plan-vs-state inconsistency.
	if !m.AllowedDomains.IsNull() {
		m.AllowedDomains = stringSliceToList(ctx, c.AllowedDomains, diags)
	}
	if len(m.ClaimRoleMappings) > 0 || len(c.ClaimRoleMappings) > 0 {
		mappings := make([]claimRoleMappingModel, 0, len(c.ClaimRoleMappings))
		for _, cm := range c.ClaimRoleMappings {
			mappings = append(mappings, claimRoleMappingModel{
				Claim:   types.StringValue(cm.Claim),
				Value:   types.StringValue(cm.Value),
				RoleKey: types.StringValue(cm.RoleKey),
			})
		}
		if len(mappings) > 0 {
			m.ClaimRoleMappings = mappings
		}
	}
}
