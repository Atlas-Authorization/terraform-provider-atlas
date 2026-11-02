package provider

import (
	"context"
	"encoding/json"
	"strings"

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

var (
	_ resource.Resource                = &mcpGatewayResource{}
	_ resource.ResourceWithImportState = &mcpGatewayResource{}
	_ resource.ResourceWithConfigure   = &mcpGatewayResource{}
)

// NewMcpGatewayResource is the factory registered with the provider.
func NewMcpGatewayResource() resource.Resource { return &mcpGatewayResource{} }

type mcpGatewayResource struct {
	client *client.Client
}

type mcpGatewayModel struct {
	ID                       types.String    `tfsdk:"id"`
	Name                     types.String    `tfsdk:"name"`
	UpstreamURL              types.String    `tfsdk:"upstream_url"`
	UpstreamAuthorization    types.String    `tfsdk:"upstream_authorization"`
	HasUpstreamAuthorization types.Bool      `tfsdk:"has_upstream_authorization"`
	DefaultPolicy            types.String    `tfsdk:"default_policy"`
	Tools                    jsonObjectValue `tfsdk:"tools"`
	Enabled                  types.Bool      `tfsdk:"enabled"`
	URL                      types.String    `tfsdk:"url"`
}

func (r *mcpGatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_gateway"
}

func (r *mcpGatewayResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *mcpGatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An MCP tool gateway: one URL that fronts your MCP server for Atlas agents. The agent presents its " +
			"agent token; the gateway lists only the tools it may use and enforces per-tool scopes, argument conditions, " +
			"human approval and response masking on each call, and audits every call.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name (1 to 100 characters); also the server name agents see.",
				Required:            true,
			},
			"upstream_url": schema.StringAttribute{
				MarkdownDescription: "The http(s) URL of your MCP server (streamable HTTP).",
				Required:            true,
			},
			"upstream_authorization": schema.StringAttribute{
				MarkdownDescription: "The `Authorization` header value the gateway sends to your MCP server. Write-only: stored " +
					"encrypted and never read back, so Terraform cannot see a change made outside it. Remove it to clear it.",
				Optional:  true,
				Sensitive: true,
			},
			"has_upstream_authorization": schema.BoolAttribute{
				MarkdownDescription: "Whether an upstream authorization is stored.",
				Computed:            true,
			},
			"default_policy": schema.StringAttribute{
				MarkdownDescription: "`deny` (default): only tools listed in `tools` are reachable. `allow`: every upstream tool is, unless its policy says otherwise.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("deny"),
			},
			"tools": schema.StringAttribute{
				CustomType: jsonObjectType{},
				MarkdownDescription: "Per-tool policy as a JSON object keyed by tool name (use `jsonencode`): `allow`, `scopes` " +
					"(agent scopes the call needs), `conditions` (`{ arg, op, value }` with `eq`, `neq`, `in`, `not_in`, `lte`, " +
					"`gte`, `present`), `approval` (a person approves through an access request) and `mask` (`fields`, `presets`). " +
					"Compared semantically; replaced as a whole on each update.",
				Optional: true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the gateway answers. Defaults to true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The gateway URL agents connect to.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func mcpGatewayBody(plan *mcpGatewayModel, diags *diag.Diagnostics) client.McpGatewayWrite {
	body := client.McpGatewayWrite{
		Name:          plan.Name.ValueString(),
		UpstreamURL:   plan.UpstreamURL.ValueString(),
		DefaultPolicy: plan.DefaultPolicy.ValueString(),
		Tools:         jsonObjectToMap(plan.Tools, path.Root("tools"), diags),
		Enabled:       client.Ptr(plan.Enabled.ValueBool()),
	}
	if body.Tools == nil {
		body.Tools = map[string]interface{}{}
	}
	if !plan.UpstreamAuthorization.IsNull() && !plan.UpstreamAuthorization.IsUnknown() {
		body.UpstreamAuthorization = client.Ptr(plan.UpstreamAuthorization.ValueString())
	}
	return body
}

func (r *mcpGatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpGatewayModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := mcpGatewayBody(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateMcpGateway(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create MCP gateway", err.Error())
		return
	}
	mapMcpGateway(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mcpGatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpGatewayModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetMcpGateway(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read MCP gateway", err.Error())
		return
	}
	mapMcpGateway(fetched, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mcpGatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state mcpGatewayModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := mcpGatewayBody(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateMcpGateway(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update MCP gateway", err.Error())
		return
	}
	plan.ID = state.ID
	mapMcpGateway(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mcpGatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpGatewayModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteMcpGateway(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete MCP gateway", err.Error())
	}
}

func (r *mcpGatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapMcpGateway copies the API view onto the model. upstream_authorization is
// write-only, so the model keeps the configured value.
func mapMcpGateway(g *client.McpGateway, m *mcpGatewayModel) {
	m.ID = types.StringValue(g.ID)
	m.Name = types.StringValue(g.Name)
	// The API normalises the URL (a bare origin gains a trailing slash): keep the configured spelling when it is the same URL.
	if m.UpstreamURL.IsNull() || m.UpstreamURL.IsUnknown() || strings.TrimSuffix(m.UpstreamURL.ValueString(), "/") != strings.TrimSuffix(g.UpstreamURL, "/") {
		m.UpstreamURL = types.StringValue(g.UpstreamURL)
	}
	m.HasUpstreamAuthorization = types.BoolValue(g.HasUpstreamAuthorization)
	m.DefaultPolicy = types.StringValue(g.DefaultPolicy)
	m.Enabled = types.BoolValue(g.Enabled)
	m.URL = types.StringValue(g.URL)
	tools := g.Tools
	if len(tools) == 0 {
		tools = json.RawMessage("{}")
	}
	m.Tools = jsonObjectFromAPI(tools)
}
