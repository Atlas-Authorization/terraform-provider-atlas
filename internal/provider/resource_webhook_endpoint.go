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
	_ resource.Resource                = &webhookEndpointResource{}
	_ resource.ResourceWithImportState = &webhookEndpointResource{}
	_ resource.ResourceWithConfigure   = &webhookEndpointResource{}
)

// NewWebhookEndpointResource is the factory registered with the provider.
func NewWebhookEndpointResource() resource.Resource { return &webhookEndpointResource{} }

type webhookEndpointResource struct {
	client *client.Client
}

type webhookEndpointModel struct {
	ID            types.String `tfsdk:"id"`
	URL           types.String `tfsdk:"url"`
	EnabledEvents types.List   `tfsdk:"enabled_events"`
	Active        types.Bool   `tfsdk:"active"`
	Secret        types.String `tfsdk:"secret"`
	DisabledAt    types.Int64  `tfsdk:"disabled_at"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
}

func (r *webhookEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_endpoint"
}

func (r *webhookEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *webhookEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A webhook endpoint that receives signed Atlas events. The signing secret (`whsec_...`) is " +
			"revealed once, on create, and stored (sensitive) in state. Changing `url` or `enabled_events` updates the " +
			"endpoint IN PLACE (via the Backend API PATCH route) and PRESERVES the signing secret, so deliveries keep " +
			"verifying without a re-subscription.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "HTTPS URL that receives event deliveries. Updatable in place (the secret is preserved).",
				Required:            true,
			},
			"enabled_events": schema.ListAttribute{
				MarkdownDescription: "Event types to deliver, or [\"*\"] for all. Updatable in place (the secret is preserved).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "The signing secret, revealed only on create.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the endpoint is currently active.",
				Computed:            true,
			},
			"disabled_at": schema.Int64Attribute{
				MarkdownDescription: "When the endpoint was disabled (epoch ms), if it has been.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch ms).",
				Computed:            true,
			},
		},
	}
}

func (r *webhookEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookEndpointModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	events := listToStringSlice(ctx, plan.EnabledEvents, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateWebhookEndpoint(ctx, plan.URL.ValueString(), events)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create webhook endpoint", err.Error())
		return
	}
	r.mapToState(ctx, created, &plan, &resp.Diagnostics, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookEndpointModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetWebhookEndpoint(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read webhook endpoint", err.Error())
		return
	}
	// A list read never re-reveals the secret; keep the create-time value.
	r.mapToState(ctx, fetched, &state, &resp.Diagnostics, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update patches url/enabled_events in place via the Backend API PATCH route,
// which does NOT rotate the signing secret. The PATCH response never returns the
// secret, so it is carried forward from prior state.
func (r *webhookEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state webhookEndpointModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	events := listToStringSlice(ctx, plan.EnabledEvents, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	patch := client.WebhookEndpointUpdate{
		URL:           optionalString(plan.URL),
		EnabledEvents: events,
		Active:        optionalBool(plan.Active),
	}
	updated, err := r.client.UpdateWebhookEndpoint(ctx, state.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update webhook endpoint", err.Error())
		return
	}
	// The PATCH response omits the secret (it is not rotated by an update); map
	// every other field from the response and carry the create-time secret
	// forward from prior state.
	r.mapToState(ctx, updated, &plan, &resp.Diagnostics, false)
	plan.Secret = state.Secret
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookEndpointModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWebhookEndpoint(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete webhook endpoint", err.Error())
	}
}

func (r *webhookEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *webhookEndpointResource) mapToState(ctx context.Context, e *client.WebhookEndpoint, m *webhookEndpointModel, diags *diag.Diagnostics, withSecret bool) {
	m.ID = types.StringValue(e.ID)
	m.URL = types.StringValue(e.URL)
	m.EnabledEvents = stringSliceToList(ctx, e.EnabledEvents, diags)
	m.Active = types.BoolValue(e.Active)
	m.DisabledAt = int64PtrToValue(e.DisabledAt)
	m.CreatedAt = types.Int64Value(e.CreatedAt)
	if withSecret {
		m.Secret = stringPtrToValue(e.Secret)
	}
}
