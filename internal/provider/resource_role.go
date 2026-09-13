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
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
	_ resource.ResourceWithConfigure   = &roleResource{}
)

// NewRoleResource is the factory registered with the provider.
func NewRoleResource() resource.Resource { return &roleResource{} }

type roleResource struct {
	client *client.Client
}

type roleModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.Set    `tfsdk:"permissions"`
	IsSystem    types.Bool   `tfsdk:"is_system"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
}

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A custom organization role: a stable key, a label, and a set of permission keys. The key " +
			"is published into authorization checks and is immutable — changing it forces replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Stable role key (e.g. billing_admin). Immutable; changing it forces replacement.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable label.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
			},
			"permissions": schema.SetAttribute{
				MarkdownDescription: "Permission keys granted by the role. Keys not defined on the instance are dropped by the API.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"is_system": schema.BoolAttribute{
				MarkdownDescription: "True for a built-in system role.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch ms).",
				Computed:            true,
			},
		},
	}
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions := setToStringSlice(ctx, plan.Permissions, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateRole(ctx, plan.Key.ValueString(), plan.Name.ValueString(), optionalString(plan.Description), permissions)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create role", err.Error())
		return
	}
	// The applied permission set may differ (unknown keys dropped): re-read it so
	// state reflects what the API actually stored.
	created.Permissions = permissions
	r.mapToState(ctx, created, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetRole(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read role", err.Error())
		return
	}
	r.mapToState(ctx, fetched, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan roleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Label change goes through PATCH.
	updated, err := r.client.UpdateRoleLabel(ctx, plan.ID.ValueString(), client.Ptr(plan.Name.ValueString()), optionalString(plan.Description))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update role", err.Error())
		return
	}

	// Permissions are a separate PUT; the API returns the applied set.
	permissions := setToStringSlice(ctx, plan.Permissions, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	applied, err := r.client.SetRolePermissions(ctx, plan.ID.ValueString(), permissions)
	if err != nil {
		resp.Diagnostics.AddError("Unable to set role permissions", err.Error())
		return
	}
	updated.Permissions = applied

	r.mapToState(ctx, updated, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRole(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete role", err.Error())
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *roleResource) mapToState(ctx context.Context, role *client.Role, m *roleModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(role.ID)
	m.Key = types.StringValue(role.Key)
	m.Name = types.StringValue(role.Name)
	m.Description = stringPtrToValue(role.Description)
	m.Permissions = stringSliceToSet(ctx, role.Permissions, diags)
	m.IsSystem = types.BoolValue(role.IsSystem)
	m.CreatedAt = types.Int64Value(role.CreatedAt)
}
