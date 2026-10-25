package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &privacyOrchestratorResource{}
	_ resource.ResourceWithImportState    = &privacyOrchestratorResource{}
	_ resource.ResourceWithConfigure      = &privacyOrchestratorResource{}
	_ resource.ResourceWithValidateConfig = &privacyOrchestratorResource{}
	_ resource.ResourceWithModifyPlan     = &privacyOrchestratorResource{}
)

// NewPrivacyOrchestratorResource is the factory for atlas_privacy_orchestrator.
func NewPrivacyOrchestratorResource() resource.Resource { return &privacyOrchestratorResource{} }

type privacyOrchestratorResource struct {
	client *client.Client
}

type privacyOrchestratorModel struct {
	ID                        types.String `tfsdk:"id"`
	ProviderID                types.String `tfsdk:"provider_id"`
	Name                      types.String `tfsdk:"name"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	CaseURL                   types.String `tfsdk:"case_url"`
	CancelURL                 types.String `tfsdk:"cancel_url"`
	AuthHeader                types.String `tfsdk:"auth_header"`
	AuthScheme                types.String `tfsdk:"auth_scheme"`
	APIKeyWO                  types.String `tfsdk:"api_key_wo"`
	APIKeyWOVersion           types.Int64  `tfsdk:"api_key_wo_version"`
	UsePartnerLink            types.Bool   `tfsdk:"use_partner_link"`
	PartnerLinkID             types.String `tfsdk:"partner_link_id"`
	SiteID                    types.String `tfsdk:"site_id"`
	CancelDeletionOnSignIn    types.Bool   `tfsdk:"cancel_deletion_on_sign_in"`
	SignRequests              types.Bool   `tfsdk:"sign_requests"`
	OpenClosureCases          types.Bool   `tfsdk:"open_closure_cases"`
	CompletionWebhookURL      types.String `tfsdk:"completion_webhook_url"`
	APIKeySet                 types.Bool   `tfsdk:"api_key_set"`
	SigningSecret             types.String `tfsdk:"signing_secret"`
	CompletionWebhookSecret   types.String `tfsdk:"completion_webhook_secret"`
	CompletionWebhookEndpoint types.String `tfsdk:"completion_webhook_endpoint_id"`
	LastCaseAt                types.Int64  `tfsdk:"last_case_at"`
	LastError                 types.String `tfsdk:"last_error"`
}

func (r *privacyOrchestratorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privacy_orchestrator"
}

func (r *privacyOrchestratorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func privacyOrchestratorSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "The data-rights orchestrator — a singleton per instance. When a person's data request (erasure, " +
			"access, restriction, ...) reaches this instance, Atlas opens a case with the orchestrator (any provider; " +
			"`cookiemunch` is a built-in preset that only fills in defaults) and takes its completion events back. Secrets go " +
			"in and never come out: the orchestrator's `api_key` is write-only (`api_key_wo`); the outbound `signing_secret` " +
			"(`sign_requests = true`) and the `completion_webhook_secret` are minted by Atlas and shown once. Each apply sets " +
			"every field this resource manages. Destroying it removes the orchestrator (and its completion webhook). " +
			"Import with any id, e.g. `privacy_orchestrator`. `api_key_wo` needs Terraform 1.11 or later.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `privacy_orchestrator`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"provider_id": schema.StringAttribute{
				MarkdownDescription: "The orchestrator: a preset (`cookiemunch`) or a short lower-case id of your own " +
					"(letters, digits, `_` and `-`). Default `custom`. Sent as `provider`.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A display name (up to 80 characters). Required for a custom orchestrator; a preset has its own.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas opens cases with the orchestrator. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"case_url": schema.StringAttribute{
				MarkdownDescription: "The https URL Atlas POSTs a new case to. A preset has a default.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cancel_url": schema.StringAttribute{
				MarkdownDescription: "The https URL Atlas calls to cancel a case; `{id}` stands for the case id. A preset has a default.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"auth_header": schema.StringAttribute{
				MarkdownDescription: "The header Atlas sends the key in (e.g. `authorization`, `x-api-key`). Default `authorization`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"auth_scheme": schema.StringAttribute{
				MarkdownDescription: "`bearer` (default) sends `Bearer <key>`; `raw` sends the key alone.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"api_key_wo": schema.StringAttribute{
				MarkdownDescription: "The credential the orchestrator issued to Atlas (16-512 characters, no spaces). " +
					"**Write-only**: never stored in state or returned by Atlas. Sent on create and whenever " +
					"`api_key_wo_version` changes; change the version with `api_key_wo` unset to clear it. Needs Terraform 1.11 or later.",
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"api_key_wo_version": schema.Int64Attribute{
				MarkdownDescription: "Any number; changing it sends `api_key_wo` again (or clears the key when it is unset).",
				Optional:            true,
			},
			"use_partner_link": schema.BoolAttribute{
				MarkdownDescription: "Authenticate with the outbound partner link to this provider (`atlas_outbound_partner_link`) " +
					"instead of `api_key_wo`. Default `false`; on when `partner_link_id` is set.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"partner_link_id": schema.StringAttribute{
				MarkdownDescription: "The id of the partner link the orchestrator's owner granted this instance " +
					"(letters, digits and `_ . : -`, at most 128). Setting it makes Atlas call the orchestrator with that grant " +
					"(an `X-Partner-Link` header) and turns on `use_partner_link`. Unset clears it.",
				Optional: true,
			},
			"site_id": schema.StringAttribute{
				MarkdownDescription: "The orchestrator-side site id a case is for (letters, digits and `_ . : -`, at most 128). " +
					"Unset: Atlas falls back to the site of a matching integration or outbound partner link, and a case with none is abandoned with an error naming `site_id`.",
				Optional: true,
			},
			"cancel_deletion_on_sign_in": schema.BoolAttribute{
				MarkdownDescription: "Cancel a pending account deletion when the person signs in again " +
					"(the instance setting `compliance.retention.cancelDeletionOnSignIn`). Left unset, the current setting is kept.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sign_requests": schema.BoolAttribute{
				MarkdownDescription: "Sign the cases Atlas opens with an `Atlas-Signature` header. Switching it on mints " +
					"`signing_secret` (shown once). Default `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"open_closure_cases": schema.BoolAttribute{
				MarkdownDescription: "Also open a case when an account is closed. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"completion_webhook_url": schema.StringAttribute{
				MarkdownDescription: "An https URL on your side that receives the orchestrator's completion events. Atlas manages " +
					"a webhook endpoint for it. Unset removes it.",
				Optional: true,
			},
			"api_key_set": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas holds an api key for the orchestrator.",
				Computed:            true,
			},
			"signing_secret": schema.StringAttribute{
				MarkdownDescription: "Verifies the `Atlas-Signature` on the cases Atlas opens. Shown once, when `sign_requests` is switched on.",
				Computed:            true,
				Sensitive:           true,
			},
			"completion_webhook_secret": schema.StringAttribute{
				MarkdownDescription: "Verifies the completion events sent to `completion_webhook_url`. Shown once, when the webhook is created.",
				Computed:            true,
				Sensitive:           true,
			},
			"completion_webhook_endpoint_id": schema.StringAttribute{
				MarkdownDescription: "The id of the managed webhook endpoint for `completion_webhook_url`.",
				Computed:            true,
			},
			"last_case_at": schema.Int64Attribute{
				MarkdownDescription: "When Atlas last opened a case (epoch ms).",
				Computed:            true,
			},
			"last_error": schema.StringAttribute{
				MarkdownDescription: "The last error Atlas hit calling the orchestrator, if any.",
				Computed:            true,
			},
		},
	}
}

func (r *privacyOrchestratorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = privacyOrchestratorSchema()
}

// ValidateConfig checks the api key length at plan time.
func (r *privacyOrchestratorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var m privacyOrchestratorModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !m.APIKeyWO.IsNull() && !m.APIKeyWO.IsUnknown() {
		n := len(m.APIKeyWO.ValueString())
		if n < 16 || n > 512 {
			resp.Diagnostics.AddAttributeError(path.Root("api_key_wo"), "Invalid api_key_wo",
				"The orchestrator's api key must be 16-512 characters.")
		}
	}
	if !m.PartnerLinkID.IsNull() && !m.PartnerLinkID.IsUnknown() && !m.UsePartnerLink.IsNull() && !m.UsePartnerLink.IsUnknown() && !m.UsePartnerLink.ValueBool() {
		resp.Diagnostics.AddAttributeError(path.Root("use_partner_link"), "Conflicting settings",
			"partner_link_id turns use_partner_link on; remove use_partner_link = false or remove partner_link_id.")
	}
	if !m.AuthScheme.IsNull() && !m.AuthScheme.IsUnknown() && m.AuthScheme.ValueString() != "bearer" && m.AuthScheme.ValueString() != "raw" {
		resp.Diagnostics.AddAttributeError(path.Root("auth_scheme"), "Invalid auth_scheme", "auth_scheme must be bearer or raw.")
	}
}

// ModifyPlan follows the server: setting partner_link_id turns use_partner_link
// on, so a plan that leaves use_partner_link unset says so.
func (r *privacyOrchestratorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var cfg, plan privacyOrchestratorModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !cfg.PartnerLinkID.IsNull() && !cfg.PartnerLinkID.IsUnknown() && cfg.UsePartnerLink.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("use_partner_link"), types.BoolValue(true))...)
	}
}

func fillPrivacyOrchestrator(m *privacyOrchestratorModel, o *client.PrivacyOrchestrator) {
	m.ID = types.StringValue("privacy_orchestrator")
	m.ProviderID = stringPtrToValue(o.Provider)
	m.Name = stringPtrToValue(o.Name)
	m.Enabled = types.BoolValue(o.Enabled)
	m.CaseURL = stringPtrToValue(o.CaseURL)
	m.CancelURL = stringPtrToValue(o.CancelURL)
	m.AuthHeader = types.StringValue(o.AuthHeader)
	m.AuthScheme = types.StringValue(o.AuthScheme)
	m.UsePartnerLink = types.BoolValue(o.UsePartnerLink)
	m.PartnerLinkID = stringPtrToValue(o.PartnerLinkID)
	m.SiteID = stringPtrToValue(o.SiteID)
	if o.CancelDeletionOnSignIn != nil {
		m.CancelDeletionOnSignIn = types.BoolValue(*o.CancelDeletionOnSignIn)
	} else {
		m.CancelDeletionOnSignIn = types.BoolValue(false)
	}
	m.SignRequests = types.BoolValue(o.SigningSecretSet)
	m.OpenClosureCases = types.BoolValue(o.OpenClosureCases)
	m.APIKeySet = types.BoolValue(o.APIKeySet)
	if o.CompletionWebhook != nil {
		m.CompletionWebhookURL = types.StringValue(o.CompletionWebhook.URL)
		m.CompletionWebhookEndpoint = types.StringValue(o.CompletionWebhook.WebhookEndpointID)
	} else {
		m.CompletionWebhookURL = types.StringNull()
		m.CompletionWebhookEndpoint = types.StringNull()
	}
	m.LastCaseAt = int64PtrToValue(o.LastCaseAt)
	m.LastError = stringPtrToValue(o.LastError)
}

// apply PUTs the planned settings. prior is the state (nil on create); config
// carries the write-only key.
func (r *privacyOrchestratorResource) apply(ctx context.Context, plan, config *privacyOrchestratorModel, prior *privacyOrchestratorModel, diags *diag.Diagnostics) {
	body := map[string]interface{}{
		"enabled":            plan.Enabled.ValueBool(),
		"use_partner_link":   plan.UsePartnerLink.ValueBool(),
		"open_closure_cases": plan.OpenClosureCases.ValueBool(),
	}
	put := func(key string, v types.String) {
		if !v.IsNull() && !v.IsUnknown() {
			body[key] = v.ValueString()
		}
	}
	put("provider", config.ProviderID)
	put("name", config.Name)
	put("case_url", config.CaseURL)
	put("cancel_url", config.CancelURL)
	put("auth_header", config.AuthHeader)
	put("auth_scheme", config.AuthScheme)

	// partner_link_id and site_id: a value sets it, removing it clears it.
	for key, v := range map[string]types.String{"partner_link_id": config.PartnerLinkID, "site_id": config.SiteID} {
		switch {
		case !v.IsNull() && !v.IsUnknown():
			body[key] = v.ValueString()
		case prior != nil:
			body[key] = nil
		}
	}
	if !config.CancelDeletionOnSignIn.IsNull() && !config.CancelDeletionOnSignIn.IsUnknown() {
		body["cancel_deletion_on_sign_in"] = config.CancelDeletionOnSignIn.ValueBool()
	}
	// The signing secret is minted whenever true is sent, so send it only when the
	// setting changes (or on create), never on every apply.
	if prior == nil || !plan.SignRequests.Equal(prior.SignRequests) {
		if prior != nil || plan.SignRequests.ValueBool() {
			body["sign_requests"] = plan.SignRequests.ValueBool()
		}
	}
	// The completion webhook: a URL creates or updates it, null removes it.
	if prior == nil || !plan.CompletionWebhookURL.Equal(prior.CompletionWebhookURL) {
		if !plan.CompletionWebhookURL.IsNull() {
			body["completion_webhook_url"] = plan.CompletionWebhookURL.ValueString()
		} else if prior != nil {
			body["completion_webhook_url"] = nil
		}
	}
	switch {
	case prior == nil:
		if !config.APIKeyWO.IsNull() {
			body["api_key"] = config.APIKeyWO.ValueString()
		}
	case !plan.APIKeyWOVersion.Equal(prior.APIKeyWOVersion):
		if config.APIKeyWO.IsNull() {
			body["api_key"] = nil
		} else {
			body["api_key"] = config.APIKeyWO.ValueString()
		}
	}
	out, err := r.client.PutPrivacyOrchestrator(ctx, body)
	if err != nil {
		diags.AddError("Unable to set the privacy orchestrator", err.Error())
		return
	}
	fillPrivacyOrchestrator(plan, out)
	// The secrets are shown once, on the write that created them; otherwise keep
	// what state already holds.
	plan.SigningSecret = secretOrPrior(out.SigningSecret, prior, func(m *privacyOrchestratorModel) types.String { return m.SigningSecret })
	plan.CompletionWebhookSecret = secretOrPrior(out.CompletionWebhookSecret, prior, func(m *privacyOrchestratorModel) types.String { return m.CompletionWebhookSecret })
}

// secretOrPrior returns the freshly minted secret, else the prior state's.
func secretOrPrior(fresh *string, prior *privacyOrchestratorModel, pick func(*privacyOrchestratorModel) types.String) types.String {
	if fresh != nil {
		return types.StringValue(*fresh)
	}
	if prior != nil {
		if v := pick(prior); !v.IsNull() && !v.IsUnknown() {
			return v
		}
	}
	return types.StringNull()
}

func (r *privacyOrchestratorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config privacyOrchestratorModel
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

func (r *privacyOrchestratorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state privacyOrchestratorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	o, err := r.client.GetPrivacyOrchestrator(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the privacy orchestrator", err.Error())
		return
	}
	if !o.Configured {
		resp.State.RemoveResource(ctx)
		return
	}
	prior := state
	fillPrivacyOrchestrator(&state, o)
	state.SigningSecret = prior.SigningSecret
	state.CompletionWebhookSecret = prior.CompletionWebhookSecret
	if state.SigningSecret.IsUnknown() {
		state.SigningSecret = types.StringNull()
	}
	if state.CompletionWebhookSecret.IsUnknown() {
		state.CompletionWebhookSecret = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *privacyOrchestratorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state privacyOrchestratorModel
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

func (r *privacyOrchestratorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DeletePrivacyOrchestrator(ctx); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove the privacy orchestrator", err.Error())
	}
}

func (r *privacyOrchestratorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "privacy_orchestrator")...)
}
