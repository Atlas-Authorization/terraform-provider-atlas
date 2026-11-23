package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &endUserAPIKeyResource{}
	_ resource.ResourceWithImportState = &endUserAPIKeyResource{}
	_ resource.ResourceWithConfigure   = &endUserAPIKeyResource{}
)

// NewEndUserAPIKeyResource is the factory for atlas_end_user_api_key.
func NewEndUserAPIKeyResource() resource.Resource { return &endUserAPIKeyResource{} }

type endUserAPIKeyResource struct {
	client *client.Client
}

type endUserAPIKeyModel struct {
	ID          types.String `tfsdk:"id"`
	SubjectType types.String `tfsdk:"subject_type"`
	SubjectID   types.String `tfsdk:"subject_id"`
	Name        types.String `tfsdk:"name"`
	Scopes      types.Set    `tfsdk:"scopes"`
	Credential  types.String `tfsdk:"credential"`
	Audience    types.String `tfsdk:"audience"`
	PublicKey   types.String `tfsdk:"public_key"`
	KeyID       types.String `tfsdk:"key_id"`
	Prefix      types.String `tfsdk:"prefix"`
	Secret      types.String `tfsdk:"secret"`
	PrivateKey  types.String `tfsdk:"private_key"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
}

func (r *endUserAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_end_user_api_key"
}

func (r *endUserAPIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *endUserAPIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The key id (also the JWT `kid` of a key-pair key).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subject_type": schema.StringAttribute{
				MarkdownDescription: "`user` or `organization`. Changing it replaces the key.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"subject_id": schema.StringAttribute{
				MarkdownDescription: "The id of the user or organization the key belongs to. Changing it replaces the key.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A label for the key.",
				Optional:            true,
			},
			"scopes": schema.SetAttribute{
				MarkdownDescription: "The scopes the key carries. Updates in place.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"credential": schema.StringAttribute{
				MarkdownDescription: "`secret` (default) or `key_pair`. A key-pair key has no shared secret: the holder signs short-lived assertions with a private key. Changing it replaces the key.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("secret"),
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"audience": schema.StringAttribute{
				MarkdownDescription: "Key pair only: the `aud` every assertion must carry (1 to 256 characters). Defaults to the instance id. Changing it replaces the key.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Key pair only: your own public key (SPKI PEM, P-256 or Ed25519). With it Atlas generates nothing and `private_key` stays null, so no private key lands in state: generate the pair with `tls_private_key` and pass `tls_private_key.x.public_key_pem`. Without it Atlas generates a P-256 pair. Changing it replaces the key.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"key_id": schema.StringAttribute{
				MarkdownDescription: "The key id to put in the `kid` header and `iss` claim of assertions.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"prefix": schema.StringAttribute{
				MarkdownDescription: "The non-secret start of a secret key, for recognising it.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "Credential `secret` only: the plaintext key, shown once at creation and kept (sensitive) in state.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "Key pair generated by Atlas only: the PKCS#8 PEM private key, shown once at creation. It is kept (sensitive) in Terraform state, so protect the state or bring your own `public_key` instead.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch ms).",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
	}
	resp.Schema.MarkdownDescription = "An end-user API key for a user or organization. With `credential = \"key_pair\"` the key has no shared secret: Atlas keeps only the public key and the holder signs short-lived assertions with the private key. A generated `private_key` is returned once and lands in Terraform state; bring your own `public_key` (for example from `tls_private_key`) to keep Atlas-side secrets out of state entirely. Destroy revokes the key. To rotate, run `terraform apply -replace=atlas_end_user_api_key.<name>`. Needs the `api_keys:write` scope."
}

// fillEndUserAPIKey maps the API key onto the model. The one-time secret and
// private key are only set when the response carried one (create); otherwise the
// model keeps what state already holds. Values the configuration pinned are kept
// as written so a re-spelled PEM is not reported as drift.
func fillEndUserAPIKey(ctx context.Context, m *endUserAPIKeyModel, k *client.EndUserAPIKey, diags *diag.Diagnostics) {
	m.ID = types.StringValue(k.ID)
	m.SubjectType = types.StringValue(k.OwnerType)
	m.SubjectID = types.StringValue(k.OwnerID)
	m.Name = stringPtrToValue(k.Name)
	m.Scopes = stringSliceToSet(ctx, k.Scopes, diags)
	m.Credential = types.StringValue(k.Credential)
	if m.Audience.IsNull() || m.Audience.IsUnknown() {
		m.Audience = stringPtrToValue(k.Audience)
	}
	if m.PublicKey.IsNull() || m.PublicKey.IsUnknown() {
		m.PublicKey = stringPtrToValue(k.PublicKey)
	}
	if k.KeyID != nil {
		m.KeyID = types.StringValue(*k.KeyID)
	} else {
		m.KeyID = types.StringValue(k.ID)
	}
	m.Prefix = types.StringValue(k.Prefix)
	if k.Secret != nil {
		m.Secret = types.StringValue(*k.Secret)
	} else if m.Secret.IsUnknown() {
		m.Secret = types.StringNull()
	}
	if k.PrivateKey != nil {
		m.PrivateKey = types.StringValue(*k.PrivateKey)
	} else if m.PrivateKey.IsUnknown() {
		m.PrivateKey = types.StringNull()
	}
	m.CreatedAt = types.Int64Value(k.CreatedAt)
}

func (r *endUserAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan endUserAPIKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.EndUserAPIKeyCreate{
		SubjectType: plan.SubjectType.ValueString(),
		SubjectID:   plan.SubjectID.ValueString(),
		Name:        optionalString(plan.Name),
		Credential:  plan.Credential.ValueString(),
		Audience:    optionalString(plan.Audience),
		PublicKey:   optionalString(plan.PublicKey),
	}
	if !plan.Scopes.IsUnknown() {
		body.Scopes = setToStringSlice(ctx, plan.Scopes, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateEndUserAPIKey(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create end-user API key", err.Error())
		return
	}
	fillEndUserAPIKey(ctx, &plan, created, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *endUserAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state endUserAPIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	k, err := r.client.GetEndUserAPIKey(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read end-user API key", err.Error())
		return
	}
	// A revoked key is gone as far as configuration is concerned.
	if k.RevokedAt != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	fillEndUserAPIKey(ctx, &state, k, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *endUserAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state endUserAPIKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	patch := client.EndUserAPIKeyUpdate{}
	if !plan.Name.Equal(state.Name) {
		if plan.Name.IsNull() {
			patch.Name = ptrStr("")
		} else {
			patch.Name = optionalString(plan.Name)
		}
	}
	if !plan.Scopes.IsUnknown() && !plan.Scopes.Equal(state.Scopes) {
		scopes := setToStringSlice(ctx, plan.Scopes, &resp.Diagnostics)
		if scopes == nil {
			scopes = []string{}
		}
		patch.Scopes = &scopes
	}
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateEndUserAPIKey(ctx, state.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update end-user API key", err.Error())
		return
	}
	plan.Secret = state.Secret
	plan.PrivateKey = state.PrivateKey
	fillEndUserAPIKey(ctx, &plan, updated, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *endUserAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state endUserAPIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RevokeEndUserAPIKey(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return // already gone
		}
		resp.Diagnostics.AddError("Unable to revoke end-user API key", err.Error())
	}
}

func (r *endUserAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
