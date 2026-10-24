package provider

import (
	"context"
	"net/http"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &partnerLinkResource{}
	_ resource.ResourceWithImportState = &partnerLinkResource{}
	_ resource.ResourceWithConfigure   = &partnerLinkResource{}
	_ resource.ResourceWithModifyPlan  = &partnerLinkResource{}
)

// NewPartnerLinkResource is the factory for atlas_partner_link.
func NewPartnerLinkResource() resource.Resource { return &partnerLinkResource{} }

type partnerLinkResource struct {
	client *client.Client
}

type partnerLinkModel struct {
	ID                  types.String `tfsdk:"id"`
	Partner             types.String `tfsdk:"partner"`
	Name                types.String `tfsdk:"name"`
	Scopes              types.Set    `tfsdk:"scopes"`
	ExpiresAt           types.String `tfsdk:"expires_at"`
	CertFingerprints    types.Set    `tfsdk:"cert_fingerprints"`
	RotateSecretTrigger types.String `tfsdk:"rotate_secret_trigger"`
	ClientID            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	TokenEndpoint       types.String `tfsdk:"token_endpoint"`
	Status              types.String `tfsdk:"status"`
	CreatedAt           types.Int64  `tfsdk:"created_at"`
}

func (r *partnerLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_link"
}

func (r *partnerLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func partnerLinkSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "An INBOUND partner link: a credential that lets another product (any partner id you choose; " +
			"`cookiemunch`, `chatly`, `forumulate` and `referalli` are presets) call this instance with exactly the scopes " +
			"granted. The `client_secret` is returned once and kept (sensitive) in state; give the partner the `client_id` " +
			"and secret, or let it mint tokens at `token_endpoint` with `grant_type=client_credentials`. `partner`, `name`, " +
			"`scopes` and `expires_at` cannot change (changing one replaces the link); `cert_fingerprints` (an mTLS lock) " +
			"updates in place, and `rotate_secret_trigger` rotates the secret in place. Destroy revokes the link. " +
			"Granting scopes needs scopes you already hold.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The link id (also its `client_id`).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"partner": schema.StringAttribute{
				MarkdownDescription: "The partner's id: 2-40 lowercase letters, digits, `-` or `_` (e.g. `cookiemunch`).",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A display name (up to 100 characters). Presets have a default.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scopes": schema.SetAttribute{
				MarkdownDescription: "The scopes the partner may use, e.g. `[\"privacy:execute\", \"users:read\"]`. `*` is not allowed.",
				Required:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "When the link stops working (RFC 3339). Defaults to one year from creation; at most one year away.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cert_fingerprints": schema.SetAttribute{
				MarkdownDescription: "mTLS: SHA-256 fingerprints (base64url) of the client certificates every use must present. " +
					"Empty (the default) means no certificate is needed.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     emptyStringSet(),
			},
			"rotate_secret_trigger": schema.StringAttribute{
				MarkdownDescription: "Any value; changing it rotates the client secret in place (the new one lands in " +
					"`client_secret`). The previous secret stops working at once.",
				Optional: true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The client id the partner presents.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The client secret, shown once at creation and on rotation.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"token_endpoint": schema.StringAttribute{
				MarkdownDescription: "Where the partner exchanges its credentials for a short-lived token.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "`active`, `expired` or `revoked`.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch ms).",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *partnerLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = partnerLinkSchema()
}

func fillPartnerLink(ctx context.Context, m *partnerLinkModel, l *client.PartnerLink, diags *diag.Diagnostics) {
	m.ID = types.StringValue(l.ID)
	m.Partner = types.StringValue(l.Partner)
	m.Name = types.StringValue(l.Name)
	m.Scopes = stringSliceToSet(ctx, l.Scopes, diags)
	m.ExpiresAt = timeFromAPI(l.ExpiresAt, m.ExpiresAt)
	m.CertFingerprints = stringSliceToSet(ctx, l.CertFingerprints, diags)
	m.ClientID = types.StringValue(l.ClientID)
	if l.ClientSecret != nil {
		m.ClientSecret = types.StringValue(*l.ClientSecret)
	} else if m.ClientSecret.IsUnknown() {
		m.ClientSecret = types.StringNull()
	}
	m.TokenEndpoint = types.StringValue(l.TokenEndpoint)
	m.Status = types.StringValue(l.Status)
	m.CreatedAt = types.Int64Value(l.CreatedAt)
}

func (r *partnerLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan partnerLinkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.PartnerLinkCreate{
		Partner:          plan.Partner.ValueString(),
		Name:             optionalString(plan.Name),
		Scopes:           setToStringSlice(ctx, plan.Scopes, &resp.Diagnostics),
		ExpiresAt:        optionalString(plan.ExpiresAt),
		CertFingerprints: setToStringSlice(ctx, plan.CertFingerprints, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreatePartnerLink(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create partner link", err.Error())
		return
	}
	fillPartnerLink(ctx, &plan, created, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *partnerLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state partnerLinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	l, err := r.client.GetPartnerLink(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read partner link", err.Error())
		return
	}
	if l.RevokedAt != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	fillPartnerLink(ctx, &state, l, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *partnerLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state partnerLinkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	plan.ClientSecret = state.ClientSecret
	current := &client.PartnerLink{}
	if !plan.CertFingerprints.Equal(state.CertFingerprints) {
		fp := setToStringSlice(ctx, plan.CertFingerprints, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		updated, err := r.client.SetPartnerLinkCertFingerprints(ctx, id, fp)
		if err != nil {
			resp.Diagnostics.AddError("Unable to update partner link", err.Error())
			return
		}
		current = updated
	}
	if !plan.RotateSecretTrigger.Equal(state.RotateSecretTrigger) {
		rotated, err := r.client.RotatePartnerLinkSecret(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Unable to rotate partner link secret", err.Error())
			return
		}
		current = rotated
	}
	if current.ID == "" {
		// Nothing the API needed (only a trigger-free no-op); read it back.
		fetched, err := r.client.GetPartnerLink(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read partner link", err.Error())
			return
		}
		current = fetched
	}
	fillPartnerLink(ctx, &plan, current, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *partnerLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state partnerLinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RevokePartnerLink(ctx, state.ID.ValueString()); err != nil {
		if apiErr, ok := err.(*client.APIError); ok && (apiErr.IsNotFound() || apiErr.Status == http.StatusConflict) {
			return
		}
		resp.Diagnostics.AddError("Unable to revoke partner link", err.Error())
	}
}

func (r *partnerLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ModifyPlan plans a new client secret when rotate_secret_trigger changes.
func (r *partnerLinkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	planSecretRotation(ctx, req, resp, "client_secret")
}
