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
	_ resource.Resource                = &policySettingsResource{}
	_ resource.ResourceWithImportState = &policySettingsResource{}
	_ resource.ResourceWithConfigure   = &policySettingsResource{}
)

// NewPolicySettingsResource is the factory for atlas_policy_settings.
func NewPolicySettingsResource() resource.Resource { return &policySettingsResource{} }

type policySettingsResource struct {
	client *client.Client
}

type policySettingsModel struct {
	ID           types.String `tfsdk:"id"`
	SessionClaim types.Bool   `tfsdk:"session_claim"`
}

func (r *policySettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_settings"
}

func (r *policySettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *policySettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Instance-wide policy acceptance settings — a singleton per instance. Destroying the " +
			"resource turns the `pol` session claim off. Import with any id, e.g. `policy_settings`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `policy_settings`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"session_claim": schema.BoolAttribute{
				MarkdownDescription: "Put the `pol` claim — the accepted version of every required policy, for example " +
					"`{\"tos\": \"2026-10-24\"}` — on session tokens, so your backend can check acceptance without a " +
					"lookup. Tokens minted after the change carry it; it can take up to 30 seconds to apply everywhere.",
				Required: true,
			},
		},
	}
}

func (r *policySettingsResource) write(ctx context.Context, plan *policySettingsModel) error {
	out, err := r.client.SetPolicySettings(ctx, plan.SessionClaim.ValueBool())
	if err != nil {
		return err
	}
	plan.ID = types.StringValue("policy_settings")
	plan.SessionClaim = types.BoolValue(out.SessionClaim)
	return nil
}

func (r *policySettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.write(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to set the policy settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policySettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policySettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.GetPolicySettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the policy settings", err.Error())
		return
	}
	state.ID = types.StringValue("policy_settings")
	state.SessionClaim = types.BoolValue(out.SessionClaim)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *policySettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan policySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.write(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to update the policy settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policySettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if _, err := r.client.SetPolicySettings(ctx, false); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to reset the policy settings", err.Error())
	}
}

func (r *policySettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
