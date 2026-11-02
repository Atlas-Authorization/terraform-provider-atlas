package provider

import (
	"context"
	"encoding/json"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Identity governance resources: the entitlement catalog, separation-of-duties
// rules and birthright rules. Access requests, grants and campaigns are runtime
// records, not configuration, so they have no resource.

var (
	_ resource.Resource                = &igaEntitlementResource{}
	_ resource.ResourceWithImportState = &igaEntitlementResource{}
	_ resource.ResourceWithConfigure   = &igaEntitlementResource{}
	_ resource.Resource                = &igaSodRuleResource{}
	_ resource.ResourceWithImportState = &igaSodRuleResource{}
	_ resource.ResourceWithConfigure   = &igaSodRuleResource{}
	_ resource.Resource                = &igaBirthrightRuleResource{}
	_ resource.ResourceWithImportState = &igaBirthrightRuleResource{}
	_ resource.ResourceWithConfigure   = &igaBirthrightRuleResource{}
)

func NewIgaEntitlementResource() resource.Resource    { return &igaEntitlementResource{} }
func NewIgaSodRuleResource() resource.Resource        { return &igaSodRuleResource{} }
func NewIgaBirthrightRuleResource() resource.Resource { return &igaBirthrightRuleResource{} }

func rawOrEmpty(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage("{}")
	}
	return raw
}

func igaOptString(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return client.Ptr(v.ValueString())
}

func igaStrOrNull(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

var idAttr = schema.StringAttribute{
	MarkdownDescription: "Atlas object id.",
	Computed:            true,
	PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
}

// ───────────────────────────── atlas_iga_entitlement ─────────────────────────────

type igaEntitlementResource struct{ client *client.Client }

type igaEntitlementModel struct {
	ID                   types.String    `tfsdk:"id"`
	Key                  types.String    `tfsdk:"key"`
	Name                 types.String    `tfsdk:"name"`
	Description          types.String    `tfsdk:"description"`
	Kind                 types.String    `tfsdk:"kind"`
	Target               jsonObjectValue `tfsdk:"target"`
	OwnerUserID          types.String    `tfsdk:"owner_user_id"`
	RiskLevel            types.String    `tfsdk:"risk_level"`
	ApprovalRules        jsonObjectValue `tfsdk:"approval_rules"`
	MaxDurationSeconds   types.Int64     `tfsdk:"max_duration_seconds"`
	Requestable          types.Bool      `tfsdk:"requestable"`
	RequireJustification types.Bool      `tfsdk:"require_justification"`
}

func (r *igaEntitlementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iga_entitlement"
}

func (r *igaEntitlementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *igaEntitlementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A requestable item in the identity-governance catalog: a role, organization membership, group, resource role, " +
			"installed-app assignment or custom item, with an owner, a risk level and approval rules.",
		Attributes: map[string]schema.Attribute{
			"id": idAttr,
			"key": schema.StringAttribute{
				MarkdownDescription: "Stable key (lowercase letters, digits and `. _ : -`). Cannot be changed.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name":        schema.StringAttribute{MarkdownDescription: "Display name.", Required: true},
			"description": schema.StringAttribute{MarkdownDescription: "What the access gives.", Optional: true},
			"kind": schema.StringAttribute{
				MarkdownDescription: "`role`, `org_membership`, `group`, `resource_role`, `app_assignment` or `custom`. Cannot be changed.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"target": schema.StringAttribute{
				CustomType: jsonObjectType{},
				MarkdownDescription: "What the entitlement points at, as JSON (use `jsonencode`): `role` `{ role_key }`, `org_membership` " +
					"`{ organization_id, role }`, `group` `{ group }`, `resource_role` `{ resource_id, role }`, `app_assignment` `{ app_id }`; free-form for `custom`.",
				Required: true,
			},
			"owner_user_id": schema.StringAttribute{MarkdownDescription: "The user who owns the entitlement (an `owner` approval step and reviewers use this).", Optional: true},
			"risk_level": schema.StringAttribute{
				MarkdownDescription: "`low` (default), `medium`, `high` or `critical`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("low"),
			},
			"approval_rules": schema.StringAttribute{
				CustomType: jsonObjectType{},
				MarkdownDescription: "Approval steps as JSON: `{ steps = [{ type = \"manager\" | \"owner\" | \"group\" | \"users\", ids = [...], mode = \"any\" | \"all\", " +
					"timeout_seconds = n, escalate_to = { type = ... } }] }`. No steps means approved at once.",
				Optional: true,
			},
			"max_duration_seconds": schema.Int64Attribute{MarkdownDescription: "The longest a grant may last; omit for until revoked.", Optional: true},
			"requestable": schema.BoolAttribute{
				MarkdownDescription: "Whether users can request it themselves. Default true.",
				Optional:            true, Computed: true, Default: booldefault.StaticBool(true),
			},
			"require_justification": schema.BoolAttribute{
				MarkdownDescription: "Whether a request must state a reason. Default true.",
				Optional:            true, Computed: true, Default: booldefault.StaticBool(true),
			},
		},
	}
}

func igaEntitlementBody(m *igaEntitlementModel, diags *diag.Diagnostics) client.IgaEntitlementWrite {
	body := client.IgaEntitlementWrite{
		Key:                  m.Key.ValueString(),
		Kind:                 m.Kind.ValueString(),
		Name:                 m.Name.ValueString(),
		Description:          igaOptString(m.Description),
		Target:               jsonObjectToMap(m.Target, path.Root("target"), diags),
		OwnerUserID:          igaOptString(m.OwnerUserID),
		RiskLevel:            m.RiskLevel.ValueString(),
		ApprovalRules:        jsonObjectToMap(m.ApprovalRules, path.Root("approval_rules"), diags),
		Requestable:          m.Requestable.ValueBool(),
		RequireJustification: m.RequireJustification.ValueBool(),
	}
	if body.Target == nil {
		body.Target = map[string]interface{}{}
	}
	if body.ApprovalRules == nil {
		body.ApprovalRules = map[string]interface{}{"steps": []interface{}{}}
	}
	if !m.MaxDurationSeconds.IsNull() && !m.MaxDurationSeconds.IsUnknown() {
		body.MaxDurationSeconds = client.Ptr(m.MaxDurationSeconds.ValueInt64())
	}
	return body
}

func mapIgaEntitlement(e *client.IgaEntitlement, m *igaEntitlementModel) {
	m.ID = types.StringValue(e.ID)
	m.Key = types.StringValue(e.Key)
	m.Name = types.StringValue(e.Name)
	m.Description = igaStrOrNull(e.Description)
	m.Kind = types.StringValue(e.Kind)
	m.Target = jsonObjectFromAPI(rawOrEmpty(e.Target))
	m.OwnerUserID = igaStrOrNull(e.OwnerUserID)
	m.RiskLevel = types.StringValue(e.RiskLevel)
	// An entitlement with no steps reads back as {"steps":[]}; keep an unset attribute null.
	if !(m.ApprovalRules.IsNull() && string(e.ApprovalRules) == `{"steps":[]}`) {
		m.ApprovalRules = jsonObjectFromAPI(rawOrEmpty(e.ApprovalRules))
	}
	if e.MaxDurationSeconds == nil {
		m.MaxDurationSeconds = types.Int64Null()
	} else {
		m.MaxDurationSeconds = types.Int64Value(*e.MaxDurationSeconds)
	}
	m.Requestable = types.BoolValue(e.Requestable)
	m.RequireJustification = types.BoolValue(e.RequireJustification)
}

func (r *igaEntitlementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan igaEntitlementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := igaEntitlementBody(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.CreateIgaEntitlement(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create entitlement", err.Error())
		return
	}
	mapIgaEntitlement(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *igaEntitlementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state igaEntitlementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.GetIgaEntitlement(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read entitlement", err.Error())
		return
	}
	mapIgaEntitlement(out, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *igaEntitlementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state igaEntitlementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := igaEntitlementBody(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.UpdateIgaEntitlement(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update entitlement", err.Error())
		return
	}
	plan.ID = state.ID
	mapIgaEntitlement(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *igaEntitlementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state igaEntitlementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteIgaEntitlement(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete entitlement", err.Error())
	}
}

func (r *igaEntitlementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ───────────────────────────── atlas_iga_sod_rule ─────────────────────────────

type igaSodRuleResource struct{ client *client.Client }

type igaSodRuleModel struct {
	ID           types.String    `tfsdk:"id"`
	Name         types.String    `tfsdk:"name"`
	Description  types.String    `tfsdk:"description"`
	EntitlementA types.String    `tfsdk:"entitlement_a"`
	EntitlementB types.String    `tfsdk:"entitlement_b"`
	Mode         types.String    `tfsdk:"mode"`
	Approver     jsonObjectValue `tfsdk:"approver"`
	Enabled      types.Bool      `tfsdk:"enabled"`
}

func (r *igaSodRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iga_sod_rule"
}

func (r *igaSodRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *igaSodRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A separation-of-duties rule: two entitlements one person should not hold together. It either blocks the second one or adds an approval step to the request.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr,
			"name":        schema.StringAttribute{MarkdownDescription: "Display name.", Required: true},
			"description": schema.StringAttribute{MarkdownDescription: "Why the two conflict.", Optional: true},
			"entitlement_a": schema.StringAttribute{
				MarkdownDescription: "The id of the first entitlement (reference `atlas_iga_entitlement.x.id`). Cannot be changed.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"entitlement_b": schema.StringAttribute{
				MarkdownDescription: "The id of the second entitlement. Cannot be changed.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "`block` (default) or `require_approval`.",
				Optional:            true, Computed: true, Default: stringdefault.StaticString("block"),
			},
			"approver": schema.StringAttribute{
				CustomType:          jsonObjectType{},
				MarkdownDescription: "For `require_approval`: the extra approval step as JSON (default: the entitlement owner), e.g. `{ type = \"group\", ids = [\"risk\"], mode = \"any\" }`.",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{MarkdownDescription: "Default true.", Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		},
	}
}

func igaSodBody(m *igaSodRuleModel, diags *diag.Diagnostics) client.IgaSodRuleWrite {
	return client.IgaSodRuleWrite{
		Name:         m.Name.ValueString(),
		Description:  igaOptString(m.Description),
		EntitlementA: m.EntitlementA.ValueString(),
		EntitlementB: m.EntitlementB.ValueString(),
		Mode:         m.Mode.ValueString(),
		Approver:     jsonObjectToMap(m.Approver, path.Root("approver"), diags),
		Enabled:      m.Enabled.ValueBool(),
	}
}

func mapIgaSod(s *client.IgaSodRule, m *igaSodRuleModel) {
	m.ID = types.StringValue(s.ID)
	m.Name = types.StringValue(s.Name)
	m.Description = igaStrOrNull(s.Description)
	m.EntitlementA = types.StringValue(s.EntitlementA)
	m.EntitlementB = types.StringValue(s.EntitlementB)
	m.Mode = types.StringValue(s.Mode)
	// An unset approver stays null; the API fills the owner default only into its own view.
	if string(s.Approver) == "null" || len(s.Approver) == 0 {
		m.Approver = newJSONObjectNull()
	} else if !m.Approver.IsNull() {
		m.Approver = jsonObjectFromAPI(s.Approver)
	}
	m.Enabled = types.BoolValue(s.Enabled)
}

func (r *igaSodRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan igaSodRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := igaSodBody(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.CreateIgaSodRule(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create separation-of-duties rule", err.Error())
		return
	}
	mapIgaSod(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *igaSodRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state igaSodRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.GetIgaSodRule(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read separation-of-duties rule", err.Error())
		return
	}
	mapIgaSod(out, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *igaSodRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state igaSodRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := igaSodBody(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.UpdateIgaSodRule(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update separation-of-duties rule", err.Error())
		return
	}
	plan.ID = state.ID
	mapIgaSod(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *igaSodRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state igaSodRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteIgaSodRule(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete separation-of-duties rule", err.Error())
	}
}

func (r *igaSodRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ───────────────────────────── atlas_iga_birthright_rule ─────────────────────────────

type igaBirthrightRuleResource struct{ client *client.Client }

type igaBirthrightRuleModel struct {
	ID              types.String    `tfsdk:"id"`
	Name            types.String    `tfsdk:"name"`
	Match           jsonObjectValue `tfsdk:"match"`
	EntitlementIDs  types.List      `tfsdk:"entitlement_ids"`
	RevokeOnUnmatch types.Bool      `tfsdk:"revoke_on_unmatch"`
	Enabled         types.Bool      `tfsdk:"enabled"`
}

func (r *igaBirthrightRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iga_birthright_rule"
}

func (r *igaBirthrightRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *igaBirthrightRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A joiner/mover rule: everyone matching `match` holds the listed entitlements, and loses them when they stop matching.",
		Attributes: map[string]schema.Attribute{
			"id":   idAttr,
			"name": schema.StringAttribute{MarkdownDescription: "Display name.", Required: true},
			"match": schema.StringAttribute{
				CustomType:          jsonObjectType{},
				MarkdownDescription: "Conditions as JSON; every one given must hold: `group` (slug), `organization_id` with optional `org_role`, `app_role`, `email_domain` (verified).",
				Required:            true,
			},
			"entitlement_ids": schema.ListAttribute{MarkdownDescription: "Ids of the entitlements the rule gives.", ElementType: types.StringType, Required: true},
			"revoke_on_unmatch": schema.BoolAttribute{
				MarkdownDescription: "Take the access back when a person stops matching. Default true.",
				Optional:            true, Computed: true, Default: booldefault.StaticBool(true),
			},
			"enabled": schema.BoolAttribute{MarkdownDescription: "Default true.", Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		},
	}
}

func igaBirthrightBody(ctx context.Context, m *igaBirthrightRuleModel, diags *diag.Diagnostics) client.IgaBirthrightRuleWrite {
	var ids []string
	diags.Append(m.EntitlementIDs.ElementsAs(ctx, &ids, false)...)
	match := jsonObjectToMap(m.Match, path.Root("match"), diags)
	if match == nil {
		match = map[string]interface{}{}
	}
	return client.IgaBirthrightRuleWrite{Name: m.Name.ValueString(), Match: match, EntitlementIDs: ids, RevokeOnUnmatch: m.RevokeOnUnmatch.ValueBool(), Enabled: m.Enabled.ValueBool()}
}

func mapIgaBirthright(b *client.IgaBirthrightRule, m *igaBirthrightRuleModel) {
	m.ID = types.StringValue(b.ID)
	m.Name = types.StringValue(b.Name)
	m.Match = jsonObjectFromAPI(rawOrEmpty(b.Match))
	elems := make([]types.String, 0, len(b.EntitlementIDs))
	for _, id := range b.EntitlementIDs {
		elems = append(elems, types.StringValue(id))
	}
	list, _ := types.ListValueFrom(context.Background(), types.StringType, elems)
	m.EntitlementIDs = list
	m.RevokeOnUnmatch = types.BoolValue(b.RevokeOnUnmatch)
	m.Enabled = types.BoolValue(b.Enabled)
}

func (r *igaBirthrightRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan igaBirthrightRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := igaBirthrightBody(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.CreateIgaBirthrightRule(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create birthright rule", err.Error())
		return
	}
	mapIgaBirthright(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *igaBirthrightRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state igaBirthrightRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.GetIgaBirthrightRule(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read birthright rule", err.Error())
		return
	}
	mapIgaBirthright(out, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *igaBirthrightRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state igaBirthrightRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := igaBirthrightBody(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.UpdateIgaBirthrightRule(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update birthright rule", err.Error())
		return
	}
	plan.ID = state.ID
	mapIgaBirthright(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *igaBirthrightRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state igaBirthrightRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteIgaBirthrightRule(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete birthright rule", err.Error())
	}
}

func (r *igaBirthrightRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
