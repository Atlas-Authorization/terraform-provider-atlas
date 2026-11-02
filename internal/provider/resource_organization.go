package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
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
}

func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An organization (tenant). Creation requires an existing user id as created_by, who becomes " +
			"the first admin — Terraform manages the org's declarative profile, not its runtime membership.",
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
				MarkdownDescription: "User id of the initial admin. Immutable — changing it forces replacement.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
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
	var maxMemberships *int64
	if !plan.MaxAllowedMemberships.IsNull() && !plan.MaxAllowedMemberships.IsUnknown() {
		maxMemberships = client.Ptr(plan.MaxAllowedMemberships.ValueInt64())
	}
	created, err := r.client.CreateOrganization(ctx, plan.Name.ValueString(), plan.Slug.ValueString(), plan.CreatedBy.ValueString(), maxMemberships)
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
	updated, err := r.client.UpdateOrganization(ctx, plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization", err.Error())
		return
	}
	// created_by is not returned by the profile projection; carry it forward.
	r.mapToState(updated, &plan)
	plan.CreatedBy = state.CreatedBy
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
	}
}
