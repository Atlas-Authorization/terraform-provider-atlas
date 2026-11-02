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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &beforeUserCreateHookResource{}
	_ resource.ResourceWithImportState = &beforeUserCreateHookResource{}
	_ resource.ResourceWithConfigure   = &beforeUserCreateHookResource{}
)

// NewBeforeUserCreateHookResource is the factory for atlas_before_user_create_hook.
func NewBeforeUserCreateHookResource() resource.Resource { return &beforeUserCreateHookResource{} }

type beforeUserCreateHookResource struct {
	client *client.Client
}

type beforeUserCreateHookModel struct {
	ID                     types.String `tfsdk:"id"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	URL                    types.String `tfsdk:"url"`
	SigningSecretWO        types.String `tfsdk:"signing_secret_wo"`
	SigningSecretWOVersion types.Int64  `tfsdk:"signing_secret_wo_version"`
	SigningSecretSet       types.Bool   `tfsdk:"signing_secret_set"`
	TimeoutMs              types.Int64  `tfsdk:"timeout_ms"`
	FailClosed             types.Bool   `tfsdk:"fail_closed"`
}

const beforeUserCreateHookID = "before_user_create_hook"

func (r *beforeUserCreateHookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_before_user_create_hook"
}

func (r *beforeUserCreateHookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *beforeUserCreateHookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `before_user_create` hook — a singleton per instance. Before Atlas creates a user it calls " +
			"your `url` (signed with the signing secret); your answer can set `external_id` and metadata, or deny the sign-up. " +
			"Destroying the resource switches the hook off and clears its URL and secret. Import with any id, e.g. " +
			"`before_user_create_hook`. Needs the `instance:read` and `instance:write` scopes.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `before_user_create_hook`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Call the hook on sign-up. Needs `url` and a signing secret. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The HTTPS endpoint Atlas calls (HTTP to localhost is allowed on a development instance). " +
					"Unset clears it.",
				Optional: true,
			},
			"signing_secret_wo": schema.StringAttribute{
				MarkdownDescription: "The secret Atlas signs each call with, 16 to 512 characters. **Write-only**: never stored " +
					"in state or returned by Atlas. Sent on create and whenever `signing_secret_wo_version` changes (unset clears it).",
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"signing_secret_wo_version": schema.Int64Attribute{
				MarkdownDescription: "Any number; changing it sends `signing_secret_wo` again (or clears the secret when it is unset).",
				Optional:            true,
			},
			"signing_secret_set": schema.BoolAttribute{
				MarkdownDescription: "Whether Atlas holds a signing secret for the hook.",
				Computed:            true,
			},
			"timeout_ms": schema.Int64Attribute{
				MarkdownDescription: "How long Atlas waits for your answer, 500 to 5000 ms. Default 3000.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"fail_closed": schema.BoolAttribute{
				MarkdownDescription: "Refuse the sign-up (`403 SIGN_UP_DENIED`) when the hook times out, fails or answers " +
					"something Atlas cannot read, instead of creating the user without your fields. While your endpoint is " +
					"down nobody can sign up. Default `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
	}
}

func fillBeforeUserCreateHook(m *beforeUserCreateHookModel, out *client.BeforeUserCreateHook) {
	m.ID = types.StringValue(beforeUserCreateHookID)
	m.Enabled = types.BoolValue(out.Enabled)
	m.URL = types.StringPointerValue(out.URL)
	m.SigningSecretSet = types.BoolValue(out.SigningSecretSet)
	m.TimeoutMs = types.Int64Value(out.TimeoutMs)
	m.FailClosed = types.BoolValue(out.FailClosed)
	m.SigningSecretWO = types.StringNull()
}

// apply sends the whole configuration in one PATCH; prior is nil on create.
func (r *beforeUserCreateHookResource) apply(ctx context.Context, plan, config, prior *beforeUserCreateHookModel, diags *diag.Diagnostics) {
	body := map[string]interface{}{
		"enabled":     plan.Enabled.ValueBool(),
		"fail_closed": plan.FailClosed.ValueBool(),
		"url":         nil,
	}
	if !plan.URL.IsNull() {
		body["url"] = plan.URL.ValueString()
	}
	if !plan.TimeoutMs.IsNull() && !plan.TimeoutMs.IsUnknown() {
		body["timeout_ms"] = plan.TimeoutMs.ValueInt64()
	}
	switch {
	case prior == nil:
		if !config.SigningSecretWO.IsNull() {
			body["signing_secret"] = config.SigningSecretWO.ValueString()
		}
	case !plan.SigningSecretWOVersion.Equal(prior.SigningSecretWOVersion):
		if config.SigningSecretWO.IsNull() {
			body["signing_secret"] = nil
		} else {
			body["signing_secret"] = config.SigningSecretWO.ValueString()
		}
	}
	out, err := r.client.PatchBeforeUserCreateHook(ctx, body)
	if err != nil {
		diags.AddError("Unable to set the before_user_create hook", err.Error())
		return
	}
	fillBeforeUserCreateHook(plan, out)
}

func (r *beforeUserCreateHookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config beforeUserCreateHookModel
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

func (r *beforeUserCreateHookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state beforeUserCreateHookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.GetBeforeUserCreateHook(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the before_user_create hook", err.Error())
		return
	}
	fillBeforeUserCreateHook(&state, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *beforeUserCreateHookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state beforeUserCreateHookModel
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

// Delete switches the hook off and clears its URL and secret.
func (r *beforeUserCreateHookResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	_, err := r.client.PatchBeforeUserCreateHook(ctx, map[string]interface{}{
		"enabled": false, "fail_closed": false, "url": nil, "signing_secret": nil,
	})
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to switch off the before_user_create hook", err.Error())
	}
}

func (r *beforeUserCreateHookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), beforeUserCreateHookID)...)
}
