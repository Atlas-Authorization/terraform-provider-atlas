package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &integrationResource{}
	_ resource.ResourceWithImportState = &integrationResource{}
	_ resource.ResourceWithConfigure   = &integrationResource{}
)

// Identity integrations: any service that verifies the signed-in user with a JWT.
// atlas_integration is the generic resource (a preset id or a custom id of your
// own); atlas_chatly_integration, atlas_forumulate_integration,
// atlas_cookiemunch_integration and atlas_referalli_integration are the same
// connection with the preset fixed and the custom-only fields removed.

// integrationPresets lists the built-in preset ids and the Terraform resource
// each gets.
var integrationPresets = []struct{ ID, Hint string }{
	{"chatly", "the chat channel public id, e.g. `chatly:support`"},
	{"forumulate", "the Forumulate organization id, e.g. `forumulate:org_123`"},
	{"cookiemunch", "the Cookie Munch site id (cbid), e.g. `cookiemunch:dv-abc`"},
	{"referalli", "the Referalli organization id, e.g. `referalli:org_123`"},
}

// NewIntegrationResource is the factory for atlas_integration (generic).
func NewIntegrationResource() resource.Resource { return &integrationResource{} }

// NewPresetIntegrationResource returns the factory for a preset resource.
func NewPresetIntegrationResource(preset string) func() resource.Resource {
	return func() resource.Resource { return &integrationResource{preset: preset} }
}

type integrationResource struct {
	// preset is "" for the generic resource, else the fixed preset id.
	preset string
	client *client.Client
}

type integrationModel struct {
	ID                      types.String `tfsdk:"id"`
	Integration             types.String `tfsdk:"integration"`
	Name                    types.String `tfsdk:"name"`
	Role                    types.String `tfsdk:"role"`
	Audience                types.String `tfsdk:"audience"`
	AudienceFormat          types.String `tfsdk:"audience_format"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	PrincipalClaim          types.String `tfsdk:"principal_claim"`
	ExtClaim                types.String `tfsdk:"ext_claim"`
	TTLSeconds              types.Int64  `tfsdk:"ttl_seconds"`
	EventsURL               types.String `tfsdk:"events_url"`
	BackchannelLogoutURL    types.String `tfsdk:"backchannel_logout_url"`
	VerifyErrorReportURL    types.String `tfsdk:"verify_error_report_url"`
	OAuthClientID           types.String `tfsdk:"oauth_client_id"`
	AtlasIDUsers            types.Bool   `tfsdk:"atlas_id_users"`
	ExternalRef             types.String `tfsdk:"external_ref"`
	Metadata                types.Map    `tfsdk:"metadata"`
	Preset                  types.String `tfsdk:"preset"`
	Issuer                  types.String `tfsdk:"issuer"`
	JWKSURL                 types.String `tfsdk:"jwks_url"`
	BackchannelLogoutSource types.String `tfsdk:"backchannel_logout_source"`
}

func (r *integrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.preset != "" {
		resp.TypeName = req.ProviderTypeName + "_" + r.preset + "_integration"
		return
	}
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *integrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func presetHint(preset string) string {
	for _, p := range integrationPresets {
		if p.ID == preset {
			return p.Hint
		}
	}
	return ""
}

func integrationSchema(preset string) schema.Schema {
	desc := "A custom or preset identity integration: Atlas mints short-lived signed tokens (`getToken({ for })`) " +
		"for the signed-in user that the service verifies with Atlas's JWKS, and can send it events and back-channel " +
		"logout. One resource per integration id; each apply stores the whole connection. Destroy removes it."
	audienceDesc := "The `aud` of the tokens: 1-200 printable characters with no spaces."
	var integration, name, role, audienceFormat schema.Attribute
	if preset == "" {
		desc += " This is the generic resource: `integration` is a preset id (`chatly`, `forumulate`, `cookiemunch`, " +
			"`referalli`) or a custom id of your own. See also `atlas_chatly_integration`, `atlas_forumulate_integration`, " +
			"`atlas_cookiemunch_integration` and `atlas_referalli_integration`."
		integration = schema.StringAttribute{
			MarkdownDescription: "The integration id. A preset id, or a custom id (lowercase letters, digits, `-`, `_`). Changing it replaces the connection.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		}
		name = schema.StringAttribute{
			MarkdownDescription: "A display name (1-80 characters). Required for a custom integration; a preset has its own.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
		role = schema.StringAttribute{
			MarkdownDescription: "What the service is: `support_chat`, `docs_community`, `consent`, `referrals` or `other` " +
				"(default for a custom integration; a preset has its own).",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
		audienceFormat = schema.StringAttribute{
			MarkdownDescription: "A custom integration only: a regular expression (up to 200 characters) the `audience` must match.",
			Optional:            true,
		}
	} else {
		desc = "The " + preset + " identity integration (a preset of `atlas_integration`). " + desc[len("A custom or preset identity integration: "):]
		audienceDesc = "The `aud` of the tokens: `" + preset + ":<id>`, where `<id>` is " + presetHint(preset) + "."
		integration = schema.StringAttribute{
			MarkdownDescription: "Always `" + preset + "`.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
		name = schema.StringAttribute{Computed: true, MarkdownDescription: "The preset's display name.",
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}
		role = schema.StringAttribute{Computed: true, MarkdownDescription: "The preset's role.",
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}
		audienceFormat = schema.StringAttribute{Computed: true, MarkdownDescription: "Not used by a preset.",
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}
	}
	return schema.Schema{
		MarkdownDescription: desc,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The integration id (same as `integration`).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"integration":     integration,
			"name":            name,
			"role":            role,
			"audience_format": audienceFormat,
			"audience": schema.StringAttribute{
				MarkdownDescription: audienceDesc,
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether tokens are minted and events sent. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"principal_claim": schema.StringAttribute{
				MarkdownDescription: "`audience_hint` (default) puts the principal claim on the token; `none` leaves it out.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("audience_hint"),
			},
			"ext_claim": schema.StringAttribute{
				MarkdownDescription: "`external_id` (default) puts the user's external id on the token; `none` leaves it out.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("external_id"),
			},
			"ttl_seconds": schema.Int64Attribute{
				MarkdownDescription: "Token lifetime, 10 to 120 seconds. Atlas's default (120) when unset.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"events_url": schema.StringAttribute{
				MarkdownDescription: "An https URL that receives signed user events. A preset has a default; leave unset to keep it.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"backchannel_logout_url": schema.StringAttribute{
				MarkdownDescription: "An https URL that receives OIDC back-channel logout tokens. A preset has a default; " +
					"otherwise the first-party OAuth client's `backchannel_logout_uri` is used.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"verify_error_report_url": schema.StringAttribute{
				MarkdownDescription: "An https URL the service reports token verification errors to (`POST /v1/integrations/verify_errors`).",
				Optional:            true,
			},
			"oauth_client_id": schema.StringAttribute{
				MarkdownDescription: "The `client_id` of the service's first-party OAuth client, whose back-channel logout " +
					"URI receives logout tokens when `backchannel_logout_url` is unset.",
				Optional: true,
			},
			"atlas_id_users": schema.BoolAttribute{
				MarkdownDescription: "Also mint tokens for users who signed in with Sign in with Atlas (a second trusted issuer). Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"external_ref": schema.StringAttribute{
				MarkdownDescription: "Your own stable id for the connection (1-128 printable characters, no spaces). One ref names one integration.",
				Optional:            true,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Tags on the connection (lowercase keys; string values).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringMap(),
			},
			"preset": schema.StringAttribute{
				MarkdownDescription: "The preset id, or null for a custom integration.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"issuer": schema.StringAttribute{
				MarkdownDescription: "The token issuer (`iss`) the service verifies against.",
				Computed:            true,
			},
			"jwks_url": schema.StringAttribute{
				MarkdownDescription: "Where the service fetches the keys to verify the tokens.",
				Computed:            true,
			},
			"backchannel_logout_source": schema.StringAttribute{
				MarkdownDescription: "Where the logout receiver comes from: `integration`, `oauth_client` or `preset`.",
				Computed:            true,
			},
		},
	}
}

func (r *integrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = integrationSchema(r.preset)
}

func claimFromAPI(v *string) types.String {
	if v == nil {
		return types.StringValue("none")
	}
	return types.StringValue(*v)
}

func claimToAPI(v types.String) interface{} {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "none" {
		return nil
	}
	return v.ValueString()
}

func fillIntegration(ctx context.Context, m *integrationModel, i *client.Integration, diags *diag.Diagnostics) {
	m.ID = types.StringValue(i.ID)
	m.Integration = types.StringValue(i.ID)
	m.Name = stringPtrToValue(i.Name)
	m.Role = stringPtrToValue(i.Role)
	m.Audience = stringPtrToValue(i.Audience)
	if m.AudienceFormat.IsUnknown() {
		m.AudienceFormat = types.StringNull()
	}
	m.Enabled = types.BoolValue(i.Enabled)
	m.PrincipalClaim = claimFromAPI(i.PrincipalClaim)
	m.ExtClaim = claimFromAPI(i.ExtClaim)
	m.TTLSeconds = int64PtrToValue(i.TTLSeconds)
	m.EventsURL = stringPtrToValue(i.EventsURL)
	m.BackchannelLogoutURL = stringPtrToValue(i.BackchannelLogoutURL)
	m.VerifyErrorReportURL = stringPtrToValue(i.VerifyErrorReportURL)
	m.OAuthClientID = stringPtrToValue(i.OAuthClientID)
	if i.AtlasIDUsers != nil {
		m.AtlasIDUsers = types.BoolValue(*i.AtlasIDUsers)
	} else {
		m.AtlasIDUsers = types.BoolValue(false)
	}
	m.ExternalRef = stringPtrToValue(i.ExternalRef)
	m.Metadata = metadataFromAPI(ctx, i.Metadata, diags)
	m.Preset = stringPtrToValue(i.Preset)
	m.Issuer = stringPtrToValue(i.Issuer)
	m.JWKSURL = stringPtrToValue(i.JWKSURL)
	m.BackchannelLogoutSource = stringPtrToValue(i.BackchannelLogoutSource)
}

// body builds the PUT body. Fields with a plan default come from the plan; the
// ones Atlas fills in itself (name, role, ttl, URLs) are sent only when the
// configuration sets them, so an omitted value keeps following Atlas's default.
func (r *integrationResource) body(ctx context.Context, plan, cfg *integrationModel, diags *diag.Diagnostics) map[string]interface{} {
	body := map[string]interface{}{
		"audience":        plan.Audience.ValueString(),
		"enabled":         plan.Enabled.ValueBool(),
		"principal_claim": claimToAPI(plan.PrincipalClaim),
		"ext_claim":       claimToAPI(plan.ExtClaim),
		"atlas_id_users":  plan.AtlasIDUsers.ValueBool(),
	}
	put := func(key string, v types.String) {
		if !v.IsNull() && !v.IsUnknown() {
			body[key] = v.ValueString()
		}
	}
	put("events_url", cfg.EventsURL)
	put("backchannel_logout_url", cfg.BackchannelLogoutURL)
	put("verify_error_report_url", cfg.VerifyErrorReportURL)
	put("oauth_client_id", cfg.OAuthClientID)
	put("external_ref", cfg.ExternalRef)
	if !cfg.TTLSeconds.IsNull() && !cfg.TTLSeconds.IsUnknown() {
		body["ttl_seconds"] = cfg.TTLSeconds.ValueInt64()
	}
	if r.preset == "" {
		put("name", cfg.Name)
		put("role", cfg.Role)
		put("audience_format", cfg.AudienceFormat)
	}
	if md := metadataToAPI(ctx, plan.Metadata, diags); md != nil {
		body["metadata"] = md
	}
	return body
}

func (r *integrationResource) id(m *integrationModel) string {
	if r.preset != "" {
		return r.preset
	}
	return m.Integration.ValueString()
}

func (r *integrationResource) apply(ctx context.Context, plan, cfg *integrationModel, diags *diag.Diagnostics) bool {
	id := r.id(plan)
	body := r.body(ctx, plan, cfg, diags)
	if diags.HasError() {
		return false
	}
	saved, err := r.client.PutIntegration(ctx, id, body)
	if err != nil {
		diags.AddError("Unable to save integration "+id, err.Error())
		return false
	}
	audienceFormat := plan.AudienceFormat
	fillIntegration(ctx, plan, saved, diags)
	plan.AudienceFormat = audienceFormat
	if r.preset != "" {
		plan.AudienceFormat = types.StringNull()
	}
	return !diags.HasError()
}

func (r *integrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, cfg integrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !r.apply(ctx, &plan, &cfg, &resp.Diagnostics) {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *integrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, cfg integrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !r.apply(ctx, &plan, &cfg, &resp.Diagnostics) {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *integrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state integrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	i, err := r.client.GetIntegration(ctx, r.id(&state))
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read integration", err.Error())
		return
	}
	if !i.Connected {
		resp.State.RemoveResource(ctx)
		return
	}
	audienceFormat := state.AudienceFormat
	fillIntegration(ctx, &state, i, &resp.Diagnostics)
	state.AudienceFormat = audienceFormat
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *integrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state integrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteIntegration(ctx, r.id(&state)); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove integration", err.Error())
	}
}

func (r *integrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	if r.preset != "" {
		id = r.preset
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("integration"), id)...)
	// The unset optional fields Read does not return stay null.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("audience_format"), types.StringNull())...)
}
