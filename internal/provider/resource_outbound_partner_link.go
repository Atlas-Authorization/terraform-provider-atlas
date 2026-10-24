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
	_ resource.Resource                = &outboundPartnerLinkResource{}
	_ resource.ResourceWithImportState = &outboundPartnerLinkResource{}
	_ resource.ResourceWithConfigure   = &outboundPartnerLinkResource{}
)

// NewOutboundPartnerLinkResource is the factory for atlas_outbound_partner_link.
func NewOutboundPartnerLinkResource() resource.Resource { return &outboundPartnerLinkResource{} }

type outboundPartnerLinkResource struct {
	client *client.Client
}

type outboundPartnerLinkModel struct {
	ID           types.String `tfsdk:"id"`
	Partner      types.String `tfsdk:"partner"`
	Name         types.String `tfsdk:"name"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Scopes       types.Set    `tfsdk:"scopes"`
	Resource     types.String `tfsdk:"resource"`
	BaseURL      types.String `tfsdk:"base_url"`
	TokenURL     types.String `tfsdk:"token_url"`
	AuthMethod   types.String `tfsdk:"auth_method"`
	Status       types.String `tfsdk:"status"`
	LastError    types.String `tfsdk:"last_error"`
}

func (r *outboundPartnerLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_outbound_partner_link"
}

func (r *outboundPartnerLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func outboundPartnerLinkSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "An OUTBOUND partner link: the credential a partner (e.g. Cookie Munch, or your own " +
			"privacy orchestrator) granted this instance, so Atlas can call the partner. One link per `partner`; each " +
			"apply stores every field. The `client_secret` is stored (sensitive) in state and sent on every apply; Atlas " +
			"never returns it. Destroy removes the credential. Import with the partner id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The stored link's id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"partner": schema.StringAttribute{
				MarkdownDescription: "The partner's id (e.g. `cookiemunch`). Changing it replaces the link.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A display name. Presets have a default.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The client id the partner issued to Atlas.",
				Required:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The client secret (or bearer secret) the partner issued to Atlas.",
				Required:            true,
				Sensitive:           true,
			},
			"scopes": schema.SetAttribute{
				MarkdownDescription: "The partner-side scopes Atlas asks for, e.g. `[\"dsar:write\"]`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringSet(),
			},
			"resource": schema.StringAttribute{
				MarkdownDescription: "The partner resource (audience) Atlas requests tokens for, if the partner needs one.",
				Optional:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The partner's API base URL. Presets have a default; any other partner needs one.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"token_url": schema.StringAttribute{
				MarkdownDescription: "Where Atlas gets tokens (default `<base_url>/v1/oauth/token`).",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "`client_credentials` (default) or `bearer_secret`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "`active` unless the partner revoked or the credential expired.",
				Computed:            true,
			},
			"last_error": schema.StringAttribute{
				MarkdownDescription: "The last error Atlas hit calling the partner, if any.",
				Computed:            true,
			},
		},
	}
}

func (r *outboundPartnerLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = outboundPartnerLinkSchema()
}

func fillOutbound(ctx context.Context, m *outboundPartnerLinkModel, l *client.OutboundPartnerLink, diags *diag.Diagnostics) {
	m.ID = types.StringValue(l.ID)
	m.Partner = types.StringValue(l.Partner)
	m.Name = types.StringValue(l.Name)
	m.ClientID = types.StringValue(l.ClientID)
	m.Scopes = stringSliceToSet(ctx, l.Scopes, diags)
	m.Resource = stringPtrToValue(l.Resource)
	m.BaseURL = stringPtrToValue(l.BaseURL)
	m.TokenURL = stringPtrToValue(l.TokenURL)
	m.AuthMethod = types.StringValue(l.AuthMethod)
	m.Status = types.StringValue(l.Status)
	m.LastError = stringPtrToValue(l.LastError)
}

func (r *outboundPartnerLinkResource) write(ctx context.Context, plan *outboundPartnerLinkModel, diags *diag.Diagnostics) *client.OutboundPartnerLink {
	body := client.OutboundPartnerLinkWrite{
		Name:         optionalString(plan.Name),
		ClientID:     plan.ClientID.ValueString(),
		ClientSecret: plan.ClientSecret.ValueString(),
		Scopes:       setToStringSlice(ctx, plan.Scopes, diags),
		Resource:     optionalString(plan.Resource),
		BaseURL:      optionalString(plan.BaseURL),
		TokenURL:     optionalString(plan.TokenURL),
		AuthMethod:   optionalString(plan.AuthMethod),
	}
	if body.Scopes == nil {
		body.Scopes = []string{}
	}
	if diags.HasError() {
		return nil
	}
	saved, err := r.client.PutOutboundPartnerLink(ctx, plan.Partner.ValueString(), body)
	if err != nil {
		diags.AddError("Unable to save outbound partner link", err.Error())
		return nil
	}
	return saved
}

func (r *outboundPartnerLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan outboundPartnerLinkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	saved := r.write(ctx, &plan, &resp.Diagnostics)
	if saved == nil {
		return
	}
	fillOutbound(ctx, &plan, saved, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *outboundPartnerLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state outboundPartnerLinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	l, err := r.client.GetOutboundPartnerLink(ctx, state.Partner.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read outbound partner link", err.Error())
		return
	}
	fillOutbound(ctx, &state, l, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *outboundPartnerLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan outboundPartnerLinkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	saved := r.write(ctx, &plan, &resp.Diagnostics)
	if saved == nil {
		return
	}
	fillOutbound(ctx, &plan, saved, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *outboundPartnerLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state outboundPartnerLinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOutboundPartnerLink(ctx, state.Partner.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove outbound partner link", err.Error())
	}
}

// ImportState takes the partner id; Read fills in the rest. The client_secret
// is not recoverable, so an imported link needs it set in configuration.
func (r *outboundPartnerLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("partner"), req.ID)...)
}
