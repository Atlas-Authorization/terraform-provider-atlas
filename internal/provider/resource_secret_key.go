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
	_ resource.Resource                = &secretKeyResource{}
	_ resource.ResourceWithImportState = &secretKeyResource{}
	_ resource.ResourceWithConfigure   = &secretKeyResource{}
)

// NewSecretKeyResource is the factory for atlas_secret_key.
func NewSecretKeyResource() resource.Resource { return &secretKeyResource{} }

type secretKeyResource struct {
	client *client.Client
}

type secretKeyModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Scopes         types.Set    `tfsdk:"scopes"`
	Owner          types.String `tfsdk:"owner"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	AllowedIPs     types.Set    `tfsdk:"allowed_ips"`
	AllowedOrigins types.Set    `tfsdk:"allowed_origins"`
	ExternalRef    types.String `tfsdk:"external_ref"`
	Metadata       types.Map    `tfsdk:"metadata"`
	Secret         types.String `tfsdk:"secret"`
	Prefix         types.String `tfsdk:"prefix"`
	Environment    types.String `tfsdk:"environment"`
	Status         types.String `tfsdk:"status"`
	CreatedAt      types.Int64  `tfsdk:"created_at"`
	LastUsedAt     types.Int64  `tfsdk:"last_used_at"`
}

func (r *secretKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_key"
}

func (r *secretKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func secretKeySchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "A scoped runtime secret key (`atlas_<env>_sk_...`) for the instance. The plaintext `secret` is " +
			"returned once, on create, and kept (sensitive) in state; Atlas never shows it again. A key expires: when " +
			"`expires_at` is left out it is one year from creation, and it cannot be more than one year away. `scopes` " +
			"and `external_ref` cannot change: changing either replaces the key. Name, owner, expiry, IP and origin " +
			"locks and `metadata` update in place. Destroy REVOKES the key (Atlas keeps the revoked record). To rotate, " +
			"run `terraform apply -replace=atlas_secret_key.<name>` (with `create_before_destroy` for an overlap). " +
			"The key the provider itself uses must not be destroyed by the same apply. Needs the `secret_keys:write` " +
			"scope, and the scopes you grant must be ones the provider's key already holds.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The key id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A label for the key (up to 200 characters).",
				Optional:            true,
			},
			"scopes": schema.SetAttribute{
				MarkdownDescription: "The scopes the key holds, e.g. `[\"users:read\", \"webhooks:write\"]`, or a preset such " +
					"as `read_only`, `privacy:execute` or `privacy:grants:revoke`. Defaults to `*`. Changing it replaces the key.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "Who is responsible for the key: an email, a name or a service id. Defaults to the " +
					"credential that created it.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "When the key stops working, as an RFC 3339 timestamp (e.g. `2027-06-01T00:00:00Z`). " +
					"Defaults to one year from creation; at most one year away; a key that expires cannot be made to never expire.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"allowed_ips": schema.SetAttribute{
				MarkdownDescription: "IP addresses or CIDR ranges the key may be used from. Empty (the default) means any. " +
					"Any other address gets `403 IP_NOT_ALLOWED`.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     emptyStringSet(),
			},
			"allowed_origins": schema.SetAttribute{
				MarkdownDescription: "Browser origins (e.g. `https://app.example.com`) the key may be used from. Empty (the " +
					"default) means any. Any other origin gets `403 ORIGIN_NOT_ALLOWED`.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     emptyStringSet(),
			},
			"external_ref": schema.StringAttribute{
				MarkdownDescription: "Your own stable id for the key (1-128 printable characters, no spaces). A second create " +
					"with the same ref returns the existing key WITHOUT its secret, so a re-run never mints a duplicate. " +
					"Changing it replaces the key.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Tags on the key (lowercase keys; string values), filterable with `?metadata[key]=value`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringMap(),
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "The plaintext key, shown once at creation. Null when the key was adopted by `external_ref`.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"prefix": schema.StringAttribute{
				MarkdownDescription: "The non-secret start of the key, for recognising it.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "`live` or `test`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "`active`, `expiring`, `expired` or `revoked`.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch ms).",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"last_used_at": schema.Int64Attribute{
				MarkdownDescription: "When the key was last used (epoch ms), if ever.",
				Computed:            true,
			},
		},
	}
}

func (r *secretKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = secretKeySchema()
}

// fill maps an API key onto the model. The secret is only set when the response
// carried one (create); otherwise the model keeps the value it had.
func fillSecretKey(ctx context.Context, m *secretKeyModel, k *client.SecretKey, diags *diag.Diagnostics) {
	m.ID = types.StringValue(k.ID)
	m.Name = stringPtrToValue(k.Name)
	m.Scopes = stringSliceToSet(ctx, k.Scopes, diags)
	m.Owner = stringPtrToValue(k.Owner)
	m.ExpiresAt = timeFromAPI(k.ExpiresAt, m.ExpiresAt)
	m.AllowedIPs = stringSliceToSet(ctx, k.AllowedIPs, diags)
	m.AllowedOrigins = stringSliceToSet(ctx, k.AllowedOrigins, diags)
	m.ExternalRef = stringPtrToValue(k.ExternalRef)
	m.Metadata = metadataFromAPI(ctx, k.Metadata, diags)
	if k.Secret != nil {
		m.Secret = types.StringValue(*k.Secret)
	} else if m.Secret.IsUnknown() {
		m.Secret = types.StringNull()
	}
	m.Prefix = types.StringValue(k.Prefix)
	m.Environment = stringPtrToValue(k.Environment)
	m.Status = types.StringValue(k.Status)
	m.CreatedAt = types.Int64Value(k.CreatedAt)
	m.LastUsedAt = int64PtrToValue(k.LastUsedAt)
}

func (r *secretKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secretKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.SecretKeyCreate{
		Name:           optionalString(plan.Name),
		Scopes:         setToStringSlice(ctx, plan.Scopes, &resp.Diagnostics),
		ExpiresAt:      optionalString(plan.ExpiresAt),
		AllowedIPs:     setToStringSlice(ctx, plan.AllowedIPs, &resp.Diagnostics),
		AllowedOrigins: setToStringSlice(ctx, plan.AllowedOrigins, &resp.Diagnostics),
		Owner:          optionalString(plan.Owner),
		ExternalRef:    optionalString(plan.ExternalRef),
		Metadata:       metadataToAPI(ctx, plan.Metadata, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateSecretKey(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create secret key", err.Error())
		return
	}
	fillSecretKey(ctx, &plan, created, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secretKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	k, err := r.client.GetSecretKey(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read secret key", err.Error())
		return
	}
	// A revoked key is gone as far as configuration is concerned.
	if k.RevokedAt != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	fillSecretKey(ctx, &state, k, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *secretKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state secretKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	patch := client.SecretKeyUpdate{Metadata: metadataToAPI(ctx, plan.Metadata, &resp.Diagnostics)}
	if !plan.Name.Equal(state.Name) {
		if plan.Name.IsNull() {
			patch.Name = ptrStr("")
		} else {
			patch.Name = optionalString(plan.Name)
		}
	}
	if !plan.Owner.IsUnknown() && !plan.Owner.Equal(state.Owner) {
		patch.Owner = optionalString(plan.Owner)
	}
	if !plan.ExpiresAt.IsUnknown() && !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.Equal(state.ExpiresAt) {
		patch.ExpiresAt = optionalString(plan.ExpiresAt)
	}
	if !plan.AllowedIPs.Equal(state.AllowedIPs) {
		ips := setToStringSlice(ctx, plan.AllowedIPs, &resp.Diagnostics)
		if ips == nil {
			ips = []string{}
		}
		patch.AllowedIPs = &ips
	}
	if !plan.AllowedOrigins.Equal(state.AllowedOrigins) {
		origins := setToStringSlice(ctx, plan.AllowedOrigins, &resp.Diagnostics)
		if origins == nil {
			origins = []string{}
		}
		patch.AllowedOrigins = &origins
	}
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateSecretKey(ctx, state.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update secret key", err.Error())
		return
	}
	plan.Secret = state.Secret
	fillSecretKey(ctx, &plan, updated, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secretKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RevokeSecretKey(ctx, state.ID.ValueString()); err != nil {
		if apiErr, ok := err.(*client.APIError); ok && (apiErr.IsNotFound() || apiErr.Status == http.StatusConflict) {
			return // already gone or already revoked
		}
		resp.Diagnostics.AddError("Unable to revoke secret key", err.Error())
	}
}

func (r *secretKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
