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
	_ resource.Resource                = &organizationResource{}
	_ resource.ResourceWithImportState = &organizationResource{}
	_ resource.ResourceWithConfigure   = &organizationResource{}
)

// NewOrganizationResource is the factory registered with the provider.
func NewOrganizationResource() resource.Resource { return &organizationResource{} }

type organizationResource struct {
	client *client.Client
}

type organizationModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Slug                  types.String `tfsdk:"slug"`
	CreatedBy             types.String `tfsdk:"created_by"`
	MaxAllowedMemberships types.Int64  `tfsdk:"max_allowed_memberships"`
	ImageURL              types.String `tfsdk:"image_url"`
	CreatedAt             types.Int64  `tfsdk:"created_at"`
	UpdatedAt             types.Int64  `tfsdk:"updated_at"`

	ExternalID      types.String    `tfsdk:"external_id"`
	PublicMetadata  jsonObjectValue `tfsdk:"public_metadata"`
	PrivateMetadata jsonObjectValue `tfsdk:"private_metadata"`
	RoleMappings    types.Map       `tfsdk:"role_mappings"`
}

func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An organization (tenant). `created_by` optionally seats an existing user as the first " +
			"admin; without it the organization starts with no members. Terraform manages the org's declarative " +
			"profile, not its runtime membership.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Organization name.",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-safe slug.",
				Required:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "User id seated as the initial admin. Optional: omit it for an organization created " +
					"by your backend, which then has no members yet. Immutable — changing it forces replacement.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"external_id": schema.StringAttribute{
				MarkdownDescription: "Your own id for this customer (1-255 printable characters, no spaces), unique per instance.",
				Optional:            true,
			},
			"public_metadata": schema.StringAttribute{
				CustomType: jsonObjectType{},
				MarkdownDescription: "Public metadata as a JSON object (use `jsonencode`). Replaces the stored object. Left " +
					"unset, Terraform does not manage it and reads back whatever is stored.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"private_metadata": schema.StringAttribute{
				CustomType: jsonObjectType{},
				MarkdownDescription: "Backend-only metadata as a JSON object (use `jsonencode`). Atlas never returns it, so " +
					"Terraform keeps the configured value and cannot detect drift; removing it clears the stored object.",
				Optional:  true,
				Sensitive: true,
			},
			"role_mappings": schema.MapAttribute{
				MarkdownDescription: "Atlas role key to your product's name for it (e.g. `{ \"org:admin\" = \"owner\" }`); a " +
					"member's mapped name rides in the `org_role_mapped` token claim. At most 50 roles.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"max_allowed_memberships": schema.Int64Attribute{
				MarkdownDescription: "Optional seat cap. 0 or omitted means unlimited.",
				Optional:            true,
				Computed:            true,
			},
			"image_url": schema.StringAttribute{
				MarkdownDescription: "Organization logo URL.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time (epoch ms)."},
			"updated_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.OrganizationCreate{
		Name:                  plan.Name.ValueString(),
		Slug:                  plan.Slug.ValueString(),
		CreatedBy:             optionalString(plan.CreatedBy),
		MaxAllowedMemberships: optionalInt64(plan.MaxAllowedMemberships),
		ExternalID:            optionalString(plan.ExternalID),
		PublicMetadata:        jsonObjectToMap(plan.PublicMetadata, path.Root("public_metadata"), &resp.Diagnostics),
		PrivateMetadata:       jsonObjectToMap(plan.PrivateMetadata, path.Root("private_metadata"), &resp.Diagnostics),
	}
	if !plan.RoleMappings.IsNull() && !plan.RoleMappings.IsUnknown() {
		body.RoleMappings = mapToStringMap(ctx, plan.RoleMappings, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateOrganization(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization", err.Error())
		return
	}
	r.mapToState(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetOrganization(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read organization", err.Error())
		return
	}
	r.mapToState(fetched, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state organizationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.OrganizationUpdate{
		Name: client.Ptr(plan.Name.ValueString()),
		Slug: client.Ptr(plan.Slug.ValueString()),
	}
	if !plan.MaxAllowedMemberships.IsNull() && !plan.MaxAllowedMemberships.IsUnknown() {
		body.MaxAllowedMemberships = client.Ptr(plan.MaxAllowedMemberships.ValueInt64())
	}
	organizationUpdateExtras(ctx, &plan, &state, &body, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateOrganization(ctx, plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization", err.Error())
		return
	}
	// created_by is not returned by the profile projection; carry it forward.
	plan.CreatedBy = state.CreatedBy
	r.mapToState(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrganization(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete organization", err.Error())
	}
}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// organizationUpdateExtras adds the fields that changed since state to the
// PATCH: an empty external_id or role_mappings clears it, and a removed
// private_metadata is sent as {} (public_metadata left unset is unmanaged).
func organizationUpdateExtras(ctx context.Context, plan, state *organizationModel, body *client.OrganizationUpdate, diags *diag.Diagnostics) {
	if !plan.ExternalID.Equal(state.ExternalID) {
		body.ExternalID = client.Ptr(plan.ExternalID.ValueString())
	}
	if !plan.RoleMappings.Equal(state.RoleMappings) {
		mappings := map[string]string{}
		if !plan.RoleMappings.IsNull() && !plan.RoleMappings.IsUnknown() {
			mappings = mapToStringMap(ctx, plan.RoleMappings, diags)
		}
		body.RoleMappings = &mappings
	}
	if !plan.PublicMetadata.IsUnknown() && !plan.PublicMetadata.Equal(state.PublicMetadata) {
		md := jsonObjectToMap(plan.PublicMetadata, path.Root("public_metadata"), diags)
		if md == nil {
			md = map[string]interface{}{}
		}
		body.PublicMetadata = &md
	}
	if !plan.PrivateMetadata.Equal(state.PrivateMetadata) {
		md := jsonObjectToMap(plan.PrivateMetadata, path.Root("private_metadata"), diags)
		if md == nil {
			md = map[string]interface{}{}
		}
		body.PrivateMetadata = &md
	}
}

func (r *organizationResource) mapToState(o *client.Organization, m *organizationModel) {
	m.ID = types.StringValue(o.ID)
	m.Name = types.StringValue(o.Name)
	m.Slug = types.StringValue(o.Slug)
	m.ImageURL = stringPtrToValue(o.ImageURL)
	m.MaxAllowedMemberships = int64PtrToValue(o.MaxAllowedMemberships)
	m.CreatedAt = types.Int64Value(o.CreatedAt)
	m.UpdatedAt = types.Int64Value(o.UpdatedAt)
	// created_by is populated by the caller (create) or preserved (update/read);
	// the profile projection may omit it, so only overwrite when present.
	if o.CreatedBy != nil {
		m.CreatedBy = types.StringValue(*o.CreatedBy)
	} else if m.CreatedBy.IsUnknown() {
		m.CreatedBy = types.StringNull()
	}
	m.ExternalID = stringPtrToValue(o.ExternalID)
	m.PublicMetadata = jsonObjectFromAPI(o.PublicMetadata)
	switch {
	case len(o.RoleMappings) > 0:
		m.RoleMappings, _ = types.MapValueFrom(context.Background(), types.StringType, o.RoleMappings)
	case !m.RoleMappings.IsNull() && !m.RoleMappings.IsUnknown() && len(m.RoleMappings.Elements()) == 0:
		// The API answers null for an empty mapping; keep a configured {} as-is.
	default:
		m.RoleMappings = types.MapNull(types.StringType)
	}
	// private_metadata is never returned: the configured value is kept.
	if m.PrivateMetadata.IsUnknown() {
		m.PrivateMetadata = newJSONObjectNull()
	}
}
