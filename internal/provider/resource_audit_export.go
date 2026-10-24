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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &auditExportResource{}
	_ resource.ResourceWithImportState = &auditExportResource{}
	_ resource.ResourceWithConfigure   = &auditExportResource{}
	_ resource.ResourceWithModifyPlan  = &auditExportResource{}
)

// NewAuditExportResource is the factory for atlas_audit_export.
func NewAuditExportResource() resource.Resource { return &auditExportResource{} }

type auditExportResource struct {
	client *client.Client
}

type auditExportModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Type                types.String `tfsdk:"type"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	ActionFilter        types.Set    `tfsdk:"action_filter"`
	Destination         types.Map    `tfsdk:"destination"`
	SecretAccessKey     types.String `tfsdk:"secret_access_key"`
	BackfillHours       types.Int64  `tfsdk:"backfill_hours"`
	RotateSecretTrigger types.String `tfsdk:"rotate_secret_trigger"`
	Secret              types.String `tfsdk:"secret"`
	Status              types.String `tfsdk:"status"`
	LastError           types.String `tfsdk:"last_error"`
	LastDeliveredAt     types.Int64  `tfsdk:"last_delivered_at"`
	ExportedCount       types.Int64  `tfsdk:"exported_count"`
	CreatedAt           types.Int64  `tfsdk:"created_at"`
}

func (r *auditExportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_export"
}

func (r *auditExportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func auditExportSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "An audit export streams the instance audit log, as CloudEvents 1.0, to a sink of your own: a " +
			"signed `webhook` (batches of events) or an `s3` bucket (any S3-compatible store; one object per batch). An " +
			"instance can have up to 5. Key lifecycle and data-rights events are always exported, whatever `action_filter` " +
			"says. For a webhook export Atlas generates the signing `secret` (shown once, kept sensitive in state); rotate " +
			"it with `rotate_secret_trigger`. Changing `type` replaces the export; destroying it deletes it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A label (up to 100 characters).",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "`webhook` or `s3`. Changing it replaces the export.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the export runs. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"action_filter": schema.SetAttribute{
				MarkdownDescription: "Audit actions or prefixes to export (e.g. `user.`, `organization.*`; up to 50). Unset exports every action.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"destination": schema.MapAttribute{
				MarkdownDescription: "Where the events go. `webhook`: `url`. `s3`: `endpoint` (https), `bucket`, `region`, " +
					"`prefix`, `access_key_id`. The S3 secret key is the separate `secret_access_key` attribute.",
				Required:    true,
				ElementType: types.StringType,
			},
			"secret_access_key": schema.StringAttribute{
				MarkdownDescription: "`s3` only: the secret access key for `destination.access_key_id`. Sent to Atlas, never read back.",
				Optional:            true,
				Sensitive:           true,
			},
			"backfill_hours": schema.Int64Attribute{
				MarkdownDescription: "On create, also export the last N hours of the audit log (0-720). Default 0: only new entries.",
				Optional:            true,
			},
			"rotate_secret_trigger": schema.StringAttribute{
				MarkdownDescription: "`webhook` only: any value; changing it rotates the signing secret in place.",
				Optional:            true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "`webhook` only: the signing secret that verifies the `Atlas-Signature` header on each batch. Shown once.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The export's state, e.g. `active` or `failing`.",
				Computed:            true,
			},
			"last_error": schema.StringAttribute{
				MarkdownDescription: "Why the last delivery failed, if it did.",
				Computed:            true,
			},
			"last_delivered_at": schema.Int64Attribute{
				MarkdownDescription: "When a batch last reached the sink (epoch ms).",
				Computed:            true,
			},
			"exported_count": schema.Int64Attribute{
				MarkdownDescription: "Events exported so far.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch ms).",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *auditExportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = auditExportSchema()
}

// auditDestinationFromAPI keeps the keys the configuration names (the API also
// echoes has_* flags, and normalises prefix to end in a slash).
func auditDestinationFromAPI(ctx context.Context, d map[string]interface{}, prior types.Map, diags *diag.Diagnostics) types.Map {
	all := map[string]string{}
	for k, v := range d {
		if s, ok := v.(string); ok {
			all[k] = s
		}
	}
	if prior.IsNull() || prior.IsUnknown() {
		return stringMapToMap(ctx, all, diags)
	}
	priorVals := mapToStringMap(ctx, prior, diags)
	kept := map[string]string{}
	for k, want := range priorVals {
		got, ok := all[k]
		if !ok {
			continue
		}
		if k == "prefix" && want+"/" == got {
			got = want
		}
		kept[k] = got
	}
	return stringMapToMap(ctx, kept, diags)
}

func fillAuditExport(ctx context.Context, m *auditExportModel, e *client.AuditExport, diags *diag.Diagnostics) {
	m.ID = types.StringValue(e.ID)
	m.Name = types.StringValue(e.Name)
	m.Type = types.StringValue(e.Type)
	m.Enabled = types.BoolValue(e.Enabled)
	if e.ActionFilter == nil {
		m.ActionFilter = types.SetNull(types.StringType)
	} else {
		m.ActionFilter = stringSliceToSet(ctx, e.ActionFilter, diags)
	}
	m.Destination = auditDestinationFromAPI(ctx, e.Destination, m.Destination, diags)
	if e.Secret != nil {
		m.Secret = types.StringValue(*e.Secret)
	} else if m.Secret.IsUnknown() {
		m.Secret = types.StringNull()
	}
	m.Status = types.StringValue(e.Status)
	m.LastError = stringPtrToValue(e.LastError)
	m.LastDeliveredAt = int64PtrToValue(e.LastDeliveredAt)
	m.ExportedCount = types.Int64Value(e.ExportedCount)
	m.CreatedAt = types.Int64Value(e.CreatedAt)
}

func (r *auditExportResource) destination(ctx context.Context, m *auditExportModel, diags *diag.Diagnostics) map[string]string {
	dest := mapToStringMap(ctx, m.Destination, diags)
	if dest == nil {
		dest = map[string]string{}
	}
	if !isBlank(m.SecretAccessKey) {
		dest["secret_access_key"] = m.SecretAccessKey.ValueString()
	}
	return dest
}

func (r *auditExportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan auditExportModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.AuditExportWrite{
		Name:          optionalString(plan.Name),
		Type:          optionalString(plan.Type),
		Destination:   r.destination(ctx, &plan, &resp.Diagnostics),
		Enabled:       optionalBool(plan.Enabled),
		BackfillHours: optionalInt64(plan.BackfillHours),
	}
	if !plan.ActionFilter.IsNull() && !plan.ActionFilter.IsUnknown() {
		filter := setToStringSlice(ctx, plan.ActionFilter, &resp.Diagnostics)
		if filter == nil {
			filter = []string{}
		}
		body.ActionFilter = filter
	}
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateAuditExport(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create audit export", err.Error())
		return
	}
	fillAuditExport(ctx, &plan, created, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *auditExportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state auditExportModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	e, err := r.client.GetAuditExport(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read audit export", err.Error())
		return
	}
	fillAuditExport(ctx, &state, e, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *auditExportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state auditExportModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.AuditExportWrite{
		Name:        optionalString(plan.Name),
		Destination: r.destination(ctx, &plan, &resp.Diagnostics),
		Enabled:     optionalBool(plan.Enabled),
	}
	if plan.ActionFilter.IsNull() {
		// Unset means every action: the API clears a filter with null.
		body.ActionFilter = json.RawMessage("null")
	} else {
		filter := setToStringSlice(ctx, plan.ActionFilter, &resp.Diagnostics)
		if filter == nil {
			filter = []string{}
		}
		body.ActionFilter = filter
	}
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateAuditExport(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update audit export", err.Error())
		return
	}
	plan.Secret = state.Secret
	if !plan.RotateSecretTrigger.Equal(state.RotateSecretTrigger) && plan.Type.ValueString() == "webhook" {
		rotated, err := r.client.RotateAuditExportSecret(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to rotate audit export secret", err.Error())
			return
		}
		updated.Secret = rotated.Secret
	}
	fillAuditExport(ctx, &plan, updated, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *auditExportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state auditExportModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAuditExport(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete audit export", err.Error())
	}
}

func (r *auditExportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ModifyPlan plans a new signing secret when rotate_secret_trigger changes.
func (r *auditExportResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	planSecretRotation(ctx, req, resp, "secret")
}
