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
	_ resource.Resource                = &notificationCategoryResource{}
	_ resource.ResourceWithImportState = &notificationCategoryResource{}
	_ resource.ResourceWithConfigure   = &notificationCategoryResource{}
)

// NewNotificationCategoryResource is the factory registered with the provider.
func NewNotificationCategoryResource() resource.Resource { return &notificationCategoryResource{} }

type notificationCategoryResource struct {
	client *client.Client
}

type notificationCategoryModel struct {
	ID        types.String `tfsdk:"id"`
	Key       types.String `tfsdk:"key"`
	Label     types.String `tfsdk:"label"`
	Optional  types.Bool   `tfsdk:"optional"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func (r *notificationCategoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_category"
}

func (r *notificationCategoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *notificationCategoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A tenant-defined end-user notification category. The `key` identifies the category; the " +
			"`label` is shown on the end-user's notification-preference toggle. A tenant category is always optional " +
			"(it cannot shadow a built-in category) and ties a `atlas_notification_template.category` to the per-user " +
			"preference gate. Deleting a category still referenced by a template is rejected (`409`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Category key: lowercase letters, digits and underscores, starting with a letter " +
					"(max 64). Unique per instance and immutable — changing it forces replacement. Also the import id.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "User-facing label (1–120 characters) shown on the preference toggle.",
				Required:            true,
			},
			"optional": schema.BoolAttribute{
				MarkdownDescription: "Always `true`: a tenant category is opt-out-able and never gates a built-in.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time (epoch ms)."},
		},
	}
}

func (r *notificationCategoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationCategoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateNotificationCategory(ctx, client.NotificationCategoryWrite{
		Key:   plan.Key.ValueString(),
		Label: plan.Label.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create notification category", err.Error())
		return
	}
	mapNotificationCategory(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationCategoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationCategoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetNotificationCategory(ctx, state.Key.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read notification category", err.Error())
		return
	}
	mapNotificationCategory(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationCategoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationCategoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateNotificationCategory(ctx, plan.Key.ValueString(), client.NotificationCategoryWrite{
		Label: plan.Label.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update notification category", err.Error())
		return
	}
	mapNotificationCategory(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationCategoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationCategoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteNotificationCategory(ctx, state.Key.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete notification category", err.Error())
	}
}

// ImportState imports by category key (the API's key), not the object id.
func (r *notificationCategoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

func mapNotificationCategory(c *client.NotificationCategory, m *notificationCategoryModel) {
	m.ID = types.StringValue(c.ID)
	m.Key = types.StringValue(c.Key)
	m.Label = types.StringValue(c.Label)
	m.Optional = types.BoolValue(c.Optional)
	m.CreatedAt = types.Int64Value(c.CreatedAt)
}
