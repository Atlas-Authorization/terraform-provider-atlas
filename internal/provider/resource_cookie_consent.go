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
	_ resource.Resource                = &cookieConsentResource{}
	_ resource.ResourceWithImportState = &cookieConsentResource{}
	_ resource.ResourceWithConfigure   = &cookieConsentResource{}
)

// NewCookieConsentResource is the factory for atlas_cookie_consent.
func NewCookieConsentResource() resource.Resource { return &cookieConsentResource{} }

type cookieConsentResource struct {
	client *client.Client
}

type cookieConsentModel struct {
	ID           types.String `tfsdk:"id"`
	ScriptURL    types.String `tfsdk:"script_url"`
	ScriptID     types.String `tfsdk:"script_id"`
	SiteID       types.String `tfsdk:"site_id"`
	BlockingMode types.String `tfsdk:"blocking_mode"`
	ExtraOrigins types.List   `tfsdk:"extra_origins"`
}

func (r *cookieConsentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cookie_consent"
}

func (r *cookieConsentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *cookieConsentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The cookie-consent banner (any consent-management platform) injected into the hosted " +
			"sign-in, sign-up and account pages — a singleton per instance. Each apply replaces the whole setting; " +
			"destroying the resource turns the banner off. Import with any id, e.g. `cookie_consent`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"script_url": schema.StringAttribute{
				MarkdownDescription: "Your CMP script URL. Must be https; its origin is added to the page CSP.",
				Required:            true,
			},
			"script_id": schema.StringAttribute{
				MarkdownDescription: "The `<script id>` some CMPs need.",
				Optional:            true,
			},
			"site_id": schema.StringAttribute{
				MarkdownDescription: "Your CMP site id (sent as `data-cbid`).",
				Optional:            true,
			},
			"blocking_mode": schema.StringAttribute{
				MarkdownDescription: "The CMP prior-blocking mode, e.g. `auto` (sent as `data-blockingmode`).",
				Optional:            true,
			},
			"extra_origins": schema.ListAttribute{
				MarkdownDescription: "Other https URLs the CMP loads from or calls; each origin is added to the page CSP.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *cookieConsentResource) apply(ctx context.Context, m *cookieConsentModel, diags *diag.Diagnostics) {
	body := client.CookieConsentWrite{
		ScriptURL:    m.ScriptURL.ValueString(),
		ScriptID:     optionalString(m.ScriptID),
		SiteID:       optionalString(m.SiteID),
		BlockingMode: optionalString(m.BlockingMode),
		ExtraOrigins: listToStringSlice(ctx, m.ExtraOrigins, diags),
	}
	if diags.HasError() {
		return
	}
	cc, err := r.client.PutCookieConsent(ctx, body)
	if err != nil {
		diags.AddError("Unable to set cookie consent", err.Error())
		return
	}
	mapCookieConsent(ctx, cc, m, diags)
}

func (r *cookieConsentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cookieConsentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *cookieConsentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cookieConsentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cc, err := r.client.GetCookieConsent(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read cookie consent", err.Error())
		return
	}
	if !cc.Enabled {
		resp.State.RemoveResource(ctx)
		return
	}
	mapCookieConsent(ctx, cc, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cookieConsentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cookieConsentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *cookieConsentResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DeleteCookieConsent(ctx); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to turn off cookie consent", err.Error())
	}
}

func (r *cookieConsentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapCookieConsent projects the API view onto the model. An empty
// extra_origins list stays null when the configuration leaves it unset.
func mapCookieConsent(ctx context.Context, cc *client.CookieConsent, m *cookieConsentModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue("cookie_consent")
	m.ScriptURL = stringPtrToValue(cc.ScriptURL)
	m.ScriptID = stringPtrToValue(cc.ScriptID)
	m.SiteID = stringPtrToValue(cc.SiteID)
	m.BlockingMode = stringPtrToValue(cc.BlockingMode)
	if len(cc.ExtraOrigins) == 0 && m.ExtraOrigins.IsNull() {
		return
	}
	m.ExtraOrigins = stringSliceToList(ctx, cc.ExtraOrigins, diags)
}
