package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &domainResource{}
	_ resource.ResourceWithImportState = &domainResource{}
	_ resource.ResourceWithConfigure   = &domainResource{}
)

// NewDomainResource is the factory registered with the provider.
func NewDomainResource() resource.Resource { return &domainResource{} }

type domainResource struct {
	client *client.Client
}

type domainModel struct {
	ID                 types.String `tfsdk:"id"`
	Role               types.String `tfsdk:"role"`
	Host               types.String `tfsdk:"host"`
	Status             types.String `tfsdk:"status"`
	CnameTarget        types.String `tfsdk:"cname_target"`
	Live               types.Bool   `tfsdk:"live"`
	CookieDomain       types.String `tfsdk:"cookie_domain"`
	FailureReason      types.String `tfsdk:"failure_reason"`
	CertificateExpires types.Int64  `tfsdk:"certificate_expires_at"`
	Method             types.String `tfsdk:"method"`
	BasePath           types.String `tfsdk:"base_path"`
	Routing            types.String `tfsdk:"routing"`
	FrontendAPIURL     types.String `tfsdk:"frontend_api_url"`
	VerificationURL    types.String `tfsdk:"verification_url"`
	VerificationToken  types.String `tfsdk:"verification_token"`
	VerificationBody   types.String `tfsdk:"verification_content"`
}

func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A custom instance domain (the frontend-api `fapi` host or the `accounts` host). Point a " +
			"CNAME at the returned cname_target (or, with `method = \"path\"`, serve `verification_content` at " +
			"`verification_url`), then verify out of band. The Backend API has no update route, so a " +
			"change to role or host replaces the domain.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Domain role: fapi or accounts. Immutable — forces replacement.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "The hostname to serve (e.g. auth.example.com). Immutable — forces replacement.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Provisioning status (pending, active, ...).",
				Computed:            true,
			},
			"cname_target": schema.StringAttribute{
				MarkdownDescription: "The CNAME target to point the host at.",
				Computed:            true,
			},
			"live": schema.BoolAttribute{
				MarkdownDescription: "Whether the domain is fully live (DNS + certificate + cookie checks pass).",
				Computed:            true,
			},
			"cookie_domain": schema.StringAttribute{
				MarkdownDescription: "The registrable cookie domain derived from the host, if resolvable.",
				Computed:            true,
			},
			"failure_reason": schema.StringAttribute{
				MarkdownDescription: "Most recent verification failure reason, if any.",
				Computed:            true,
			},
			"certificate_expires_at": schema.Int64Attribute{
				MarkdownDescription: "Certificate expiry (epoch ms), once issued.",
				Computed:            true,
			},
			"method": schema.StringAttribute{
				MarkdownDescription: "How the host is proven: `cname` (default; a CNAME to Atlas) or `path` (the Frontend API " +
					"served under `base_path` on a host you already run, proven by an https file, no DNS). Immutable — forces replacement.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"base_path": schema.StringAttribute{
				MarkdownDescription: "`method = \"path\"` only: the prefix the Frontend API is served under (default `/__atlas`). " +
					"Immutable — forces replacement.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"routing": schema.StringAttribute{
				MarkdownDescription: "Path mount, once verified: `atlas_ingress` (Atlas routes the prefix) or `customer_proxy` " +
					"(add the reverse-proxy rule yourself).",
				Computed: true,
			},
			"frontend_api_url": schema.StringAttribute{
				MarkdownDescription: "Path mount: the Frontend API URL (and, once verified, the issuer), `https://<host><base_path>`.",
				Computed:            true,
			},
			"verification_url": schema.StringAttribute{
				MarkdownDescription: "Path mount: the https URL to serve the verification file at.",
				Computed:            true,
			},
			"verification_token": schema.StringAttribute{
				MarkdownDescription: "Path mount: the verification token, bound to this instance, host and base path.",
				Computed:            true,
			},
			"verification_content": schema.StringAttribute{
				MarkdownDescription: "Path mount: the exact body to serve at `verification_url`.",
				Computed:            true,
			},
		},
	}
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateDomain(ctx, client.DomainCreate{
		Role:     plan.Role.ValueString(),
		Host:     plan.Host.ValueString(),
		Method:   plan.Method.ValueString(),
		BasePath: plan.BasePath.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create domain", err.Error())
		return
	}
	r.mapToState(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetDomain(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read domain", err.Error())
		return
	}
	r.mapToState(fetched, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is unreachable: role and host are RequiresReplace. It satisfies the
// resource.Resource interface.
func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteDomain(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete domain", err.Error())
	}
}

func (r *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *domainResource) mapToState(d *client.Domain, m *domainModel) {
	m.ID = types.StringValue(d.ID)
	m.Role = types.StringValue(d.Role)
	m.Host = types.StringValue(d.Host)
	m.Status = types.StringValue(d.Status)
	m.CnameTarget = stringPtrToValue(d.CnameTarget)
	m.Live = types.BoolValue(d.Live)
	m.CookieDomain = stringPtrToValue(d.CookieDomain)
	m.FailureReason = stringPtrToValue(d.FailureReason)
	m.CertificateExpires = int64PtrToValue(d.CertificateExpires)
	m.Method = types.StringValue(d.Method)
	if d.Method == "" {
		m.Method = types.StringValue("cname")
	}
	m.BasePath = stringPtrToValue(d.BasePath)
	m.Routing = stringPtrToValue(d.Routing)
	m.FrontendAPIURL = stringPtrToValue(d.FrontendAPIURL)
	m.VerificationURL, m.VerificationToken, m.VerificationBody = types.StringNull(), types.StringNull(), types.StringNull()
	if v := d.Verification; v != nil {
		m.VerificationURL = types.StringValue(v.URL)
		m.VerificationToken = types.StringValue(v.Token)
		m.VerificationBody = types.StringValue(v.Content)
	}
}
