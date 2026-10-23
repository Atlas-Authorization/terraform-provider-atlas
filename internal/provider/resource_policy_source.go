package provider

import (
	"context"
	"net/url"

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
	_ resource.Resource                   = &policySourceResource{}
	_ resource.ResourceWithImportState    = &policySourceResource{}
	_ resource.ResourceWithConfigure      = &policySourceResource{}
	_ resource.ResourceWithValidateConfig = &policySourceResource{}
)

// NewPolicySourceResource is the factory for atlas_policy_source.
func NewPolicySourceResource() resource.Resource { return &policySourceResource{} }

type policySourceResource struct {
	client *client.Client
}

type policySourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	EndpointURL            types.String `tfsdk:"endpoint_url"`
	Sync                   types.String `tfsdk:"sync"`
	PollIntervalMinutes    types.Int64  `tfsdk:"poll_interval_minutes"`
	SigningSecretWO        types.String `tfsdk:"signing_secret_wo"`
	SigningSecretWOVersion types.Int64  `tfsdk:"signing_secret_wo_version"`
	SigningSecretSet       types.Bool   `tfsdk:"signing_secret_set"`
	WebhookURL             types.String `tfsdk:"webhook_url"`
	ProtocolVersion        types.String `tfsdk:"protocol_version"`
	LastSyncedAt           types.Int64  `tfsdk:"last_synced_at"`
	LastSyncOK             types.Bool   `tfsdk:"last_sync_ok"`
	LastSyncError          types.String `tfsdk:"last_sync_error"`
}

const (
	policySourceMinSecretLength = 16
	policySourceMaxSecretLength = 512
)

func (r *policySourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_source"
}

func (r *policySourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *policySourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The external policy source — a singleton per instance. Connect the system that serves " +
			"your policy documents (a CMS, a docs platform, your own service): Atlas fetches " +
			"`GET {endpoint_url}/policies` signed with the signing secret, and in `webhook` mode your system tells " +
			"Atlas about a new version with a signed `policy.published` / `policy.retired` event sent to " +
			"`webhook_url`. Each apply sets every field this resource manages. Destroying it disconnects the source " +
			"(the policies and their history stay). Import with any id, e.g. `policy_source`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `policy_source`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas syncs from the source. Enabling needs `endpoint_url` and a signing " +
					"secret. Default `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"endpoint_url": schema.StringAttribute{
				MarkdownDescription: "The source's base URL, without query or fragment: `https` (`http` only for " +
					"localhost on a development instance). Atlas calls `GET {endpoint_url}/policies`.",
				Required: true,
			},
			"sync": schema.StringAttribute{
				MarkdownDescription: "`webhook` (default): your source sends signed events to `webhook_url` and Atlas " +
					"re-fetches. `poll`: Atlas fetches every `poll_interval_minutes`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("webhook"),
			},
			"poll_interval_minutes": schema.Int64Attribute{
				MarkdownDescription: "Poll mode: minutes between fetches (5-1440). Atlas's default when unset.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"signing_secret_wo": schema.StringAttribute{
				MarkdownDescription: "The signing secret shared with your source (16-512 characters; a `whsec_` secret " +
					"works). **Write-only**: never stored in state or returned by Atlas. It is sent on create and " +
					"whenever `signing_secret_wo_version` changes; change the version to rotate it, or change the " +
					"version with `signing_secret_wo` unset to clear it. Needs Terraform 1.11 or later.",
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"signing_secret_wo_version": schema.Int64Attribute{
				MarkdownDescription: "Any number; changing it sends `signing_secret_wo` again (or clears the secret when " +
					"`signing_secret_wo` is unset).",
				Optional: true,
			},
			"signing_secret_set": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas holds a signing secret (never the secret).",
				Computed:            true,
			},
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "Where your source sends `policy.published` / `policy.retired` events.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"protocol_version": schema.StringAttribute{
				MarkdownDescription: "The policy source protocol version Atlas speaks.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"last_synced_at": schema.Int64Attribute{
				MarkdownDescription: "When the last sync finished (epoch ms); unset before the first.",
				Computed:            true,
			},
			"last_sync_ok": schema.BoolAttribute{
				MarkdownDescription: "Whether the last sync succeeded.",
				Computed:            true,
			},
			"last_sync_error": schema.StringAttribute{
				MarkdownDescription: "Why the last sync failed, if it did.",
				Computed:            true,
			},
		},
	}
}

// ValidateConfig checks the API's rules at plan time.
func (r *policySourceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var m policySourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !m.EndpointURL.IsNull() && !m.EndpointURL.IsUnknown() {
		u, err := url.Parse(m.EndpointURL.ValueString())
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
			resp.Diagnostics.AddAttributeError(path.Root("endpoint_url"), "Invalid endpoint_url",
				"endpoint_url must be an absolute https URL such as https://cms.example.com/atlas, with no query or fragment.")
		}
	}
	if !m.Sync.IsNull() && !m.Sync.IsUnknown() && m.Sync.ValueString() != "webhook" && m.Sync.ValueString() != "poll" {
		resp.Diagnostics.AddAttributeError(path.Root("sync"), "Invalid sync", "sync must be webhook or poll.")
	}
	if !m.PollIntervalMinutes.IsNull() && !m.PollIntervalMinutes.IsUnknown() &&
		(m.PollIntervalMinutes.ValueInt64() < 5 || m.PollIntervalMinutes.ValueInt64() > 1440) {
		resp.Diagnostics.AddAttributeError(path.Root("poll_interval_minutes"), "Invalid poll_interval_minutes",
			"poll_interval_minutes must be between 5 and 1440.")
	}
	if !m.SigningSecretWO.IsNull() && !m.SigningSecretWO.IsUnknown() {
		n := len(m.SigningSecretWO.ValueString())
		if n < policySourceMinSecretLength || n > policySourceMaxSecretLength {
			resp.Diagnostics.AddAttributeError(path.Root("signing_secret_wo"), "Invalid signing_secret_wo",
				"The signing secret must be 16-512 characters.")
		}
	}
}

// apply PUTs the planned settings. prior is the state (nil on create); config
// carries the write-only secret.
func (r *policySourceResource) apply(plan *policySourceModel, config policySourceModel, prior *policySourceModel, ctx context.Context, diags *diag.Diagnostics) {
	body := client.PolicySourceWrite{
		Enabled:             plan.Enabled.ValueBool(),
		EndpointURL:         plan.EndpointURL.ValueString(),
		Sync:                plan.Sync.ValueString(),
		PollIntervalMinutes: optionalInt64(plan.PollIntervalMinutes),
	}
	secret := optionalString(config.SigningSecretWO)
	switch {
	case prior == nil:
		// First apply: send the secret when one is configured.
		if secret != nil {
			body.SendSecret, body.Secret = true, secret
		}
	case !plan.SigningSecretWOVersion.Equal(prior.SigningSecretWOVersion):
		// A new version sends the secret again, or clears it when unset.
		body.SendSecret, body.Secret = true, secret
	}
	out, err := r.client.PutPolicySource(ctx, body)
	if err != nil {
		diags.AddError("Unable to set the policy source", err.Error())
		return
	}
	mapPolicySource(out, plan)
}

func (r *policySourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config policySourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(&plan, config, nil, ctx, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *policySourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policySourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.GetPolicySource(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the policy source", err.Error())
		return
	}
	if !out.Enabled && out.EndpointURL == nil && !out.SigningSecretSet {
		// Never configured, or disconnected outside Terraform.
		resp.State.RemoveResource(ctx)
		return
	}
	mapPolicySource(out, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *policySourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state policySourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(&plan, config, &state, ctx, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *policySourceResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DeletePolicySource(ctx); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to disconnect the policy source", err.Error())
	}
}

func (r *policySourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapPolicySource projects the API view onto the model. The write-only secret
// is always null in state; signing_secret_wo_version is kept from the model.
func mapPolicySource(out *client.PolicySource, m *policySourceModel) {
	m.ID = types.StringValue("policy_source")
	m.Enabled = types.BoolValue(out.Enabled)
	if out.EndpointURL != nil {
		m.EndpointURL = types.StringValue(*out.EndpointURL)
	}
	m.Sync = types.StringValue(out.Sync)
	m.PollIntervalMinutes = types.Int64Value(out.PollIntervalMinutes)
	m.SigningSecretWO = types.StringNull()
	m.SigningSecretSet = types.BoolValue(out.SigningSecretSet)
	m.WebhookURL = stringPtrToValue(out.WebhookURL)
	m.ProtocolVersion = types.StringValue(out.ProtocolVersion)
	m.LastSyncedAt = int64PtrToValue(out.LastSyncedAt)
	m.LastSyncOK = types.BoolNull()
	if out.LastSyncOK != nil {
		m.LastSyncOK = types.BoolValue(*out.LastSyncOK)
	}
	m.LastSyncError = stringPtrToValue(out.LastSyncError)
}
