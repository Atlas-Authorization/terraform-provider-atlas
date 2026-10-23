package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &policyResource{}
	_ resource.ResourceWithImportState    = &policyResource{}
	_ resource.ResourceWithConfigure      = &policyResource{}
	_ resource.ResourceWithValidateConfig = &policyResource{}
)

// NewPolicyResource is the factory for atlas_policy.
func NewPolicyResource() resource.Resource { return &policyResource{} }

type policyResource struct {
	client *client.Client
}

type policyModel struct {
	ID        types.String `tfsdk:"id"`
	Key       types.String `tfsdk:"key"`
	Title     types.String `tfsdk:"title"`
	Display   types.String `tfsdk:"display"`
	Required  types.Bool   `tfsdk:"required"`
	AppliesTo types.String `tfsdk:"applies_to"`
	Reaccept  types.String `tfsdk:"reaccept"`
	GraceDays types.Int64  `tfsdk:"grace_days"`
	Status    types.String `tfsdk:"status"`
}

var (
	policyKeyRe         = regexp.MustCompile(`^[a-z0-9_-]+$`)
	policyDisplays      = map[string]bool{"checkbox": true, "scroll": true, "link": true}
	policyAppliesTo     = map[string]bool{"sign_up": true, "sign_in_and_sign_up": true}
	policyReacceptModes = map[string]bool{"none": true, "next_sign_in": true, "grace": true}
)

func (r *policyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *policyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A policy your users accept — terms of service, a privacy policy, any key. The policy " +
			"asks nobody anything until a version is published. **Versions are content**, so they are not managed " +
			"here: publish them from the dashboard, the Backend API or SDKs (`policies.versions`), or serve them from " +
			"an external system with `atlas_policy_source`. Destroying the resource deletes a policy that was never " +
			"published and *retires* one with a published version (its acceptances are kept as evidence). A retired " +
			"policy is treated as gone: the next apply re-creates it by reviving the retired policy with the same key. " +
			"Import by id or key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id (`pol_…`).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Slug, unique per instance: lowercase letters, digits, `_` and `-` (for example " +
					"`tos`, `privacy`). It becomes the `policy` of every consent record. Immutable — changing it " +
					"forces replacement.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "The name users see (up to 200 characters).",
				Required:            true,
			},
			"display": schema.StringAttribute{
				MarkdownDescription: "How the policy is presented: `checkbox` (default), `scroll` (an unskippable " +
					"reader; accepting needs scrolled-to-end) or `link` (clickwrap).",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("checkbox"),
			},
			"required": schema.BoolAttribute{
				MarkdownDescription: "Whether users must accept it to sign up. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"applies_to": schema.StringAttribute{
				MarkdownDescription: "`sign_up` (default): asked once at sign-up. `sign_in_and_sign_up`: existing " +
					"users also accept new versions, per `reaccept`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("sign_up"),
			},
			"reaccept": schema.StringAttribute{
				MarkdownDescription: "When a new version takes effect for users who accepted an earlier one: `none` " +
					"(default), `next_sign_in` (the next sign-in is held until they accept) or `grace` (they may " +
					"keep signing in for `grace_days`). Only meaningful with `applies_to = \"sign_in_and_sign_up\"`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("none"),
			},
			"grace_days": schema.Int64Attribute{
				MarkdownDescription: "The grace window in days (1-365). Required with `reaccept = \"grace\"`, and only " +
					"allowed with it.",
				Optional: true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "`active` (a retired policy is dropped from state, see above).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

// ValidateConfig checks the API's rules at plan time.
func (r *policyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var m policyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	known := func(s types.String) bool { return !s.IsNull() && !s.IsUnknown() }
	if known(m.Key) && !policyKeyRe.MatchString(m.Key.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("key"), "Invalid key",
			"key must be lowercase letters, digits, \"_\" and \"-\".")
	}
	if known(m.Title) && (m.Title.ValueString() == "" || len([]rune(m.Title.ValueString())) > 200) {
		resp.Diagnostics.AddAttributeError(path.Root("title"), "Invalid title", "title must be 1-200 characters.")
	}
	if known(m.Display) && !policyDisplays[m.Display.ValueString()] {
		resp.Diagnostics.AddAttributeError(path.Root("display"), "Invalid display", "display must be checkbox, scroll or link.")
	}
	if known(m.AppliesTo) && !policyAppliesTo[m.AppliesTo.ValueString()] {
		resp.Diagnostics.AddAttributeError(path.Root("applies_to"), "Invalid applies_to", "applies_to must be sign_up or sign_in_and_sign_up.")
	}
	if known(m.Reaccept) && !policyReacceptModes[m.Reaccept.ValueString()] {
		resp.Diagnostics.AddAttributeError(path.Root("reaccept"), "Invalid reaccept", "reaccept must be none, next_sign_in or grace.")
	}
	if m.GraceDays.IsUnknown() || m.Reaccept.IsUnknown() {
		return
	}
	grace := known(m.Reaccept) && m.Reaccept.ValueString() == "grace"
	switch {
	case grace && m.GraceDays.IsNull():
		resp.Diagnostics.AddAttributeError(path.Root("grace_days"), "grace_days is required",
			"Set grace_days (1-365) with reaccept = \"grace\".")
	case !grace && !m.GraceDays.IsNull():
		resp.Diagnostics.AddAttributeError(path.Root("grace_days"), "grace_days needs reaccept = \"grace\"",
			"grace_days only applies with reaccept = \"grace\"; remove it or change reaccept.")
	case !m.GraceDays.IsNull() && (m.GraceDays.ValueInt64() < 1 || m.GraceDays.ValueInt64() > 365):
		resp.Diagnostics.AddAttributeError(path.Root("grace_days"), "Invalid grace_days", "grace_days must be between 1 and 365.")
	}
}

func policyWriteOf(m policyModel) client.PolicyWrite {
	return client.PolicyWrite{
		Title:     m.Title.ValueString(),
		Display:   m.Display.ValueString(),
		Required:  m.Required.ValueBool(),
		AppliesTo: m.AppliesTo.ValueString(),
		Reaccept:  m.Reaccept.ValueString(),
		GraceDays: optionalInt64(m.GraceDays),
	}
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := policyWriteOf(plan)
	body.Key = plan.Key.ValueString()
	created, err := r.client.CreatePolicy(ctx, body)
	if err != nil && client.IsConflict(err) {
		// The key exists. A retired policy (destroyed earlier, or retired in the
		// dashboard) is revived with this configuration; anything else is an error.
		existing, getErr := r.client.GetPolicy(ctx, plan.Key.ValueString())
		if getErr == nil && existing.Status == "retired" {
			body.Key = ""
			body.Status = "active"
			created, err = r.client.UpdatePolicy(ctx, existing.ID, body)
		} else if getErr == nil {
			resp.Diagnostics.AddError("Policy already exists",
				fmt.Sprintf("A policy with key %q already exists (%s). Import it with `terraform import` to manage it.", plan.Key.ValueString(), existing.ID))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to create policy", err.Error())
		return
	}
	mapPolicy(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read policy", err.Error())
		return
	}
	if got.Status == "retired" {
		// Destroyed (retired) outside Terraform: the next apply revives it.
		resp.State.RemoveResource(ctx)
		return
	}
	mapPolicy(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state policyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := policyWriteOf(plan)
	body.Status = "active"
	updated, err := r.client.UpdatePolicy(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update policy", err.Error())
		return
	}
	mapPolicy(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state policyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeletePolicy(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete policy", err.Error())
	}
}

func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The API reads by id or key; Read then stores the real id.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapPolicy(p *client.Policy, m *policyModel) {
	m.ID = types.StringValue(p.ID)
	m.Key = types.StringValue(p.Key)
	m.Title = types.StringValue(p.Title)
	m.Display = types.StringValue(p.Display)
	m.Required = types.BoolValue(p.Required)
	m.AppliesTo = types.StringValue(p.AppliesTo)
	m.Reaccept = types.StringValue(p.Reaccept)
	m.GraceDays = int64PtrToValue(p.GraceDays)
	m.Status = types.StringValue(p.Status)
}
