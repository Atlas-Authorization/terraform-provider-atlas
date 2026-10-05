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
	_ resource.Resource                = &orgSettingsSchemaResource{}
	_ resource.ResourceWithImportState = &orgSettingsSchemaResource{}
	_ resource.ResourceWithConfigure   = &orgSettingsSchemaResource{}
)

// orgSettingsSchemaID is the fixed id of the singleton registry entry.
const orgSettingsSchemaID = "organization_settings_schema"

// NewOrgSettingsSchemaResource is the factory registered with the provider.
func NewOrgSettingsSchemaResource() resource.Resource { return &orgSettingsSchemaResource{} }

type orgSettingsSchemaResource struct {
	client *client.Client
}

type orgSettingsSchemaModel struct {
	ID        types.String `tfsdk:"id"`
	Schema    types.String `tfsdk:"schema"`
	Version   types.Int64  `tfsdk:"version"`
	UpdatedAt types.Int64  `tfsdk:"updated_at"`
}

func (r *orgSettingsSchemaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_settings_schema"
}

func (r *orgSettingsSchemaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *orgSettingsSchemaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The per-instance org-settings JSON Schema registry. An instance registers ONE JSON Schema " +
			"describing the shape of its organization-level settings; Atlas then validates every per-organization " +
			"settings write against it, versions and audits it. This resource is a singleton per instance. Each change " +
			"re-registers the schema and bumps `version`. Atlas exposes no delete for the registry, so destroying this " +
			"resource only removes it from Terraform state; the registered schema stays in place.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `organization_settings_schema` (the registry is a singleton per instance).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "The JSON Schema document as a JSON string (use `jsonencode(...)`). Must be a JSON " +
					"object describing the org settings shape, e.g. `{\"type\":\"object\",\"properties\":{...}}`.",
				Required: true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Schema version, incremented by Atlas on every registration.",
				Computed:            true,
			},
			"updated_at": schema.Int64Attribute{
				MarkdownDescription: "Last registration time (epoch ms).",
				Computed:            true,
			},
		},
	}
}

func (r *orgSettingsSchemaResource) put(ctx context.Context, plan *orgSettingsSchemaModel, action string, diags *diag.Diagnostics) {
	raw, err := stringToJSONRaw(plan.Schema)
	if err != nil {
		diags.AddAttributeError(path.Root("schema"), "Invalid JSON", err.Error())
		return
	}
	if raw == nil {
		diags.AddAttributeError(path.Root("schema"), "Schema is empty", "schema must be a JSON Schema object.")
		return
	}
	out, err := r.client.PutOrgSettingsSchema(ctx, raw)
	if err != nil {
		diags.AddError("Unable to "+action+" org settings schema", err.Error())
		return
	}
	r.mapToState(out, plan)
}

func (r *orgSettingsSchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan orgSettingsSchemaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.put(ctx, &plan, "register", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *orgSettingsSchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state orgSettingsSchemaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.GetOrgSettingsSchema(ctx)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read org settings schema", err.Error())
		return
	}
	r.mapToState(out, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *orgSettingsSchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan orgSettingsSchemaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.put(ctx, &plan, "update", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete only forgets the resource: the API has no way to unregister a schema.
func (r *orgSettingsSchemaResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *orgSettingsSchemaResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The registry is a singleton; any import id resolves to it.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), orgSettingsSchemaID)...)
}

// mapToState keeps the configured document verbatim when it is semantically
// equal to what the server holds (whitespace/key order), so a round-trip never
// shows a spurious diff; real drift is surfaced as the server's canonical JSON.
func (r *orgSettingsSchemaResource) mapToState(s *client.OrgSettingsSchema, m *orgSettingsSchemaModel) {
	m.ID = types.StringValue(orgSettingsSchemaID)
	m.Version = types.Int64Value(s.Version)
	m.UpdatedAt = types.Int64Value(s.UpdatedAt)
	server, err := canonicalJSON(s.Schema)
	if err != nil {
		m.Schema = types.StringValue(string(s.Schema))
		return
	}
	if !m.Schema.IsNull() && !m.Schema.IsUnknown() {
		if cur, cerr := canonicalJSON([]byte(m.Schema.ValueString())); cerr == nil && cur == server {
			return
		}
	}
	m.Schema = types.StringValue(server)
}
