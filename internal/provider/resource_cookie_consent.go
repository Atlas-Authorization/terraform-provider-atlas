package provider

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                   = &cookieConsentResource{}
	_ resource.ResourceWithImportState    = &cookieConsentResource{}
	_ resource.ResourceWithConfigure      = &cookieConsentResource{}
	_ resource.ResourceWithValidateConfig = &cookieConsentResource{}
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
	SettingsLink types.Object `tfsdk:"settings_link"`
}

// cookieSettingsLinkModel is the settings_link object.
type cookieSettingsLinkModel struct {
	Label              types.String `tfsdk:"label"`
	Preset             types.String `tfsdk:"preset"`
	Attributes         types.Map    `tfsdk:"attributes"`
	RenderedAttributes types.Map    `tfsdk:"rendered_attributes"`
}

var cookieSettingsLinkAttrTypes = map[string]attr.Type{
	"label":               types.StringType,
	"preset":              types.StringType,
	"attributes":          types.MapType{ElemType: types.StringType},
	"rendered_attributes": types.MapType{ElemType: types.StringType},
}

// The API's limits on settings_link, checked at plan time so a bad link fails
// `terraform plan` rather than the apply.
var cookieSettingsPresets = []string{"cookiemunch", "cookiebot", "onetrust", "custom"}

const (
	cookieSettingsLabelMax     = 60
	cookieSettingsMaxAttrs     = 10
	cookieSettingsAttrNameMax  = 64
	cookieSettingsAttrValueMax = 200
)

var (
	cookieSettingsDataAttr = regexp.MustCompile(`^data-[a-z0-9][a-z0-9._-]*$`)
	cookieSettingsAriaAttr = regexp.MustCompile(`^aria-[a-z]+$`)
)

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
			"settings_link": schema.SingleNestedAttribute{
				MarkdownDescription: "A **Cookie settings** link in the footer of the hosted pages that reopens your CMP's " +
					"preferences, so a visitor can change or withdraw consent after the banner is gone. Set a `preset` " +
					"for your CMP, your own `attributes`, or both. When unset, Atlas supplies the default link for a CMP it " +
					"recognises from `script_url` (Cookiebot, OneTrust, Cookie Munch), and none otherwise. Removing the block later leaves the current link in place.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"label": schema.StringAttribute{
						MarkdownDescription: "The link text, 1–60 characters, e.g. `Cookie settings`.",
						Required:            true,
					},
					"preset": schema.StringAttribute{
						MarkdownDescription: "The CMP's reopen trigger: `cookiemunch` (adds `data-cc=\"show-settings\"`), " +
							"`cookiebot` (calls `Cookiebot.renew()`), `onetrust` (calls `OneTrust.ToggleInfoDisplay()`) or " +
							"`custom` (adds nothing; use `attributes`).",
						Optional: true,
					},
					"attributes": schema.MapAttribute{
						MarkdownDescription: "Extra HTML attributes your CMP's click handler listens for. Allowed names: " +
							"`data-*`, `aria-*`, `id`, `class` and `href` (an https URL or a `#fragment`), lower case; at " +
							"most 10, values up to 200 characters. `data-atlas-*` is reserved for the presets.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"rendered_attributes": schema.MapAttribute{
						MarkdownDescription: "What Atlas renders on the link: `attributes` with the preset's trigger applied.",
						Computed:            true,
						ElementType:         types.StringType,
					},
				},
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
		SettingsLink: settingsLinkFromPlan(ctx, m.SettingsLink, diags),
	}
	if diags.HasError() {
		return
	}
	cc, err := r.client.PutCookieConsent(ctx, body)
	if err != nil {
		diags.AddError("Unable to set cookie consent", err.Error())
		return
	}
	// The PUT answers with the stored setting; a read adds the default settings
	// link Atlas derives for a recognised CMP, so state matches the next refresh.
	if got, gerr := r.client.GetCookieConsent(ctx); gerr == nil && got.Enabled {
		cc = got
	}
	mapCookieConsent(ctx, cc, m, diags)
}

// ValidateConfig checks settings_link against the API's rules at plan time.
func (r *cookieConsentResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var obj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("settings_link"), &obj)...)
	if resp.Diagnostics.HasError() || obj.IsNull() || obj.IsUnknown() {
		return
	}
	var link cookieSettingsLinkModel
	resp.Diagnostics.Append(obj.As(ctx, &link, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	base := path.Root("settings_link")
	if !link.Label.IsUnknown() {
		label := link.Label.ValueString()
		if n := len([]rune(label)); n < 1 || n > cookieSettingsLabelMax || strings.TrimSpace(label) != label {
			resp.Diagnostics.AddAttributeError(base.AtName("label"), "Invalid settings_link label",
				fmt.Sprintf("label must be 1-%d characters without leading or trailing spaces.", cookieSettingsLabelMax))
		}
	}
	presetSet := false
	if !link.Preset.IsNull() && !link.Preset.IsUnknown() {
		presetSet = true
		if !containsString(cookieSettingsPresets, link.Preset.ValueString()) {
			resp.Diagnostics.AddAttributeError(base.AtName("preset"), "Invalid settings_link preset",
				"preset must be one of "+strings.Join(cookieSettingsPresets, ", ")+".")
		}
	}
	if link.Preset.IsUnknown() || link.Attributes.IsUnknown() {
		return
	}
	attrs := mapToStringMap(ctx, link.Attributes, &resp.Diagnostics)
	if len(attrs) > cookieSettingsMaxAttrs {
		resp.Diagnostics.AddAttributeError(base.AtName("attributes"), "Too many settings_link attributes",
			fmt.Sprintf("At most %d attributes are allowed; found %d.", cookieSettingsMaxAttrs, len(attrs)))
	}
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if why := cookieSettingsAttributeError(name, attrs[name]); why != "" {
			resp.Diagnostics.AddAttributeError(base.AtName("attributes").AtMapKey(name), "Invalid settings_link attribute",
				fmt.Sprintf("attributes[%q] %s.", name, why))
		}
	}
	if !presetSet && len(attrs) == 0 {
		resp.Diagnostics.AddAttributeError(base, "settings_link needs a preset or attributes",
			"Set a preset for your CMP, or at least one attribute its click handler listens for.")
	}
}

// cookieSettingsAttributeError mirrors the API's attribute allowlist; "" when
// the attribute is allowed. Names must be lower case (Atlas stores them so).
func cookieSettingsAttributeError(name, value string) string {
	if len([]rune(value)) > cookieSettingsAttrValueMax {
		return fmt.Sprintf("must be at most %d characters", cookieSettingsAttrValueMax)
	}
	if name != strings.ToLower(name) {
		return "must be a lower-case attribute name"
	}
	if len(name) > cookieSettingsAttrNameMax {
		return "is not an allowed attribute name"
	}
	if strings.HasPrefix(name, "on") {
		return "is an event handler, which is not allowed"
	}
	if strings.HasPrefix(name, "data-atlas-") {
		return "is reserved (data-atlas-*); use a preset"
	}
	switch {
	case name == "id" || name == "class":
		return ""
	case cookieSettingsDataAttr.MatchString(name) || cookieSettingsAriaAttr.MatchString(name):
		return ""
	case name == "href":
		if strings.HasPrefix(value, "#") {
			return ""
		}
		if u, err := url.Parse(value); err == nil && u.Scheme == "https" && u.Host != "" {
			return ""
		}
		return "must be an https URL or a #fragment"
	}
	return "is not an allowed attribute (data-*, aria-*, id, class, href)"
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// settingsLinkFromPlan builds the PUT's settings_link; nil (sent as null,
// which removes the link) when the block is unset.
func settingsLinkFromPlan(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *client.CookieSettingsLinkWrite {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var link cookieSettingsLinkModel
	diags.Append(obj.As(ctx, &link, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	out := &client.CookieSettingsLinkWrite{Label: link.Label.ValueString(), Preset: optionalString(link.Preset)}
	out.Attributes = mapToStringMap(ctx, link.Attributes, diags)
	return out
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
	if len(cc.ExtraOrigins) > 0 || !m.ExtraOrigins.IsNull() {
		m.ExtraOrigins = stringSliceToList(ctx, cc.ExtraOrigins, diags)
	}
	m.SettingsLink = settingsLinkToObject(ctx, cc.SettingsLink, m.SettingsLink, diags)
}

// settingsLinkToObject projects the API's settings_link. An empty attributes
// map stays null when the configuration leaves it unset.
func settingsLinkToObject(ctx context.Context, link *client.CookieSettingsLink, prior types.Object, diags *diag.Diagnostics) types.Object {
	if link == nil {
		return types.ObjectNull(cookieSettingsLinkAttrTypes)
	}
	priorAttrsNull := true
	if !prior.IsNull() && !prior.IsUnknown() {
		var p cookieSettingsLinkModel
		diags.Append(prior.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
		priorAttrsNull = p.Attributes.IsNull()
	}
	attributes := types.MapNull(types.StringType)
	if len(link.Attributes) > 0 || !priorAttrsNull {
		attributes = stringMapToMap(ctx, link.Attributes, diags)
	}
	obj, d := types.ObjectValue(cookieSettingsLinkAttrTypes, map[string]attr.Value{
		"label":               types.StringValue(link.Label),
		"preset":              stringPtrToValue(link.Preset),
		"attributes":          attributes,
		"rendered_attributes": stringMapToMap(ctx, link.RenderedAttributes, diags),
	})
	diags.Append(d...)
	return obj
}
