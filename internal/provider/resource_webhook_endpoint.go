package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &webhookEndpointResource{}
	_ resource.ResourceWithImportState    = &webhookEndpointResource{}
	_ resource.ResourceWithConfigure      = &webhookEndpointResource{}
	_ resource.ResourceWithModifyPlan     = &webhookEndpointResource{}
	_ resource.ResourceWithValidateConfig = &webhookEndpointResource{}
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
	RotateTrigger types.String `tfsdk:"rotate_secret_trigger"`
	FailingSince  types.Int64  `tfsdk:"failing_since"`
	DisabledAt    types.Int64  `tfsdk:"disabled_at"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`

	Description         types.String `tfsdk:"description"`
	EventNaming         types.String `tfsdk:"event_naming"`
	External            types.Bool   `tfsdk:"external"`
	SignatureVersions   types.Set    `tfsdk:"signature_versions"`
	AlertContacts       types.Set    `tfsdk:"alert_contacts"`
	Destination         types.Map    `tfsdk:"destination"`
	DestinationCreds    types.Map    `tfsdk:"destination_credentials"`
	DestinationCredsSet types.Bool   `tfsdk:"destination_credentials_set"`
	ExternalRef         types.String `tfsdk:"external_ref"`
	Metadata            types.Map    `tfsdk:"metadata"`
	Health              types.String `tfsdk:"health"`
	LastSuccessAt       types.Int64  `tfsdk:"last_success_at"`
	ConsecutiveFailures types.Int64  `tfsdk:"consecutive_failures"`
	LastFailingSince    types.Int64  `tfsdk:"last_failing_since"`
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
			"stored (sensitive) in state: it is returned on create, and read back from Atlas on import. Changing `url`, " +
			"`enabled_events` or `active` updates the endpoint IN PLACE and PRESERVES the signing secret, so deliveries " +
			"keep verifying. Change `rotate_secret_trigger` to rotate the secret in place. The receiver is an HTTPS " +
			"`url`, or an event `destination` (Amazon EventBridge, Google Cloud Pub/Sub or Cloudflare Queues) with " +
			"your own `destination_credentials`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "HTTPS URL that receives event deliveries. Updatable in place (the secret is preserved). " +
					"Leave it out when you set `destination`.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled_events": schema.ListAttribute{
				MarkdownDescription: "Event types to deliver, or [\"*\"] for all. Updatable in place (the secret is preserved).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "The signing secret. Returned on create and on rotation; on import it is read back " +
					"from Atlas (an audited reveal).",
				Computed:      true,
				Sensitive:     true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"rotate_secret_trigger": schema.StringAttribute{
				MarkdownDescription: "Any value; changing it rotates the signing secret in place (the new one lands in " +
					"`secret`). Deliveries signed with the old secret stop verifying at once, so update your verifier.",
				Optional: true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the endpoint receives deliveries. Atlas switches an endpoint off after " +
					"repeated failures; set `true` to switch it back on (or `false` to pause it).",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"failing_since": schema.Int64Attribute{
				MarkdownDescription: "Start of the current run of failed deliveries (epoch ms), if any.",
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
			"description": schema.StringAttribute{
				MarkdownDescription: "A note on what the endpoint is for.",
				Optional:            true,
			},
			"event_naming": schema.StringAttribute{
				MarkdownDescription: "`canonical` (the default for a new endpoint) delivers the canonical event names; " +
					"`legacy` delivers the deprecated aliases. An endpoint that existed before the canonical names keeps `legacy` until you change it.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"external": schema.BoolAttribute{
				MarkdownDescription: "Marks an endpoint that belongs to a third party (a partner you forward events to).",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"signature_versions": schema.SetAttribute{
				MarkdownDescription: "Signature versions sent on each delivery: `v1` (HMAC over `id.ts.body`) and/or `v2` " +
					"(HMAC over `id.ts.url.body`). Verifiers ignore versions they do not know.",
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"alert_contacts": schema.SetAttribute{
				MarkdownDescription: "Emails told when the endpoint starts failing or is disabled. Unset means the owners of the account.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"destination": schema.MapAttribute{
				MarkdownDescription: "An event destination instead of an HTTPS `url`: a map with `type` (`aws_eventbridge`, " +
					"`gcp_pubsub` or `cloudflare_queue`) and its settings: `region`, `event_bus_name`, `source` (EventBridge); " +
					"`project_id`, `topic_id` (Pub/Sub); `account_id`, `queue_id` (Cloudflare). Needs `destination_credentials`.",
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			},
			"destination_credentials": schema.MapAttribute{
				MarkdownDescription: "The credentials for `destination`: `access_key_id`, `secret_access_key`, `session_token` " +
					"(EventBridge); `service_account_json` (Pub/Sub); `api_token` (Cloudflare). Sent to Atlas, never read back; " +
					"`destination_credentials_set` says whether Atlas holds them.",
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"destination_credentials_set": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas holds credentials for the event destination.",
				Computed:            true,
			},
			"external_ref": schema.StringAttribute{
				MarkdownDescription: "Your own stable id for the endpoint. A second create with the same ref returns the existing endpoint; changing it replaces the endpoint.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Tags on the endpoint (lowercase keys; string values).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringMap(),
			},
			"health": schema.StringAttribute{
				MarkdownDescription: "`healthy`, `failing` or `disabled`.",
				Computed:            true,
			},
			"last_success_at": schema.Int64Attribute{
				MarkdownDescription: "When a delivery last succeeded (epoch ms), if ever.",
				Computed:            true,
			},
			"consecutive_failures": schema.Int64Attribute{
				MarkdownDescription: "Deliveries failed in a row.",
				Computed:            true,
			},
			"last_failing_since": schema.Int64Attribute{
				MarkdownDescription: "Start of the most recent run of failed deliveries (epoch ms), kept after recovery " +
					"so a replay can start from it.",
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
	}
}

// ValidateConfig requires a receiver: an HTTPS url or an event destination.
func (r *webhookEndpointResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg webhookEndpointModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if cfg.URL.IsUnknown() || cfg.Destination.IsUnknown() {
		return
	}
	if cfg.URL.IsNull() && cfg.Destination.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("url"), "Missing receiver",
			"Set url (an HTTPS endpoint) or destination (an event destination with destination_credentials).")
	}
	if !cfg.URL.IsNull() && !cfg.Destination.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("destination"), "Conflicting receivers",
			"Set either url or destination, not both.")
	}
}

// writeBody builds the create / PATCH body. Fields Atlas fills in itself come
// from the configuration, so an omitted value keeps Atlas's default; the rest
// come from the plan. update trims the body to what changed.
func (r *webhookEndpointResource) writeBody(ctx context.Context, plan, cfg *webhookEndpointModel, state *webhookEndpointModel, diags *diag.Diagnostics) client.WebhookEndpointWrite {
	body := client.WebhookEndpointWrite{
		URL:           optionalString(cfg.URL),
		EnabledEvents: listToStringSlice(ctx, plan.EnabledEvents, diags),
		Active:        nil,
		Description:   optionalString(cfg.Description),
		EventNaming:   optionalString(cfg.EventNaming),
		External:      optionalBool(cfg.External),
		Metadata:      metadataToAPI(ctx, plan.Metadata, diags),
		ExternalRef:   optionalString(cfg.ExternalRef),
	}
	if !cfg.SignatureVersions.IsNull() && !cfg.SignatureVersions.IsUnknown() {
		body.SignatureVersions = setToStringSlice(ctx, cfg.SignatureVersions, diags)
	}
	if !cfg.AlertContacts.IsNull() && !cfg.AlertContacts.IsUnknown() {
		contacts := setToStringSlice(ctx, cfg.AlertContacts, diags)
		if contacts == nil {
			contacts = []string{}
		}
		body.AlertContacts = &contacts
	}
	if !cfg.Destination.IsNull() && !cfg.Destination.IsUnknown() {
		dest := map[string]interface{}{}
		for k, v := range mapToStringMap(ctx, cfg.Destination, diags) {
			dest[k] = v
		}
		body.Destination = dest
	}
	if !cfg.DestinationCreds.IsNull() && !cfg.DestinationCreds.IsUnknown() {
		creds := map[string]interface{}{}
		for k, v := range mapToStringMap(ctx, cfg.DestinationCreds, diags) {
			creds[k] = v
		}
		body.DestinationCredentials = creds
	}
	if state != nil {
		// A PATCH sends only what the plan changes.
		if plan.URL.Equal(state.URL) {
			body.URL = nil
		}
		body.ExternalRef = nil
		if body.Destination != nil && cfg.Destination.Equal(state.Destination) && body.DestinationCredentials == nil {
			body.Destination = nil
		}
	}
	return body
}

func (r *webhookEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, cfg webhookEndpointModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := r.writeBody(ctx, &plan, &cfg, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateWebhookEndpointFull(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create webhook endpoint", err.Error())
		return
	}
	// A new endpoint starts active; pause it when the configuration says so.
	if !plan.Active.IsNull() && !plan.Active.IsUnknown() && !plan.Active.ValueBool() {
		secret := created.Secret
		paused, err := r.client.UpdateWebhookEndpointFull(ctx, created.ID, client.WebhookEndpointWrite{Active: client.Ptr(false)})
		if err != nil {
			resp.Diagnostics.AddError("Unable to pause webhook endpoint", err.Error())
			return
		}
		created = paused
		created.Secret = secret
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
	fetched, err := r.client.GetWebhookEndpointFull(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read webhook endpoint", err.Error())
		return
	}
	// A read never re-reveals the secret; keep the stored value. After an
	// import there is none, so read it back once via the audited reveal route.
	r.mapToState(ctx, fetched, &state, &resp.Diagnostics, false)
	if state.Secret.IsNull() || state.Secret.IsUnknown() {
		secret, err := r.client.RevealWebhookSecret(ctx, fetched.ID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read webhook signing secret", err.Error())
			return
		}
		state.Secret = types.StringValue(secret)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update patches the endpoint in place via the Backend API PATCH route, which
// does NOT rotate the signing secret. The PATCH response never returns the
// secret, so it is carried forward from prior state.
func (r *webhookEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, cfg, state webhookEndpointModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	patch := r.writeBody(ctx, &plan, &cfg, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	patch.Active = optionalBool(plan.Active)
	if !plan.Description.Equal(state.Description) && plan.Description.IsNull() {
		patch.Description = ptrStr("")
	}
	updated, err := r.client.UpdateWebhookEndpointFull(ctx, state.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update webhook endpoint", err.Error())
		return
	}
	// The PATCH response omits the secret (it is not rotated by an update); map
	// every other field from the response and carry the create-time secret
	// forward from prior state.
	r.mapToState(ctx, updated, &plan, &resp.Diagnostics, false)
	plan.Secret = state.Secret
	if !plan.RotateTrigger.Equal(state.RotateTrigger) {
		secret, err := r.client.RotateWebhookSecret(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to rotate webhook signing secret", err.Error())
			return
		}
		plan.Secret = types.StringValue(secret)
	}
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

// destinationFromAPI flattens the API's destination object into string settings
// (the type plus its config); credentials_set is exposed separately.
//
// When the configuration (prior) names the settings, only those keys are kept,
// with the API's values: Atlas fills in defaults (an EventBridge `source`) the
// configuration did not ask for, which would otherwise read as drift.
func destinationFromAPI(ctx context.Context, d map[string]interface{}, prior types.Map, diags *diag.Diagnostics) (types.Map, bool) {
	all := map[string]string{}
	set := false
	for k, v := range d {
		if k == "credentials_set" {
			if b, ok := v.(bool); ok {
				set = b
			}
			continue
		}
		if s, ok := v.(string); ok {
			all[k] = s
		}
	}
	if prior.IsNull() || prior.IsUnknown() {
		return stringMapToMap(ctx, all, diags), set
	}
	kept := map[string]string{}
	for k := range mapToStringMap(ctx, prior, diags) {
		if v, ok := all[k]; ok {
			kept[k] = v
		}
	}
	return stringMapToMap(ctx, kept, diags), set
}

func (r *webhookEndpointResource) mapToState(ctx context.Context, e *client.WebhookEndpointFull, m *webhookEndpointModel, diags *diag.Diagnostics, withSecret bool) {
	m.ID = types.StringValue(e.ID)
	m.URL = stringPtrToValue(e.URL)
	m.EnabledEvents = stringSliceToList(ctx, e.EnabledEvents, diags)
	m.Active = types.BoolValue(e.Active)
	m.FailingSince = int64PtrToValue(e.FailingSince)
	m.DisabledAt = int64PtrToValue(e.DisabledAt)
	m.CreatedAt = types.Int64Value(e.CreatedAt)
	m.Description = stringPtrToValue(e.Description)
	m.EventNaming = strOrNull(e.EventNaming)
	m.External = types.BoolValue(e.External)
	m.SignatureVersions = stringSliceToSet(ctx, e.SignatureVersions, diags)
	if e.AlertContacts == nil {
		m.AlertContacts = types.SetNull(types.StringType)
	} else {
		m.AlertContacts = stringSliceToSet(ctx, e.AlertContacts, diags)
	}
	dest, credsSet := destinationFromAPI(ctx, e.Destination, m.Destination, diags)
	m.Destination = dest
	m.DestinationCredsSet = types.BoolValue(credsSet)
	if m.DestinationCreds.IsUnknown() {
		m.DestinationCreds = types.MapNull(types.StringType)
	}
	m.ExternalRef = stringPtrToValue(e.ExternalRef)
	m.Metadata = metadataFromAPI(ctx, e.Metadata, diags)
	m.Health = strOrNull(e.Health)
	m.LastSuccessAt = int64PtrToValue(e.LastSuccessAt)
	m.ConsecutiveFailures = types.Int64Value(e.ConsecutiveFailures)
	m.LastFailingSince = int64PtrToValue(e.LastFailingSince)
	if withSecret {
		m.Secret = stringPtrToValue(e.Secret)
	}
}

// ModifyPlan plans a new secret when rotate_secret_trigger changes.
func (r *webhookEndpointResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	planSecretRotation(ctx, req, resp, "secret")
}
