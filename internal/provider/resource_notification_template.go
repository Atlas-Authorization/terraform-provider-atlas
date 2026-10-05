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
	_ resource.Resource                = &notificationTemplateResource{}
	_ resource.ResourceWithImportState = &notificationTemplateResource{}
	_ resource.ResourceWithConfigure   = &notificationTemplateResource{}
)

// NewNotificationTemplateResource is the factory registered with the provider.
func NewNotificationTemplateResource() resource.Resource { return &notificationTemplateResource{} }

type notificationTemplateResource struct {
	client *client.Client
}

type notificationTemplateModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Subject   types.String `tfsdk:"subject"`
	Body      types.String `tfsdk:"body"`
	Category  types.String `tfsdk:"category"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
	UpdatedAt types.Int64  `tfsdk:"updated_at"`
}

func (r *notificationTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_template"
}

func (r *notificationTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *notificationTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A tenant-authored end-user notification template. `POST /v1/notifications` resolves a " +
			"template by `name` ahead of the built-ins, substituting `{{variable}}` placeholders from the send " +
			"request's `data`. The `category` ties the template to the per-user notification preference gate.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Template name, unique per instance; used to reference the template when sending. " +
					"Immutable — changing it forces replacement. Also the import id.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "Notification subject. May contain `{{variable}}` placeholders.",
				Required:            true,
			},
			"body": schema.StringAttribute{
				MarkdownDescription: "Notification body. May contain `{{variable}}` placeholders.",
				Required:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "End-user notification category gating the template against the user's " +
					"preferences: `new_device`, `unattended_access` or `credential_change`.",
				Required: true,
			},
			"created_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time (epoch ms)."},
			"updated_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

func (r *notificationTemplateResource) write(m *notificationTemplateModel) client.NotificationTemplateWrite {
	return client.NotificationTemplateWrite{
		Name:     m.Name.ValueString(),
		Subject:  m.Subject.ValueString(),
		Body:     m.Body.ValueString(),
		Category: m.Category.ValueString(),
	}
}

func (r *notificationTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateNotificationTemplate(ctx, r.write(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create notification template", err.Error())
		return
	}
	mapNotificationTemplate(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetNotificationTemplate(ctx, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read notification template", err.Error())
		return
	}
	mapNotificationTemplate(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateNotificationTemplate(ctx, plan.Name.ValueString(), r.write(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update notification template", err.Error())
		return
	}
	mapNotificationTemplate(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteNotificationTemplate(ctx, state.Name.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete notification template", err.Error())
	}
}

// ImportState imports by template name (the API's key), not the object id.
func (r *notificationTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func mapNotificationTemplate(t *client.NotificationTemplate, m *notificationTemplateModel) {
	m.ID = types.StringValue(t.ID)
	m.Name = types.StringValue(t.Name)
	m.Subject = types.StringValue(t.Subject)
	m.Body = types.StringValue(t.Body)
	m.Category = types.StringValue(t.Category)
	m.CreatedAt = types.Int64Value(t.CreatedAt)
	m.UpdatedAt = types.Int64Value(t.UpdatedAt)
}
