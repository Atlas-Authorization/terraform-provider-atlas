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
	_ resource.Resource                = &billingPlanResource{}
	_ resource.ResourceWithImportState = &billingPlanResource{}
	_ resource.ResourceWithConfigure   = &billingPlanResource{}
)

// NewBillingPlanResource is the factory registered with the provider.
func NewBillingPlanResource() resource.Resource { return &billingPlanResource{} }

type billingPlanResource struct {
	client *client.Client
}

type billingPlanModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Slug          types.String `tfsdk:"slug"`
	StripePriceID types.String `tfsdk:"stripe_price_id"`
	Free          types.Bool   `tfsdk:"free"`
	Audience      types.String `tfsdk:"audience"`
	Interval      types.String `tfsdk:"interval"`
	Amount        types.String `tfsdk:"amount"`
	Currency      types.String `tfsdk:"currency"`
	Features      types.List   `tfsdk:"features"`
	Active        types.Bool   `tfsdk:"active"`
	PricingModel  types.String `tfsdk:"pricing_model"`
	TrialDays     types.Int64  `tfsdk:"trial_days"`
	StripeMeterID types.String `tfsdk:"stripe_meter_id"`
	UsageUnit     types.String `tfsdk:"usage_unit"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	UpdatedAt     types.Int64  `tfsdk:"updated_at"`
}

func (r *billingPlanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_plan"
}

func (r *billingPlanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *billingPlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A billing plan a tenant defines for their app's users (or orgs). The plan's `slug` becomes " +
			"the `pla` session claim and its `features` the `fea` claim. A plan with no `stripe_price_id` is the free " +
			"tier (`free` is then true). A subscription's status is written only by the verified Stripe webhook, never here.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable plan name.",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-safe, instance-unique plan key — surfaced as the `pla` session claim.",
				Required:            true,
			},
			"stripe_price_id": schema.StringAttribute{
				MarkdownDescription: "The recurring Stripe price a subscriber is charged. Omit for the free tier.",
				Optional:            true,
			},
			"free": schema.BoolAttribute{
				MarkdownDescription: "True when the plan has no `stripe_price_id` (the free tier).",
				Computed:            true,
			},
			"audience": schema.StringAttribute{
				MarkdownDescription: "Who may subscribe: `user` or `org`. Defaults to `user`.",
				Optional:            true,
				Computed:            true,
			},
			"interval": schema.StringAttribute{
				MarkdownDescription: "Billing interval: `month` or `year`. Defaults to `month`.",
				Optional:            true,
				Computed:            true,
			},
			"amount": schema.StringAttribute{
				MarkdownDescription: "A free-form display string for the price, stored verbatim by Atlas — not " +
					"parsed or converted. Use whatever units you display (e.g. `12.00` for $12, or `1200` for cents); " +
					"Atlas does not interpret it. Pair it with `currency` and `interval`.",
				Optional: true,
				Computed: true,
			},
			"currency": schema.StringAttribute{
				MarkdownDescription: "ISO currency code (e.g. `usd`). Defaults to `usd`.",
				Optional:            true,
				Computed:            true,
			},
			"features": schema.ListAttribute{
				MarkdownDescription: "Feature keys this plan grants, surfaced as the `fea` session claim.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the plan is offered. Defaults to true.",
				Optional:            true,
				Computed:            true,
			},
			"pricing_model": schema.StringAttribute{
				MarkdownDescription: "`flat`, `per_seat` or `metered`. Defaults to `flat`.",
				Optional:            true,
				Computed:            true,
			},
			"trial_days": schema.Int64Attribute{
				MarkdownDescription: "Free trial length in days (0–3650), or omitted for none.",
				Optional:            true,
			},
			"stripe_meter_id": schema.StringAttribute{
				MarkdownDescription: "Stripe meter id, for a `metered` plan.",
				Optional:            true,
			},
			"usage_unit": schema.StringAttribute{
				MarkdownDescription: "Usage unit label, for a `metered` plan.",
				Optional:            true,
			},
			"created_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time (epoch ms)."},
			"updated_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

func (r *billingPlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan billingPlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.BillingPlanWrite{
		Name:          client.Ptr(plan.Name.ValueString()),
		Slug:          client.Ptr(plan.Slug.ValueString()),
		StripePriceID: optionalString(plan.StripePriceID),
		Audience:      optionalString(plan.Audience),
		Interval:      optionalString(plan.Interval),
		Amount:        optionalString(plan.Amount),
		Currency:      optionalString(plan.Currency),
		Features:      listToStringSlice(ctx, plan.Features, &resp.Diagnostics),
		Active:        optionalBool(plan.Active),
		PricingModel:  optionalString(plan.PricingModel),
		TrialDays:     optionalInt64(plan.TrialDays),
		StripeMeterID: optionalString(plan.StripeMeterID),
		UsageUnit:     optionalString(plan.UsageUnit),
	}
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateBillingPlan(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create billing plan", err.Error())
		return
	}
	r.mapToState(ctx, created, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *billingPlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state billingPlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetBillingPlan(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read billing plan", err.Error())
		return
	}
	r.mapToState(ctx, fetched, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *billingPlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan billingPlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.BillingPlanWrite{
		Name:          client.Ptr(plan.Name.ValueString()),
		Slug:          client.Ptr(plan.Slug.ValueString()),
		StripePriceID: optionalString(plan.StripePriceID),
		Audience:      optionalString(plan.Audience),
		Interval:      optionalString(plan.Interval),
		Amount:        optionalString(plan.Amount),
		Currency:      optionalString(plan.Currency),
		Features:      listToStringSlice(ctx, plan.Features, &resp.Diagnostics),
		Active:        optionalBool(plan.Active),
		PricingModel:  optionalString(plan.PricingModel),
		TrialDays:     optionalInt64(plan.TrialDays),
		StripeMeterID: optionalString(plan.StripeMeterID),
		UsageUnit:     optionalString(plan.UsageUnit),
	}
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateBillingPlan(ctx, plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update billing plan", err.Error())
		return
	}
	r.mapToState(ctx, updated, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *billingPlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state billingPlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteBillingPlan(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete billing plan", err.Error())
	}
}

func (r *billingPlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *billingPlanResource) mapToState(ctx context.Context, p *client.BillingPlan, m *billingPlanModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(p.ID)
	m.Name = types.StringValue(p.Name)
	m.Slug = types.StringValue(p.Slug)
	m.StripePriceID = stringPtrToValue(p.StripePriceID)
	m.Free = types.BoolValue(p.Free)
	m.Audience = types.StringValue(p.Audience)
	m.Interval = types.StringValue(p.Interval)
	m.Amount = stringPtrToValue(p.Amount)
	m.Currency = types.StringValue(p.Currency)
	m.Features = stringSliceToList(ctx, p.Features, diags)
	m.Active = types.BoolValue(p.Active)
	m.PricingModel = types.StringValue(p.PricingModel)
	m.TrialDays = int64PtrToValue(p.TrialDays)
	m.StripeMeterID = stringPtrToValue(p.StripeMeterID)
	m.UsageUnit = stringPtrToValue(p.UsageUnit)
	m.CreatedAt = types.Int64Value(p.CreatedAt)
	m.UpdatedAt = types.Int64Value(p.UpdatedAt)
}
