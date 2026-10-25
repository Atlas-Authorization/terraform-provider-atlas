package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &atlasIDProvisionedAccountResource{}
	_ resource.ResourceWithImportState = &atlasIDProvisionedAccountResource{}
	_ resource.ResourceWithConfigure   = &atlasIDProvisionedAccountResource{}
)

// minProvisionedPasswordLength mirrors the API's minimum.
const minProvisionedPasswordLength = 12

// NewAtlasIDProvisionedAccountResource is the factory for atlas_id_provisioned_account.
func NewAtlasIDProvisionedAccountResource() resource.Resource {
	return &atlasIDProvisionedAccountResource{}
}

type atlasIDProvisionedAccountResource struct {
	client *client.Client
}

type atlasIDProvisionedAccountModel struct {
	ID                types.String `tfsdk:"id"`
	AppID             types.String `tfsdk:"app_id"`
	Email             types.String `tfsdk:"email"`
	Label             types.String `tfsdk:"label"`
	EmailVerified     types.Bool   `tfsdk:"email_verified"`
	TwoFactorExempt   types.Bool   `tfsdk:"two_factor_exempt"`
	ExpiresAt         types.String `tfsdk:"expires_at"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64  `tfsdk:"password_wo_version"`
	GeneratedPassword types.String `tfsdk:"generated_password"`
	Sub               types.String `tfsdk:"sub"`
	Kind              types.String `tfsdk:"kind"`
	Expired           types.Bool   `tfsdk:"expired"`
	CreatedAt         types.Int64  `tfsdk:"created_at"`
}

func (r *atlasIDProvisionedAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_id_provisioned_account"
}

func (r *atlasIDProvisionedAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *atlasIDProvisionedAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A **test or review account** for a Sign in with Atlas app: a real Atlas account (it signs in on " +
			"`atlasauth.net` with its email and password) that is **confined to this one app**. Every other app, the Atlas " +
			"dashboard and Atlas account settings refuse it. Use it for directory and app-store reviewer logins, end-to-end " +
			"tests in CI, demos and partner testing. Needs the provider's `platform_key` with the scope " +
			"`platform:atlas_id_apps:provision`, or `atlas_id_apps:<app id>:provision` for one app. The password is " +
			"**write-only** (`password_wo`, Terraform 1.11 or later): it is sent to Atlas and never kept in state. Leave it " +
			"out and Atlas generates one, which lands in the sensitive `generated_password`. Everything but the password can " +
			"only be changed by replacing the account; destroying it removes the account and ends its sessions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The test account id.",
				Computed:            true,
				PlanModifiers:       keep,
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The Sign in with Atlas app (`atlas_id_app.id`) the account belongs to. Changing it replaces the account.",
				Required:            true,
				PlanModifiers:       replace,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The account's email address. It must not belong to an existing Atlas account and, when the app " +
					"sets `provisioned_verified_domains_only`, must be on a domain you verified. Changing it replaces the account.",
				Required:      true,
				PlanModifiers: replace,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "What the account is for, for example `app-store reviewer` or `ci-e2e` (at most 100 characters). " +
					"Shown on the account's badge and in the audit log. Changing it replaces the account.",
				Required:      true,
				PlanModifiers: replace,
			},
			"email_verified": schema.BoolAttribute{
				MarkdownDescription: "Whether the address counts as verified. Default `true`: reviewer and CI accounts never read the inbox.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace(), boolplanmodifier.UseStateForUnknown()},
			},
			"two_factor_exempt": schema.BoolAttribute{
				MarkdownDescription: "**A risk you opt into.** The account is never asked for a second factor when it signs in to this " +
					"app, for review teams that cannot receive codes. Anyone who holds the password can then sign in to the app " +
					"as this account. Default `false`.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace(), boolplanmodifier.UseStateForUnknown()},
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "When Atlas removes the account by itself, as an RFC 3339 time. Omit it for a non-expiring account " +
					"(a standing reviewer login). Changing it replaces the account.",
				Optional:      true,
				PlanModifiers: replace,
			},
			"password_wo": schema.StringAttribute{
				MarkdownDescription: "The account password (at least 12 characters). **Write-only**: sent to Atlas on create and whenever " +
					"`password_wo_version` changes, and never stored in state. Omit it to let Atlas generate one. Needs Terraform 1.11 or later.",
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"password_wo_version": schema.Int64Attribute{
				MarkdownDescription: "Any number; changing it rotates the password in place (sending `password_wo`, or asking Atlas for a new generated one).",
				Optional:            true,
			},
			"generated_password": schema.StringAttribute{
				MarkdownDescription: "The password Atlas generated, when `password_wo` is not set. Returned by Atlas once, so this is the only " +
					"place to read it: treat the state as sensitive. Null when you supply `password_wo`.",
				Computed:  true,
				Sensitive: true,
			},
			"sub": schema.StringAttribute{
				MarkdownDescription: "The Atlas ID subject the app sees for this account.",
				Computed:            true,
				PlanModifiers:       keep,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "`provisioned` for an account created here.",
				Computed:            true,
				PlanModifiers:       keep,
			},
			"expired": schema.BoolAttribute{
				MarkdownDescription: "Whether `expires_at` has passed.",
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

// fill projects the API object onto the model. The password-related attributes
// are left to the caller.
func fillProvisionedAccount(m *atlasIDProvisionedAccountModel, a *client.AtlasIDProvisionedAccount) {
	m.ID = types.StringValue(a.ID)
	if a.Email != nil {
		m.Email = types.StringValue(*a.Email)
	}
	m.Label = types.StringValue(a.Label)
	m.EmailVerified = types.BoolValue(a.EmailVerified)
	m.TwoFactorExempt = types.BoolValue(a.TwoFactorExempt)
	m.ExpiresAt = timeFromAPI(a.ExpiresAt, m.ExpiresAt)
	m.Sub = types.StringValue(a.Sub)
	m.Kind = types.StringValue(a.Kind)
	m.Expired = types.BoolValue(a.Expired)
	m.CreatedAt = types.Int64Value(a.CreatedAt)
}

func suppliedPassword(cfg *atlasIDProvisionedAccountModel, diags *diag.Diagnostics) (string, bool) {
	if cfg.PasswordWO.IsNull() || cfg.PasswordWO.IsUnknown() {
		return "", false
	}
	pw := cfg.PasswordWO.ValueString()
	if len(pw) < minProvisionedPasswordLength {
		diags.AddAttributeError(path.Root("password_wo"), "Password too short",
			fmt.Sprintf("password_wo must be at least %d characters; omit it to let Atlas generate one.", minProvisionedPasswordLength))
		return "", false
	}
	return pw, true
}

func generatedPasswordValue(a *client.AtlasIDProvisionedAccount, supplied bool) types.String {
	if supplied || a.Password == nil {
		return types.StringNull()
	}
	return types.StringValue(*a.Password)
}

func (r *atlasIDProvisionedAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, cfg atlasIDProvisionedAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]interface{}{
		"email": plan.Email.ValueString(),
		"label": plan.Label.ValueString(),
	}
	if !plan.EmailVerified.IsNull() && !plan.EmailVerified.IsUnknown() {
		body["email_verified"] = plan.EmailVerified.ValueBool()
	}
	if !plan.TwoFactorExempt.IsNull() && !plan.TwoFactorExempt.IsUnknown() {
		body["two_factor_exempt"] = plan.TwoFactorExempt.ValueBool()
	}
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		t, err := time.Parse(time.RFC3339, plan.ExpiresAt.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("expires_at"), "Invalid expires_at", "expires_at must be an RFC 3339 time such as 2027-01-31T00:00:00Z.")
			return
		}
		body["expires_at"] = t.UnixMilli()
	}
	pw, supplied := suppliedPassword(&cfg, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if supplied {
		body["password"] = pw
	}
	created, err := r.client.CreateAtlasIDProvisionedAccount(ctx, plan.AppID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create the Sign in with Atlas test account", err.Error())
		return
	}
	fillProvisionedAccount(&plan, created)
	plan.PasswordWO = types.StringNull()
	plan.GeneratedPassword = generatedPasswordValue(created, supplied)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *atlasIDProvisionedAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state atlasIDProvisionedAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetAtlasIDProvisionedAccount(ctx, state.AppID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read the Sign in with Atlas test account", err.Error())
		return
	}
	fillProvisionedAccount(&state, got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update only ever rotates the password: every other attribute replaces.
func (r *atlasIDProvisionedAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, cfg, state atlasIDProvisionedAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.PasswordWO = types.StringNull()
	if plan.PasswordWOVersion.Equal(state.PasswordWOVersion) {
		plan.GeneratedPassword = state.GeneratedPassword
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}
	body := map[string]interface{}{}
	pw, supplied := suppliedPassword(&cfg, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if supplied {
		body["password"] = pw
	}
	rotated, err := r.client.RotateAtlasIDProvisionedAccountPassword(ctx, state.AppID.ValueString(), state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to rotate the Sign in with Atlas test account password", err.Error())
		return
	}
	fillProvisionedAccount(&plan, rotated)
	plan.GeneratedPassword = generatedPasswordValue(rotated, supplied)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *atlasIDProvisionedAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state atlasIDProvisionedAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAtlasIDProvisionedAccount(ctx, state.AppID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove the Sign in with Atlas test account", err.Error())
	}
}

// ImportState imports by "<app id>/<account id>". The password cannot be recovered.
func (r *atlasIDProvisionedAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	app, id, ok := strings.Cut(req.ID, "/")
	if !ok || app == "" || id == "" {
		resp.Diagnostics.AddError("Invalid import id", "Use <app id>/<account id>, for example oac_123/iacct_456.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), app)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
