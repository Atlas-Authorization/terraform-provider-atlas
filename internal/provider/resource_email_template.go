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
	_ resource.Resource                = &emailTemplateResource{}
	_ resource.ResourceWithImportState = &emailTemplateResource{}
	_ resource.ResourceWithConfigure   = &emailTemplateResource{}
)

// NewEmailTemplateResource is the factory registered with the provider.
func NewEmailTemplateResource() resource.Resource { return &emailTemplateResource{} }

type emailTemplateResource struct {
	client *client.Client
}

type emailTemplateModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Subject types.String `tfsdk:"subject"`
	Text    types.String `tfsdk:"text"`
}

func (r *emailTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_template"
}

func (r *emailTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *emailTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The copy override for ONE built-in transactional email template (for example " +
			"`verification_code` or `magic_link`). Uses the dedicated email-template endpoint, which validates " +
			"`{{placeholders}}` at save time: a template may only use the variables that template type provides, and " +
			"must keep any required one (a sign-up code email without `{{code}}` is rejected). An invalid template fails " +
			"`apply` with the server's message. Destroying the resource reverts the template to the built-in copy. " +
			"Prefer this over carrying email copy inside `atlas_instance_config`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Equals `name`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The built-in template to override, e.g. `verification_code`, `magic_link`, " +
					"`password_reset`, `organization_invitation`, `new_device_sign_in`. An unknown name is rejected by " +
					"the API. Immutable — changing it forces replacement. Also the import id.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "Override subject line (max 200 characters). May contain `{{variable}}` " +
					"placeholders. Omit to keep the built-in subject.",
				Optional: true,
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "Override plain-text body (max 5000 characters). May contain `{{variable}}` " +
					"placeholders. Omit to keep the built-in body.",
				Optional: true,
			},
		},
	}
}

func (r *emailTemplateResource) put(ctx context.Context, m *emailTemplateModel) error {
	_, err := r.client.PutEmailTemplate(ctx, m.Name.ValueString(), client.EmailTemplateOverride{
		Subject: optionalString(m.Subject),
		Text:    optionalString(m.Text),
	})
	m.ID = types.StringValue(m.Name.ValueString())
	return err
}

func (r *emailTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emailTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.put(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to save email template", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *emailTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state emailTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetEmailTemplate(ctx, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read email template", err.Error())
		return
	}
	// Reverted out of band (no override stored): drop from state so plan recreates it.
	if !got.Customised || got.Override == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.ID = types.StringValue(got.Name)
	state.Subject = stringPtrToValue(got.Override.Subject)
	state.Text = stringPtrToValue(got.Override.Text)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *emailTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan emailTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.put(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to save email template", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete reverts the template to the built-in copy.
func (r *emailTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state emailTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEmailTemplate(ctx, state.Name.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to reset email template", err.Error())
	}
}

func (r *emailTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
