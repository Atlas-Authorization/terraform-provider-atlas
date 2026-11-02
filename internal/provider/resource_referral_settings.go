package provider

import (
	"context"
	"strings"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &referralSettingsResource{}
	_ resource.ResourceWithImportState    = &referralSettingsResource{}
	_ resource.ResourceWithConfigure      = &referralSettingsResource{}
	_ resource.ResourceWithValidateConfig = &referralSettingsResource{}
)

// NewReferralSettingsResource is the factory for atlas_referral_settings.
func NewReferralSettingsResource() resource.Resource { return &referralSettingsResource{} }

type referralSettingsResource struct {
	client *client.Client
}

type referralSettingsModel struct {
	ID                          types.String `tfsdk:"id"`
	Enabled                     types.Bool   `tfsdk:"enabled"`
	ProviderID                  types.String `tfsdk:"provider_id"`
	EndpointURL                 types.String `tfsdk:"endpoint_url"`
	ProgramID                   types.String `tfsdk:"program_id"`
	ConnectionID                types.String `tfsdk:"connection_id"`
	SigningSecretWO             types.String `tfsdk:"signing_secret_wo"`
	SigningSecretWOVersion      types.Int64  `tfsdk:"signing_secret_wo_version"`
	APIKeyWO                    types.String `tfsdk:"api_key_wo"`
	APIKeyWOVersion             types.Int64  `tfsdk:"api_key_wo_version"`
	URLParams                   types.List   `tfsdk:"url_params"`
	ContextCookies              types.List   `tfsdk:"context_cookies"`
	ConsentCategory             types.String `tfsdk:"consent_category"`
	ConsentFallbackCategory     types.String `tfsdk:"consent_fallback_category"`
	AutoEnrollUsers             types.Bool   `tfsdk:"auto_enroll_users"`
	LinkTemplate                types.String `tfsdk:"link_template"`
	TimeoutMs                   types.Int64  `tfsdk:"timeout_ms"`
	CodeLookupURL               types.String `tfsdk:"code_lookup_url"`
	SignupCodeField             types.Bool   `tfsdk:"signup_code_field"`
	BrowserSources              types.List   `tfsdk:"browser_sources"`
	ConsentProvider             types.String `tfsdk:"consent_provider"`
	ChallengeReferredSignups    types.String `tfsdk:"challenge_referred_signups"`
	OnHighRisk                  types.String `tfsdk:"on_high_risk"`
	HoldUntilEmailVerified      types.Bool   `tfsdk:"hold_until_email_verified"`
	MaxReferredSignupsPerIPHour types.Int64  `tfsdk:"max_referred_signups_per_ip_hour"`
	ForwardLifecycle            types.Bool   `tfsdk:"forward_lifecycle"`
	MirrorToPublicMetadata      types.Bool   `tfsdk:"mirror_to_public_metadata"`
	IncludeReferralClaims       types.Bool   `tfsdk:"include_referral_claims"`
	SigningSecretSet            types.Bool   `tfsdk:"signing_secret_set"`
	APIKeySet                   types.Bool   `tfsdk:"api_key_set"`
	Missing                     types.List   `tfsdk:"missing"`
}

func (r *referralSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_referral_settings"
}

func (r *referralSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func refStr(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func refBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
	}
}

func referralSettingsSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Referral capture and attribution — a singleton per instance. Atlas captures the referral code a " +
			"visitor arrives with (URL parameters and the referral program's cookies, gated on consent), attributes the " +
			"sign-up to the referrer and forwards lifecycle events to the referral platform. An attribute you leave out " +
			"keeps Atlas's current value (Atlas fills in defaults for the chosen platform). Secrets are write-only. " +
			"Destroying this resource switches referral capture off. Import with any id, e.g. `referral_settings`. " +
			"The `*_wo` attributes need Terraform 1.11 or later.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `referral_settings`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether referral capture runs. Enabling needs what `missing` lists. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"provider_id":   refStr("The referral platform: `referalli`, `webhook` or another id from the catalogue. The retired spelling `referally` is no longer accepted here. Sent as `provider`."),
			"endpoint_url":  refStr("The platform's API or webhook URL (https)."),
			"program_id":    refStr("The program or campaign id on the platform."),
			"connection_id": refStr("The platform connection id (set by the one-click install, or by hand)."),
			"signing_secret_wo": schema.StringAttribute{
				MarkdownDescription: "The signing secret shared with the platform. **Write-only**: never stored in state or returned by Atlas. " +
					"Sent on create and whenever `signing_secret_wo_version` changes (unset clears it).",
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"signing_secret_wo_version": schema.Int64Attribute{
				MarkdownDescription: "Any number; changing it sends `signing_secret_wo` again (or clears the secret when it is unset).",
				Optional:            true,
			},
			"api_key_wo": schema.StringAttribute{
				MarkdownDescription: "The platform API key. **Write-only**: never stored in state or returned by Atlas. " +
					"Sent on create and whenever `api_key_wo_version` changes (unset clears it).",
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"api_key_wo_version": schema.Int64Attribute{
				MarkdownDescription: "Any number; changing it sends `api_key_wo` again (or clears the key when it is unset).",
				Optional:            true,
			},
			"url_params": schema.ListAttribute{
				MarkdownDescription: "Query parameters that carry the referral code, in the order they are tried (1-N names).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"context_cookies": schema.ListAttribute{
				MarkdownDescription: "Cookies the referral program sets in the browser, in the order they are tried. Several are " +
					"allowed (a program often sets more than one).",
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"consent_category":          refStr("The consent category the capture needs (default: the platform's)."),
			"consent_fallback_category": refStr("A second consent category that also permits capture."),
			"auto_enroll_users":         refBool("Enrol every new user in the program as a referrer."),
			"link_template":             refStr("The referral link template, e.g. `https://app.example.com/?ref={code}`."),
			"timeout_ms": schema.Int64Attribute{
				MarkdownDescription: "How long Atlas waits for the platform on each call, in milliseconds.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"code_lookup_url":   refStr("Hosted sign-up: the URL that resolves a referral code to the referrer's name for the \"Invited by\" line."),
			"signup_code_field": refBool("Hosted sign-up: show an optional referral code field."),
			"browser_sources": schema.ListAttribute{
				MarkdownDescription: "The browser-side sources read for the code (the programs' scripts and cookies).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"consent_provider":           refStr("Which consent source gates capture: `atlas`, `atlasconsent` (the Atlas consent banner), `gpc` (Global Privacy Control) or another supported manager."),
			"challenge_referred_signups": refStr("Fraud protection: `always` (default) challenges referred sign-ups; `inherit` follows the instance setting."),
			"on_high_risk":               refStr("Fraud protection: what to do with a high-risk referred sign-up: `flag`, `strip` the referral, or `hold` it."),
			"hold_until_email_verified":  refBool("Fraud protection: credit the referral only after the new user verifies an email."),
			"max_referred_signups_per_ip_hour": schema.Int64Attribute{
				MarkdownDescription: "Fraud protection: referred sign-ups allowed from one IP address per hour (default 5).",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"forward_lifecycle":         refBool("Forward user lifecycle events (verified, activated, deleted) to the platform. Default on."),
			"mirror_to_public_metadata": refBool("Copy the referral onto the user's public metadata."),
			"include_referral_claims":   refBool("Put the referral on session and ID tokens."),
			"signing_secret_set": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas holds a signing secret.",
				Computed:            true,
			},
			"api_key_set": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas holds an API key.",
				Computed:            true,
			},
			"missing": schema.ListAttribute{
				MarkdownDescription: "What the chosen platform still needs before `enabled` can be true.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *referralSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = referralSettingsSchema()
}

// ValidateConfig rejects the retired provider spelling at plan time, since the
// API would normalise it and the plan would then disagree with the result.
func (r *referralSettingsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var m referralSettingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !m.ProviderID.IsNull() && !m.ProviderID.IsUnknown() && strings.EqualFold(m.ProviderID.ValueString(), "referally") {
		resp.Diagnostics.AddAttributeError(path.Root("provider_id"), "Retired provider id",
			"The provider id is spelled \"referalli\". The old spelling \"referally\" is still accepted by the API but is renamed on read, so use \"referalli\".")
	}
}

func fillReferralSettings(ctx context.Context, m *referralSettingsModel, s *client.ReferralSettings, diags *diag.Diagnostics) {
	m.ID = types.StringValue("referral_settings")
	m.Enabled = types.BoolValue(s.Enabled)
	m.ProviderID = types.StringValue(s.Provider)
	m.EndpointURL = stringPtrToValue(s.EndpointURL)
	m.ProgramID = stringPtrToValue(s.ProgramID)
	m.ConnectionID = stringPtrToValue(s.ConnectionID)
	m.URLParams = stringSliceToList(ctx, s.URLParams, diags)
	m.ContextCookies = stringSliceToList(ctx, s.ContextCookies, diags)
	m.ConsentCategory = stringPtrToValue(s.ConsentCategory)
	m.ConsentFallbackCategory = stringPtrToValue(s.ConsentFallbackCategory)
	m.AutoEnrollUsers = types.BoolValue(s.AutoEnrollUsers)
	m.LinkTemplate = stringPtrToValue(s.LinkTemplate)
	m.TimeoutMs = int64PtrToValue(s.TimeoutMs)
	m.CodeLookupURL = stringPtrToValue(s.CodeLookupURL)
	m.SignupCodeField = types.BoolValue(s.SignupCodeField)
	m.BrowserSources = stringSliceToList(ctx, s.BrowserSources, diags)
	m.ConsentProvider = stringPtrToValue(s.ConsentProvider)
	m.ChallengeReferredSignups = strOrNull(s.Security.ChallengeReferredSignups)
	m.OnHighRisk = strOrNull(s.Security.OnHighRisk)
	m.HoldUntilEmailVerified = types.BoolValue(s.Security.HoldUntilEmailVerified)
	m.MaxReferredSignupsPerIPHour = types.Int64Value(s.Security.MaxReferredSignupsPerIPHour)
	m.ForwardLifecycle = types.BoolValue(s.ForwardLifecycle)
	m.MirrorToPublicMetadata = types.BoolValue(s.MirrorToPublicMetadata)
	m.IncludeReferralClaims = types.BoolValue(s.IncludeReferralClaims)
	m.SigningSecretSet = types.BoolValue(s.SigningSecretSet)
	m.APIKeySet = types.BoolValue(s.APIKeySet)
	m.Missing = stringSliceToList(ctx, s.Missing, diags)
}

// apply PATCHes the settings the configuration names (an omitted attribute keeps
// Atlas's value). prior is the state (nil on create); config carries the
// write-only secrets.
func (r *referralSettingsResource) apply(ctx context.Context, plan, config, prior *referralSettingsModel, diags *diag.Diagnostics) {
	body := map[string]interface{}{"enabled": plan.Enabled.ValueBool()}
	str := func(key string, v types.String) {
		if !v.IsNull() && !v.IsUnknown() {
			body[key] = v.ValueString()
		}
	}
	boolean := func(key string, v types.Bool) {
		if !v.IsNull() && !v.IsUnknown() {
			body[key] = v.ValueBool()
		}
	}
	str("provider", config.ProviderID)
	str("endpoint_url", config.EndpointURL)
	str("program_id", config.ProgramID)
	str("connection_id", config.ConnectionID)
	str("consent_category", config.ConsentCategory)
	str("consent_fallback_category", config.ConsentFallbackCategory)
	str("link_template", config.LinkTemplate)
	str("code_lookup_url", config.CodeLookupURL)
	str("consent_provider", config.ConsentProvider)
	boolean("auto_enroll_users", config.AutoEnrollUsers)
	boolean("signup_code_field", config.SignupCodeField)
	boolean("forward_lifecycle", config.ForwardLifecycle)
	boolean("mirror_to_public_metadata", config.MirrorToPublicMetadata)
	boolean("include_referral_claims", config.IncludeReferralClaims)
	if !config.TimeoutMs.IsNull() && !config.TimeoutMs.IsUnknown() {
		body["timeout_ms"] = config.TimeoutMs.ValueInt64()
	}
	list := func(key string, v types.List) {
		if !v.IsNull() && !v.IsUnknown() {
			items := listToStringSlice(ctx, v, diags)
			if items == nil {
				items = []string{}
			}
			body[key] = items
		}
	}
	list("url_params", config.URLParams)
	list("context_cookies", config.ContextCookies)
	list("browser_sources", config.BrowserSources)

	security := map[string]interface{}{}
	if !config.ChallengeReferredSignups.IsNull() && !config.ChallengeReferredSignups.IsUnknown() {
		security["challenge_referred_signups"] = config.ChallengeReferredSignups.ValueString()
	}
	if !config.OnHighRisk.IsNull() && !config.OnHighRisk.IsUnknown() {
		security["on_high_risk"] = config.OnHighRisk.ValueString()
	}
	if !config.HoldUntilEmailVerified.IsNull() && !config.HoldUntilEmailVerified.IsUnknown() {
		security["hold_until_email_verified"] = config.HoldUntilEmailVerified.ValueBool()
	}
	if !config.MaxReferredSignupsPerIPHour.IsNull() && !config.MaxReferredSignupsPerIPHour.IsUnknown() {
		security["max_referred_signups_per_ip_hour"] = config.MaxReferredSignupsPerIPHour.ValueInt64()
	}
	if len(security) > 0 {
		body["security"] = security
	}

	secret := func(key string, wo types.String, version types.Int64, priorVersion *types.Int64) {
		switch {
		case priorVersion == nil:
			if !wo.IsNull() {
				body[key] = wo.ValueString()
			}
		case !version.Equal(*priorVersion):
			if wo.IsNull() {
				body[key] = nil
			} else {
				body[key] = wo.ValueString()
			}
		}
	}
	var priorSecretVersion, priorKeyVersion *types.Int64
	if prior != nil {
		priorSecretVersion, priorKeyVersion = &prior.SigningSecretWOVersion, &prior.APIKeyWOVersion
	}
	secret("signing_secret", config.SigningSecretWO, plan.SigningSecretWOVersion, priorSecretVersion)
	secret("api_key", config.APIKeyWO, plan.APIKeyWOVersion, priorKeyVersion)
	if diags.HasError() {
		return
	}

	out, err := r.client.PatchReferralSettings(ctx, client.ReferralSettingsWrite{Fields: body})
	if err != nil {
		diags.AddError("Unable to set the referral settings", err.Error())
		return
	}
	fillReferralSettings(ctx, plan, out, diags)
}

func (r *referralSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config referralSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, &config, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *referralSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state referralSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	s, err := r.client.GetReferralSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the referral settings", err.Error())
		return
	}
	fillReferralSettings(ctx, &state, s, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *referralSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state referralSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, &config, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete switches referral capture off (the settings row stays).
func (r *referralSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	_, err := r.client.PatchReferralSettings(ctx, client.ReferralSettingsWrite{Fields: map[string]interface{}{"enabled": false}})
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to switch off referral capture", err.Error())
	}
}

func (r *referralSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "referral_settings")...)
}
