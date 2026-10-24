package provider

import (
	"context"

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
	_ resource.Resource                = &statusSubscriptionResource{}
	_ resource.ResourceWithImportState = &statusSubscriptionResource{}
	_ resource.ResourceWithConfigure   = &statusSubscriptionResource{}
	_ resource.ResourceWithModifyPlan  = &statusSubscriptionResource{}
	_ resource.Resource                = &statusDependencyResource{}
	_ resource.ResourceWithImportState = &statusDependencyResource{}
	_ resource.ResourceWithConfigure   = &statusDependencyResource{}
)

// ───────────────────────────── status subscription ─────────────────────────────

// NewStatusSubscriptionResource is the factory for atlas_status_subscription.
func NewStatusSubscriptionResource() resource.Resource { return &statusSubscriptionResource{} }

type statusSubscriptionResource struct {
	client *client.Client
}

type statusSubscriptionModel struct {
	ID                 types.String `tfsdk:"id"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	Regions            types.Set    `tfsdk:"regions"`
	ComponentIDs       types.Set    `tfsdk:"component_ids"`
	MinImpact          types.String `tfsdk:"min_impact"`
	IncludeMaintenance types.Bool   `tfsdk:"include_maintenance"`
	EmailOwners        types.Bool   `tfsdk:"email_owners"`
	Webhook            types.Bool   `tfsdk:"webhook"`
	RotateFeedTrigger  types.String `tfsdk:"rotate_feed_trigger"`
	RSSURL             types.String `tfsdk:"rss_url"`
	AtomURL            types.String `tfsdk:"atom_url"`
}

func (r *statusSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_subscription"
}

func (r *statusSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func statusSubscriptionSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Which Atlas status incidents concern this instance, and how its owners hear about them — a " +
			"singleton per instance. Incidents matching the filters are emailed to the owners, sent as " +
			"`status.incident.*` webhook events and published on a private RSS/Atom feed. Each apply sets every field " +
			"this resource manages. Destroying it switches the subscription off. Import with any id, e.g. `status_subscription`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `status_subscription`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether incidents are delivered. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"regions": schema.SetAttribute{
				MarkdownDescription: "Only incidents in these regions (ids from the status page). Empty (the default) means all.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringSet(),
			},
			"component_ids": schema.SetAttribute{
				MarkdownDescription: "Only incidents touching these components. Empty (the default) means all.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringSet(),
			},
			"min_impact": schema.StringAttribute{
				MarkdownDescription: "The lowest impact worth hearing about: `minor` (default), `major` or `critical`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("minor"),
			},
			"include_maintenance": schema.BoolAttribute{
				MarkdownDescription: "Also deliver scheduled maintenance. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"email_owners": schema.BoolAttribute{
				MarkdownDescription: "Email the account owners. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"webhook": schema.BoolAttribute{
				MarkdownDescription: "Send `status.incident.*` events to the instance's webhook endpoints. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"rotate_feed_trigger": schema.StringAttribute{
				MarkdownDescription: "Any value; changing it mints a new private feed token, so `rss_url` and `atom_url` change " +
					"and the old URLs stop working.",
				Optional: true,
			},
			"rss_url": schema.StringAttribute{
				MarkdownDescription: "The private RSS feed of matching incidents. The URL carries the feed token.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"atom_url": schema.StringAttribute{
				MarkdownDescription: "The private Atom feed of matching incidents. The URL carries the feed token.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *statusSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = statusSubscriptionSchema()
}

func fillStatusSubscription(ctx context.Context, m *statusSubscriptionModel, s *client.StatusSubscription, diags *diag.Diagnostics) {
	m.ID = types.StringValue("status_subscription")
	m.Enabled = types.BoolValue(s.Enabled)
	m.Regions = stringSliceToSet(ctx, s.Regions, diags)
	m.ComponentIDs = stringSliceToSet(ctx, s.ComponentIDs, diags)
	m.MinImpact = types.StringValue(s.MinImpact)
	m.IncludeMaintenance = types.BoolValue(s.IncludeMaintenance)
	m.EmailOwners = types.BoolValue(s.EmailOwners)
	m.Webhook = types.BoolValue(s.Webhook)
	m.RSSURL = stringPtrToValue(s.RSSURL)
	m.AtomURL = stringPtrToValue(s.AtomURL)
}

func (r *statusSubscriptionResource) apply(ctx context.Context, plan *statusSubscriptionModel, prior *statusSubscriptionModel, diags *diag.Diagnostics) {
	regions := setToStringSlice(ctx, plan.Regions, diags)
	if regions == nil {
		regions = []string{}
	}
	components := setToStringSlice(ctx, plan.ComponentIDs, diags)
	if components == nil {
		components = []string{}
	}
	if diags.HasError() {
		return
	}
	saved, err := r.client.PutStatusSubscription(ctx, client.StatusSubscriptionWrite{
		Enabled:            optionalBool(plan.Enabled),
		Regions:            &regions,
		ComponentIDs:       &components,
		MinImpact:          optionalString(plan.MinImpact),
		IncludeMaintenance: optionalBool(plan.IncludeMaintenance),
		EmailOwners:        optionalBool(plan.EmailOwners),
		Webhook:            optionalBool(plan.Webhook),
	})
	if err != nil {
		diags.AddError("Unable to set the status subscription", err.Error())
		return
	}
	if prior != nil && !plan.RotateFeedTrigger.Equal(prior.RotateFeedTrigger) {
		saved, err = r.client.RotateStatusFeed(ctx)
		if err != nil {
			diags.AddError("Unable to rotate the status feed token", err.Error())
			return
		}
	}
	fillStatusSubscription(ctx, plan, saved, diags)
}

func (r *statusSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan statusSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state statusSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	s, err := r.client.GetStatusSubscription(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the status subscription", err.Error())
		return
	}
	if !s.Subscribed {
		resp.State.RemoveResource(ctx)
		return
	}
	fillStatusSubscription(ctx, &state, s, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *statusSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state statusSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DisableStatusSubscription(ctx); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to switch off the status subscription", err.Error())
	}
}

func (r *statusSubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "status_subscription")...)
}

// ModifyPlan plans new feed URLs when rotate_feed_trigger changes.
func (r *statusSubscriptionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var planned, prior types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("rotate_feed_trigger"), &planned)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("rotate_feed_trigger"), &prior)...)
	if resp.Diagnostics.HasError() || planned.Equal(prior) {
		return
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("rss_url"), types.StringUnknown())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("atom_url"), types.StringUnknown())...)
}

// ───────────────────────────── status dependency ─────────────────────────────

// NewStatusDependencyResource is the factory for atlas_status_dependency.
func NewStatusDependencyResource() resource.Resource { return &statusDependencyResource{} }

type statusDependencyResource struct {
	client *client.Client
}

type statusDependencyModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	URL             types.String `tfsdk:"url"`
	ComponentFilter types.Set    `tfsdk:"component_filter"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	State           types.String `tfsdk:"state"`
	Indicator       types.String `tfsdk:"indicator"`
	Description     types.String `tfsdk:"description"`
	LastCheckedAt   types.Int64  `tfsdk:"last_checked_at"`
	LastError       types.String `tfsdk:"last_error"`
	Stale           types.Bool   `tfsdk:"stale"`
}

func (r *statusDependencyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_dependency"
}

func (r *statusDependencyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func statusDependencySchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "An upstream status page this instance depends on (any Statuspage-compatible page, e.g. your " +
			"email or SMS provider). Atlas polls it, shows its state next to the instance's own, and the instance preflight " +
			"reports `status_dependency_unreachable` / `status_dependency_outage`. An instance can declare a limited number " +
			"of dependencies. `url` cannot change (changing it replaces the dependency).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A label (up to 80 characters).",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The https address of the Statuspage-compatible status page. A trailing slash is removed.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"component_filter": schema.SetAttribute{
				MarkdownDescription: "Only these components (names or ids, up to 50) count toward the state. Empty (the default) means all.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringSet(),
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas polls the page. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "`operational`, `degraded`, `partial_outage`, `major_outage`, `maintenance` or `unknown`.",
				Computed:            true,
			},
			"indicator": schema.StringAttribute{
				MarkdownDescription: "The page's own overall indicator.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The page's own overall description.",
				Computed:            true,
			},
			"last_checked_at": schema.Int64Attribute{
				MarkdownDescription: "When Atlas last polled the page (epoch ms).",
				Computed:            true,
			},
			"last_error": schema.StringAttribute{
				MarkdownDescription: "Why the last poll failed, if it did.",
				Computed:            true,
			},
			"stale": schema.BoolAttribute{
				MarkdownDescription: "The last poll failed or has never succeeded, so `state` may be out of date.",
				Computed:            true,
			},
		},
	}
}

func (r *statusDependencyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = statusDependencySchema()
}

func fillStatusDependency(ctx context.Context, m *statusDependencyModel, d *client.StatusDependency, diags *diag.Diagnostics) {
	m.ID = types.StringValue(d.ID)
	m.Name = types.StringValue(d.Name)
	// The API stores the address without a trailing slash; keep a configured
	// spelling that differs only by that slash.
	keepConfigured := !m.URL.IsNull() && !m.URL.IsUnknown() && m.URL.ValueString() == d.URL+"/"
	if !keepConfigured {
		m.URL = types.StringValue(d.URL)
	}
	m.ComponentFilter = stringSliceToSet(ctx, d.ComponentFilter, diags)
	m.Enabled = types.BoolValue(d.Enabled)
	m.State = types.StringValue(d.State)
	m.Indicator = stringPtrToValue(d.Indicator)
	m.Description = stringPtrToValue(d.Description)
	m.LastCheckedAt = int64PtrToValue(d.LastCheckedAt)
	m.LastError = stringPtrToValue(d.LastError)
	m.Stale = types.BoolValue(d.Stale)
}

func (r *statusDependencyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan statusDependencyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	filter := setToStringSlice(ctx, plan.ComponentFilter, &resp.Diagnostics)
	if filter == nil {
		filter = []string{}
	}
	created, err := r.client.CreateStatusDependency(ctx, client.StatusDependencyWrite{
		Name:            optionalString(plan.Name),
		URL:             optionalString(plan.URL),
		ComponentFilter: &filter,
		Enabled:         optionalBool(plan.Enabled),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create status dependency", err.Error())
		return
	}
	fillStatusDependency(ctx, &plan, created, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusDependencyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state statusDependencyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	d, err := r.client.GetStatusDependency(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read status dependency", err.Error())
		return
	}
	fillStatusDependency(ctx, &state, d, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *statusDependencyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state statusDependencyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	filter := setToStringSlice(ctx, plan.ComponentFilter, &resp.Diagnostics)
	if filter == nil {
		filter = []string{}
	}
	updated, err := r.client.UpdateStatusDependency(ctx, state.ID.ValueString(), client.StatusDependencyWrite{
		Name:            optionalString(plan.Name),
		ComponentFilter: &filter,
		Enabled:         optionalBool(plan.Enabled),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update status dependency", err.Error())
		return
	}
	fillStatusDependency(ctx, &plan, updated, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusDependencyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state statusDependencyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteStatusDependency(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete status dependency", err.Error())
	}
}

func (r *statusDependencyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
