package provider

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                   = &betterAuthBridgeResource{}
	_ resource.ResourceWithImportState    = &betterAuthBridgeResource{}
	_ resource.ResourceWithConfigure      = &betterAuthBridgeResource{}
	_ resource.ResourceWithValidateConfig = &betterAuthBridgeResource{}
)

// NewBetterAuthBridgeResource is the factory for atlas_better_auth_bridge.
func NewBetterAuthBridgeResource() resource.Resource { return &betterAuthBridgeResource{} }

type betterAuthBridgeResource struct {
	client *client.Client
}

type betterAuthBridgeModel struct {
	ID           types.String `tfsdk:"id"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	EndsAt       types.String `tfsdk:"ends_at"`
	CookiePrefix types.String `tfsdk:"cookie_prefix"`
	LiveFallback types.Object `tfsdk:"live_fallback"`
	ExchangePath types.String `tfsdk:"exchange_path"`
}

type betterAuthLiveFallbackModel struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	BaseURL         types.String `tfsdk:"base_url"`
	SecretWO        types.String `tfsdk:"secret_wo"`
	SecretWOVersion types.Int64  `tfsdk:"secret_wo_version"`
	HasSecret       types.Bool   `tfsdk:"has_secret"`
}

var betterAuthLiveFallbackAttrTypes = map[string]attr.Type{
	"enabled":           types.BoolType,
	"base_url":          types.StringType,
	"secret_wo":         types.StringType,
	"secret_wo_version": types.Int64Type,
	"has_secret":        types.BoolType,
}

const (
	betterAuthDefaultCookiePrefix = "better-auth"
	betterAuthMinSecretLength     = 16
)

var betterAuthCookiePrefixRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func (r *betterAuthBridgeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_better_auth_bridge"
}

func (r *betterAuthBridgeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *betterAuthBridgeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Better Auth session bridge — a singleton per instance. During the migration window a " +
			"returning user's Better Auth session is exchanged once for an Atlas session, so moving from Better Auth " +
			"signs nobody out. Each apply sets every field this resource manages; destroying it turns the bridge off " +
			"and deletes the stored secret. Import with any id, e.g. `better_auth_bridge`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `better_auth_bridge`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas accepts Better Auth sessions for exchange. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ends_at": schema.StringAttribute{
				MarkdownDescription: "End of the migration window, an RFC 3339 timestamp (e.g. `2026-12-31T00:00:00Z`). " +
					"After it every exchange is refused. Unset = open until disabled.",
				Optional: true,
			},
			"cookie_prefix": schema.StringAttribute{
				MarkdownDescription: "Better Auth's `advanced.cookiePrefix` (1-64 letters, digits, `.`, `_` or `-`). " +
					"Default `better-auth`; the session cookie is `<prefix>.session_token`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(betterAuthDefaultCookiePrefix),
			},
			"live_fallback": schema.SingleNestedAttribute{
				MarkdownDescription: "Validate a session missing from the import (created after the export) by asking the " +
					"old app, server to server. Unset = off, and any stored secret is cleared.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Turn the live check on. Requires `base_url`. Default `false`.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"base_url": schema.StringAttribute{
						MarkdownDescription: "The old app's Better Auth base URL including its base path, e.g. " +
							"`https://app.example.com/api/auth`; Atlas calls `<base_url>/get-session`. Must be a public " +
							"https URL on a production instance.",
						Optional: true,
					},
					"secret_wo": schema.StringAttribute{
						MarkdownDescription: "The old app's `BETTER_AUTH_SECRET` (at least 16 characters). **Write-only**: " +
							"never stored in state or returned by Atlas. It is sent on create and whenever " +
							"`secret_wo_version` changes; change the version to rotate it, or change the version with " +
							"`secret_wo` unset to clear it. Needs Terraform 1.11 or later.",
						Optional:  true,
						Sensitive: true,
						WriteOnly: true,
					},
					"secret_wo_version": schema.Int64Attribute{
						MarkdownDescription: "Any number; changing it sends `secret_wo` again (or clears the secret when " +
							"`secret_wo` is unset).",
						Optional: true,
					},
					"has_secret": schema.BoolAttribute{
						MarkdownDescription: "Whether Atlas holds a secret for the old app (never the secret).",
						Computed:            true,
					},
				},
			},
			"exchange_path": schema.StringAttribute{
				MarkdownDescription: "The Frontend API path the SDKs call for the browser exchange.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

// ValidateConfig checks the API's rules at plan time.
func (r *betterAuthBridgeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var m betterAuthBridgeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !m.CookiePrefix.IsNull() && !m.CookiePrefix.IsUnknown() && !betterAuthCookiePrefixRe.MatchString(m.CookiePrefix.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("cookie_prefix"), "Invalid cookie_prefix",
			"cookie_prefix must be 1-64 letters, digits, \".\", \"_\" or \"-\".")
	}
	if !m.EndsAt.IsNull() && !m.EndsAt.IsUnknown() {
		if _, err := time.Parse(time.RFC3339, m.EndsAt.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("ends_at"), "Invalid ends_at",
				"ends_at must be an RFC 3339 timestamp such as 2026-12-31T00:00:00Z.")
		}
	}
	if m.LiveFallback.IsNull() || m.LiveFallback.IsUnknown() {
		return
	}
	var live betterAuthLiveFallbackModel
	resp.Diagnostics.Append(m.LiveFallback.As(ctx, &live, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}
	base := path.Root("live_fallback")
	if !live.BaseURL.IsNull() && !live.BaseURL.IsUnknown() {
		u, err := url.Parse(live.BaseURL.ValueString())
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			resp.Diagnostics.AddAttributeError(base.AtName("base_url"), "Invalid live_fallback.base_url",
				"base_url must be an absolute URL such as https://app.example.com/api/auth.")
		}
	}
	if live.Enabled.ValueBool() && live.BaseURL.IsNull() {
		resp.Diagnostics.AddAttributeError(base.AtName("base_url"), "live_fallback.base_url is required",
			"Set base_url to enable the live fallback.")
	}
	if !live.SecretWO.IsNull() && !live.SecretWO.IsUnknown() && len(live.SecretWO.ValueString()) < betterAuthMinSecretLength {
		resp.Diagnostics.AddAttributeError(base.AtName("secret_wo"), "live_fallback.secret_wo is too short",
			fmt.Sprintf("The Better Auth secret must be at least %d characters.", betterAuthMinSecretLength))
	}
}

// liveFallbackOf reads a live_fallback object (nil when the block is unset).
func liveFallbackOf(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *betterAuthLiveFallbackModel {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var m betterAuthLiveFallbackModel
	diags.Append(obj.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	return &m
}

// apply PUTs the planned settings. prior is the state's live_fallback (nil on
// create); config carries the write-only secret.
func (r *betterAuthBridgeResource) apply(ctx context.Context, plan *betterAuthBridgeModel, config betterAuthBridgeModel, prior *betterAuthLiveFallbackModel, creating bool, diags *diag.Diagnostics) {
	body := client.BetterAuthBridgeWrite{
		Enabled:      plan.Enabled.ValueBool(),
		CookiePrefix: plan.CookiePrefix.ValueString(),
	}
	if plan.CookiePrefix.IsNull() || plan.CookiePrefix.IsUnknown() {
		body.CookiePrefix = betterAuthDefaultCookiePrefix
	}
	if s := optionalString(plan.EndsAt); s != nil {
		body.EndsAt = s
	}
	live := liveFallbackOf(ctx, plan.LiveFallback, diags)
	cfgLive := liveFallbackOf(ctx, config.LiveFallback, diags)
	if diags.HasError() {
		return
	}
	switch {
	case live == nil:
		// No block: the live fallback is off and any stored secret is cleared.
		body.LiveFallback = client.BetterAuthBridgeLiveFallbackWrite{SendSecret: true}
	default:
		body.LiveFallback.Enabled = live.Enabled.ValueBool()
		body.LiveFallback.BaseURL = optionalString(live.BaseURL)
		var secret *string
		if cfgLive != nil {
			secret = optionalString(cfgLive.SecretWO)
		}
		versionChanged := prior == nil || !live.SecretWOVersion.Equal(prior.SecretWOVersion)
		switch {
		case creating || prior == nil:
			// First apply of the block: send the secret when one is configured.
			if secret != nil {
				body.LiveFallback.SendSecret, body.LiveFallback.Secret = true, secret
			}
		case versionChanged:
			// A new version sends the secret again, or clears it when unset.
			body.LiveFallback.SendSecret, body.LiveFallback.Secret = true, secret
		}
	}
	out, err := r.client.PutBetterAuthBridge(ctx, body)
	if err != nil {
		diags.AddError("Unable to set the Better Auth bridge", err.Error())
		return
	}
	mapBetterAuthBridge(ctx, out, plan, diags)
}

func (r *betterAuthBridgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config betterAuthBridgeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, config, nil, true, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *betterAuthBridgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state betterAuthBridgeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.GetBetterAuthBridge(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the Better Auth bridge", err.Error())
		return
	}
	if out.CreatedAt == nil {
		// Never configured, or removed outside Terraform.
		resp.State.RemoveResource(ctx)
		return
	}
	mapBetterAuthBridge(ctx, out, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *betterAuthBridgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state betterAuthBridgeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	prior := liveFallbackOf(ctx, state.LiveFallback, &resp.Diagnostics)
	r.apply(ctx, &plan, config, prior, false, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *betterAuthBridgeResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DeleteBetterAuthBridge(ctx); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove the Better Auth bridge", err.Error())
	}
}

func (r *betterAuthBridgeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// formatEndsAt renders the API's epoch-ms window end, keeping the prior
// configured spelling when it names the same instant (so an offset or a
// fractional second in the configuration does not show a diff).
func formatEndsAt(ms *int64, prior types.String) types.String {
	if ms == nil {
		return types.StringNull()
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		if t, err := time.Parse(time.RFC3339, prior.ValueString()); err == nil && t.UnixMilli() == *ms {
			return prior
		}
	}
	return types.StringValue(time.UnixMilli(*ms).UTC().Format(time.RFC3339Nano))
}

// mapBetterAuthBridge projects the API view onto the model. The write-only
// secret is always null in state; secret_wo_version is kept from the model. An
// unset live_fallback block stays null while the API reports the defaults.
func mapBetterAuthBridge(ctx context.Context, out *client.BetterAuthBridge, m *betterAuthBridgeModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue("better_auth_bridge")
	m.Enabled = types.BoolValue(out.Enabled)
	m.EndsAt = formatEndsAt(out.EndsAt, m.EndsAt)
	prefix := out.CookiePrefix
	if prefix == "" {
		prefix = betterAuthDefaultCookiePrefix
	}
	m.CookiePrefix = types.StringValue(prefix)
	m.ExchangePath = types.StringValue(out.ExchangePath)

	prior := liveFallbackOf(ctx, m.LiveFallback, diags)
	lf := out.LiveFallback
	if prior == nil && !lf.Enabled && lf.BaseURL == nil && !lf.HasSecret {
		m.LiveFallback = types.ObjectNull(betterAuthLiveFallbackAttrTypes)
		return
	}
	version := types.Int64Null()
	if prior != nil {
		version = prior.SecretWOVersion
	}
	baseURL := stringPtrToValue(lf.BaseURL)
	// Atlas stores the URL without trailing slashes; keep the configured spelling.
	if prior != nil && lf.BaseURL != nil && !prior.BaseURL.IsNull() && !prior.BaseURL.IsUnknown() &&
		strings.TrimRight(prior.BaseURL.ValueString(), "/") == *lf.BaseURL {
		baseURL = prior.BaseURL
	}
	obj, d := types.ObjectValue(betterAuthLiveFallbackAttrTypes, map[string]attr.Value{
		"enabled":           types.BoolValue(lf.Enabled),
		"base_url":          baseURL,
		"secret_wo":         types.StringNull(),
		"secret_wo_version": version,
		"has_secret":        types.BoolValue(lf.HasSecret),
	})
	diags.Append(d...)
	m.LiveFallback = obj
}
