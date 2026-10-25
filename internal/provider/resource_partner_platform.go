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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &partnerPlatformResource{}
	_ resource.ResourceWithImportState = &partnerPlatformResource{}
	_ resource.ResourceWithConfigure   = &partnerPlatformResource{}
)

// NewPartnerPlatformResource is the factory for atlas_partner_platform.
func NewPartnerPlatformResource() resource.Resource { return &partnerPlatformResource{} }

type partnerPlatformResource struct {
	client *client.Client
}

type partnerPlatformModel struct {
	ID                  types.String `tfsdk:"id"`
	Partner             types.String `tfsdk:"partner"`
	Name                types.String `tfsdk:"name"`
	PlatformID          types.String `tfsdk:"platform_id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	JWKSURL             types.String `tfsdk:"jwks_url"`
	SignatureHeader     types.String `tfsdk:"signature_header"`
	SignatureScheme     types.String `tfsdk:"signature_scheme"`
	ToleranceSeconds    types.Int64  `tfsdk:"tolerance_seconds"`
	AllowedSources      types.Set    `tfsdk:"allowed_sources"`
	RecordConsents      types.Bool   `tfsdk:"record_consents"`
	TokenWO             types.String `tfsdk:"token_wo"`
	TokenWOVersion      types.Int64  `tfsdk:"token_wo_version"`
	APIBaseURL          types.String `tfsdk:"api_base_url"`
	Preset              types.Bool   `tfsdk:"preset"`
	ReceiverURL         types.String `tfsdk:"receiver_url"`
	ReceiverURLTemplate types.String `tfsdk:"receiver_url_template"`
	RegistrationStatus  types.String `tfsdk:"registration_status"`
	RegisteredAt        types.String `tfsdk:"registered_at"`
	TokenSet            types.Bool   `tfsdk:"token_set"`
	TokenPrefix         types.String `tfsdk:"token_prefix"`
	LastDeliveryAt      types.String `tfsdk:"last_delivery_at"`
	LastDeliveryError   types.String `tfsdk:"last_delivery_error"`
}

func (r *partnerPlatformResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_platform"
}

func (r *partnerPlatformResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func partnerPlatformSchema() schema.Schema {
	keepString := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	return schema.Schema{
		MarkdownDescription: "This instance as a **partner platform** of another product: a product that signs the events it " +
			"delivers to you (for example consent events) with a published Ed25519 key set. Atlas verifies each delivery against " +
			"the partner's key set and applies it to your users. One resource per partner. A preset partner fills in the defaults " +
			"on create; the partner token the partner hands back is **write-only** (`token_wo`): it is stored once, encrypted, " +
			"and never returned or kept in state. Changing a setting you removed from the configuration keeps the stored value. " +
			"Destroying the resource removes the platform and its stored token. `token_wo` needs Terraform 1.11 or later.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The Atlas id of the partner platform (same for the life of the resource).",
				Computed:            true,
				PlanModifiers:       keepString,
			},
			"partner": schema.StringAttribute{
				MarkdownDescription: "The partner: a preset id, or a short lower-case id of your own. Changing it replaces the resource.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A display name (1-80 characters). A preset has its own.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       keepString,
			},
			"platform_id": schema.StringAttribute{
				MarkdownDescription: "The lower-case id the partner knows this platform by. Required for a partner that is not a preset.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       keepString,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas accepts deliveries from the partner.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"jwks_url": schema.StringAttribute{
				MarkdownDescription: "The https URL of the partner's published key set. Required for a partner that is not a preset.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       keepString,
			},
			"signature_header": schema.StringAttribute{
				MarkdownDescription: "The header the partner sends its signature in. Required for a partner that is not a preset.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       keepString,
			},
			"signature_scheme": schema.StringAttribute{
				MarkdownDescription: "The signature scheme. Only `t_kid_ed25519` (`t=<unix>,kid=<key id>,ed25519=<signature>`) exists today.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       keepString,
			},
			"tolerance_seconds": schema.Int64Attribute{
				MarkdownDescription: "How far a delivery's timestamp may differ from now, 30 to 900 seconds.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"allowed_sources": schema.SetAttribute{
				MarkdownDescription: "The partner-side site ids whose deliveries are accepted (at most 50). A delivery from a site not listed is ignored (`source_not_allowed`), so list the sites that belong to you.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"record_consents": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas records the consents carried by deliveries on the user.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"token_wo": schema.StringAttribute{
				MarkdownDescription: "The partner token the partner issued to you. **Write-only**: never stored in state or returned by Atlas. " +
					"Sent on create and whenever `token_wo_version` changes (a later send replaces the stored token). Needs Terraform 1.11 or later.",
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"token_wo_version": schema.Int64Attribute{
				MarkdownDescription: "Any number; changing it sends `token_wo` again (and replaces the stored token).",
				Optional:            true,
			},
			"api_base_url": schema.StringAttribute{
				MarkdownDescription: "An https base URL for the partner's API, when it differs from the preset's. Sent together with the token; it takes effect when the token is sent.",
				Optional:            true,
			},
			"preset": schema.BoolAttribute{
				MarkdownDescription: "Whether the partner is a built-in preset.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"receiver_url": schema.StringAttribute{
				MarkdownDescription: "Where the partner delivers events for this instance. Give it to the partner to register.",
				Computed:            true,
				PlanModifiers:       keepString,
			},
			"receiver_url_template": schema.StringAttribute{
				MarkdownDescription: "The receiver URL with the partner's placeholder in it, when the partner registers per site.",
				Computed:            true,
				PlanModifiers:       keepString,
			},
			"registration_status": schema.StringAttribute{
				MarkdownDescription: "`unregistered` or `registered`.",
				Computed:            true,
			},
			"registered_at": schema.StringAttribute{
				MarkdownDescription: "When the partner registered the receiver (RFC 3339), if it has.",
				Computed:            true,
			},
			"token_set": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas holds a partner token.",
				Computed:            true,
			},
			"token_prefix": schema.StringAttribute{
				MarkdownDescription: "The first characters of the stored token, to tell tokens apart. Not the token.",
				Computed:            true,
			},
			"last_delivery_at": schema.StringAttribute{
				MarkdownDescription: "When the last delivery arrived (RFC 3339).",
				Computed:            true,
			},
			"last_delivery_error": schema.StringAttribute{
				MarkdownDescription: "The last error Atlas hit applying a delivery, if any.",
				Computed:            true,
			},
		},
	}
}

func (r *partnerPlatformResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = partnerPlatformSchema()
}

func fillPartnerPlatform(ctx context.Context, m *partnerPlatformModel, p *client.PartnerPlatform, diags *diag.Diagnostics) {
	m.ID = types.StringValue(p.ID)
	m.Partner = types.StringValue(p.Partner)
	m.Name = types.StringValue(p.Name)
	m.PlatformID = types.StringValue(p.PlatformID)
	m.Enabled = types.BoolValue(p.Enabled)
	m.JWKSURL = types.StringValue(p.JWKSURL)
	m.SignatureHeader = types.StringValue(p.SignatureHeader)
	m.SignatureScheme = types.StringValue(p.SignatureScheme)
	m.ToleranceSeconds = types.Int64Value(p.ToleranceSeconds)
	m.AllowedSources = stringSliceToSet(ctx, p.AllowedSources, diags)
	m.RecordConsents = types.BoolValue(p.RecordConsents)
	m.Preset = types.BoolValue(p.Preset)
	m.ReceiverURL = types.StringValue(p.ReceiverURL)
	m.ReceiverURLTemplate = strOrNull(p.ReceiverURLTemplate)
	m.RegistrationStatus = types.StringValue(p.Registration.Status)
	m.RegisteredAt = timeFromAPI(p.Registration.RegisteredAt, types.StringNull())
	m.TokenSet = types.BoolValue(p.TokenSet)
	m.TokenPrefix = stringPtrToValue(p.TokenPrefix)
	m.LastDeliveryAt = timeFromAPI(p.LastDeliveryAt, types.StringNull())
	m.LastDeliveryError = stringPtrToValue(p.LastDeliveryError)
}

// putBody is the PUT body: only what the configuration sets (the API keeps the
// rest, and a preset fills the rest on create).
func (r *partnerPlatformResource) putBody(ctx context.Context, cfg *partnerPlatformModel, diags *diag.Diagnostics) map[string]interface{} {
	body := map[string]interface{}{}
	str := func(key string, v types.String) {
		if !v.IsNull() && !v.IsUnknown() {
			body[key] = v.ValueString()
		}
	}
	str("name", cfg.Name)
	str("platform_id", cfg.PlatformID)
	str("jwks_url", cfg.JWKSURL)
	str("signature_header", cfg.SignatureHeader)
	str("signature_scheme", cfg.SignatureScheme)
	if !cfg.Enabled.IsNull() && !cfg.Enabled.IsUnknown() {
		body["enabled"] = cfg.Enabled.ValueBool()
	}
	if !cfg.RecordConsents.IsNull() && !cfg.RecordConsents.IsUnknown() {
		body["record_consents"] = cfg.RecordConsents.ValueBool()
	}
	if !cfg.ToleranceSeconds.IsNull() && !cfg.ToleranceSeconds.IsUnknown() {
		body["tolerance_seconds"] = cfg.ToleranceSeconds.ValueInt64()
	}
	if !cfg.AllowedSources.IsNull() && !cfg.AllowedSources.IsUnknown() {
		body["allowed_sources"] = setToStringSlice(ctx, cfg.AllowedSources, diags)
	}
	return body
}

// apply saves the settings, then stores the token when it is to be sent. prior
// is the state (nil on create); cfg carries the write-only token.
func (r *partnerPlatformResource) apply(ctx context.Context, plan, cfg, prior *partnerPlatformModel, diags *diag.Diagnostics) {
	partner := plan.Partner.ValueString()
	body := r.putBody(ctx, cfg, diags)
	if diags.HasError() {
		return
	}
	saved, err := r.client.PutPartnerPlatform(ctx, partner, body)
	if err != nil {
		diags.AddError("Unable to save partner platform "+partner, err.Error())
		return
	}
	sendToken := false
	if prior == nil {
		sendToken = !cfg.TokenWO.IsNull()
	} else if !plan.TokenWOVersion.Equal(prior.TokenWOVersion) {
		if cfg.TokenWO.IsNull() {
			diags.AddAttributeError(path.Root("token_wo"), "token_wo is required",
				"token_wo_version changed, which sends the partner token again, but token_wo is not set. A stored token cannot be cleared; destroy the resource to remove it.")
			return
		}
		sendToken = true
	}
	if sendToken {
		base := ""
		if !cfg.APIBaseURL.IsNull() && !cfg.APIBaseURL.IsUnknown() {
			base = cfg.APIBaseURL.ValueString()
		}
		rotate := saved.TokenSet
		saved, err = r.client.StorePartnerPlatformToken(ctx, partner, cfg.TokenWO.ValueString(), rotate, base)
		if err != nil {
			diags.AddError("Unable to store the token of partner platform "+partner, err.Error())
			return
		}
	}
	fillPartnerPlatform(ctx, plan, saved, diags)
}

func (r *partnerPlatformResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, cfg partnerPlatformModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.TokenWO = types.StringNull()
	r.apply(ctx, &plan, &cfg, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *partnerPlatformResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, cfg, state partnerPlatformModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.TokenWO = types.StringNull()
	r.apply(ctx, &plan, &cfg, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *partnerPlatformResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state partnerPlatformModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.client.GetPartnerPlatform(ctx, state.Partner.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read partner platform", err.Error())
		return
	}
	fillPartnerPlatform(ctx, &state, p, &resp.Diagnostics)
	state.TokenWO = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *partnerPlatformResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state partnerPlatformModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeletePartnerPlatform(ctx, state.Partner.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove partner platform", err.Error())
	}
}

func (r *partnerPlatformResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("partner"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
