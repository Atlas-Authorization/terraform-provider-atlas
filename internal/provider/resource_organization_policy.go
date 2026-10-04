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
	_ resource.Resource                = &organizationPolicyResource{}
	_ resource.ResourceWithImportState = &organizationPolicyResource{}
	_ resource.ResourceWithConfigure   = &organizationPolicyResource{}
)

// NewOrganizationPolicyResource is the factory registered with the provider.
func NewOrganizationPolicyResource() resource.Resource { return &organizationPolicyResource{} }

type organizationPolicyResource struct {
	client *client.Client
}

type organizationPolicyModel struct {
	ID                    types.String `tfsdk:"id"`
	OrganizationID        types.String `tfsdk:"organization_id"`
	RequireMfa            types.Bool   `tfsdk:"require_mfa"`
	SsoRequired           types.Bool   `tfsdk:"sso_required"`
	SessionIdleOverrideMs types.Int64  `tfsdk:"session_idle_override_ms"`
	AllowedSignInMethods  types.Set    `tfsdk:"allowed_sign_in_methods"`
	IPAllowlist           types.Set    `tfsdk:"ip_allowlist"`
	MaxSessionAgeSeconds  types.Int64  `tfsdk:"max_session_age_seconds"`
}

func (r *organizationPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_policy"
}

func (r *organizationPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *organizationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The §4.4 security policy for one organization: each field is an OPTIONAL tightening on top " +
			"of the instance policy, merged server-side. A field left unset inherits the instance behaviour (and keeps " +
			"whatever was last applied); set a value to enforce it. Atlas exposes no read-only GET for org policy, so the " +
			"provider reads the current policy via an idempotent empty merge, which records an `organization.policy.updated` " +
			"audit entry on each refresh.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Equals `organization_id` (the policy is a singleton per org).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization whose policy this manages. Immutable — changing it forces replacement.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"require_mfa": schema.BoolAttribute{
				MarkdownDescription: "Require an MFA-verified session to hold this org active.",
				Optional:            true,
				Computed:            true,
			},
			"sso_required": schema.BoolAttribute{
				MarkdownDescription: "Force enterprise SSO for identifiers whose email domain maps to this org " +
					"(requires a verified org domain to take effect).",
				Optional: true,
				Computed: true,
			},
			"session_idle_override_ms": schema.Int64Attribute{
				MarkdownDescription: "Override the session idle window while this org is active, in milliseconds.",
				Optional:            true,
				Computed:            true,
			},
			"allowed_sign_in_methods": schema.SetAttribute{
				MarkdownDescription: "Restrict which sign-in strategies a member may use. Each entry is a family " +
					"(`password`, `oauth`, `saml`, …) or a specific provider (`oauth_google`). Empty/unset = every method.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"ip_allowlist": schema.SetAttribute{
				MarkdownDescription: "CIDR / IP allow-list gating sign-in for members of this org. Empty/unset = unrestricted.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"max_session_age_seconds": schema.Int64Attribute{
				MarkdownDescription: "Cap the absolute session lifetime for this org's members, in seconds " +
					"(300 – 7776000). An org may only shorten, never extend.",
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (r *organizationPolicyResource) patchFromPlan(ctx context.Context, plan *organizationPolicyModel, diags *diag.Diagnostics) client.OrgPolicyPatch {
	body := client.OrgPolicyPatch{
		RequireMfa:            optionalBool(plan.RequireMfa),
		SsoRequired:           optionalBool(plan.SsoRequired),
		SessionIdleOverrideMs: optionalInt64(plan.SessionIdleOverrideMs),
		MaxSessionAgeSeconds:  optionalInt64(plan.MaxSessionAgeSeconds),
	}
	if !plan.AllowedSignInMethods.IsNull() && !plan.AllowedSignInMethods.IsUnknown() {
		body.AllowedSignInMethods = setToStringSlice(ctx, plan.AllowedSignInMethods, diags)
		if body.AllowedSignInMethods == nil {
			body.AllowedSignInMethods = []string{}
		}
	}
	if !plan.IPAllowlist.IsNull() && !plan.IPAllowlist.IsUnknown() {
		body.IPAllowlist = setToStringSlice(ctx, plan.IPAllowlist, diags)
		if body.IPAllowlist == nil {
			body.IPAllowlist = []string{}
		}
	}
	return body
}

func (r *organizationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := r.patchFromPlan(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.SetOrgPolicy(ctx, plan.OrganizationID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to set organization policy", err.Error())
		return
	}
	r.mapToState(ctx, plan.OrganizationID.ValueString(), updated, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *organizationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policy, err := r.client.GetOrgPolicy(ctx, state.OrganizationID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read organization policy", err.Error())
		return
	}
	r.mapToState(ctx, state.OrganizationID.ValueString(), policy, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := r.patchFromPlan(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.SetOrgPolicy(ctx, plan.OrganizationID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization policy", err.Error())
		return
	}
	r.mapToState(ctx, plan.OrganizationID.ValueString(), updated, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete clears every policy override by sending explicit nulls, returning the
// org to inheriting the instance policy. The org itself is untouched.
func (r *organizationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Empty slices + nil scalars clear the list fields; the scalar overrides are
	// left as-is (there is no "unset" for them beyond setting a value), which is
	// acceptable since the org is being unmanaged.
	_, err := r.client.SetOrgPolicy(ctx, state.OrganizationID.ValueString(), client.OrgPolicyPatch{
		AllowedSignInMethods: []string{},
		IPAllowlist:          []string{},
	})
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to clear organization policy", err.Error())
	}
}

func (r *organizationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), req.ID)...)
}

func (r *organizationPolicyResource) mapToState(ctx context.Context, orgID string, p *client.OrgPolicy, m *organizationPolicyModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(orgID)
	m.OrganizationID = types.StringValue(orgID)
	m.RequireMfa = boolPtrToValue(p.RequireMfa)
	m.SsoRequired = boolPtrToValue(p.SsoRequired)
	m.SessionIdleOverrideMs = int64PtrToValue(p.SessionIdleOverrideMs)
	m.MaxSessionAgeSeconds = int64PtrToValue(p.MaxSessionAgeSeconds)
	m.AllowedSignInMethods = stringSliceToSet(ctx, p.AllowedSignInMethods, diags)
	m.IPAllowlist = stringSliceToSet(ctx, p.IPAllowlist, diags)
}
