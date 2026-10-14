package provider

import (
	"context"
	"strings"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &entitlementOverrideResource{}
	_ resource.ResourceWithImportState = &entitlementOverrideResource{}
	_ resource.ResourceWithConfigure   = &entitlementOverrideResource{}
)

// NewEntitlementOverrideResource is the factory for atlas_entitlement_override.
func NewEntitlementOverrideResource() resource.Resource { return &entitlementOverrideResource{} }

type entitlementOverrideResource struct {
	client *client.Client
}

type entitlementOverrideModel struct {
	ID          types.String `tfsdk:"id"`
	SubjectType types.String `tfsdk:"subject_type"`
	SubjectID   types.String `tfsdk:"subject_id"`
	Plan        types.String `tfsdk:"plan"`
	Features    types.Set    `tfsdk:"features"`
	Limits      types.Map    `tfsdk:"limits"`
	ExpiresAt   types.Int64  `tfsdk:"expires_at"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
}

func (r *entitlementOverrideResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entitlement_override"
}

func (r *entitlementOverrideResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *entitlementOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A sales-led entitlement override: pins a plan, features and/or limits onto a user or " +
			"organization without a Stripe subscription (for example an Enterprise contract). A live override wins " +
			"over the subscription-derived entitlement from the subject's next session. Overrides are grant-only, " +
			"so every change replaces the override. Import as `<id>` or `<subject_type>/<subject_id>/<id>`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subject_type": schema.StringAttribute{
				MarkdownDescription: "`user` or `org`.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"subject_id": schema.StringAttribute{
				MarkdownDescription: "The user or organization id.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: "Plan slug to grant. At least one of `plan`, `features` or `limits` is required.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"features": schema.SetAttribute{
				MarkdownDescription: "Feature keys to grant.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
			},
			"limits": schema.MapAttribute{
				MarkdownDescription: "Limit name to number (`-1` = unlimited).",
				Optional:            true,
				ElementType:         types.Float64Type,
				PlanModifiers:       []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"expires_at": schema.Int64Attribute{
				MarkdownDescription: "When the override lapses (epoch milliseconds). Unset = never.",
				Optional:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch ms).",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *entitlementOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan entitlementOverrideModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.EntitlementOverrideWrite{
		SubjectType: plan.SubjectType.ValueString(),
		SubjectID:   plan.SubjectID.ValueString(),
		Plan:        optionalString(plan.Plan),
		Features:    setToStringSlice(ctx, plan.Features, &resp.Diagnostics),
		Limits:      floatMapToPtr(ctx, plan.Limits, &resp.Diagnostics),
		ExpiresAt:   optionalInt64(plan.ExpiresAt),
	}
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateEntitlementOverride(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create entitlement override", err.Error())
		return
	}
	mapEntitlementOverride(ctx, created, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *entitlementOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state entitlementOverrideModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	o, err := r.client.GetEntitlementOverride(ctx, state.ID.ValueString(), state.SubjectType.ValueString(), state.SubjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read entitlement override", err.Error())
		return
	}
	mapEntitlementOverride(ctx, o, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is never reached with a real change (every input forces
// replacement); it only carries computed values forward.
func (r *entitlementOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan entitlementOverrideModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *entitlementOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state entitlementOverrideModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEntitlementOverride(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to revoke entitlement override", err.Error())
	}
}

// ImportState accepts `<id>` or `<subject_type>/<subject_id>/<id>` (the latter
// narrows the list read to that subject).
func (r *entitlementOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) == 3 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subject_type"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subject_id"), parts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapEntitlementOverride(ctx context.Context, o *client.EntitlementOverride, m *entitlementOverrideModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(o.ID)
	m.SubjectType = types.StringValue(o.SubjectType)
	m.SubjectID = types.StringValue(o.SubjectID)
	m.Plan = stringPtrToValue(o.Plan)
	if o.Features == nil {
		m.Features = types.SetNull(types.StringType)
	} else {
		m.Features = stringSliceToSet(ctx, o.Features, diags)
	}
	if o.Limits == nil {
		m.Limits = types.MapNull(types.Float64Type)
	} else {
		nums := map[string]float64{}
		for k, v := range o.Limits {
			if f, ok := v.(float64); ok {
				nums[k] = f
			}
		}
		m.Limits = floatMapToMap(ctx, nums, diags)
	}
	m.ExpiresAt = int64PtrToValue(o.ExpiresAt)
	m.CreatedAt = types.Int64Value(o.CreatedAt)
}
