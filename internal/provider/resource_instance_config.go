package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &instanceConfigResource{}
	_ resource.ResourceWithImportState = &instanceConfigResource{}
	_ resource.ResourceWithConfigure   = &instanceConfigResource{}
)

// NewInstanceConfigResource is the factory registered with the provider.
func NewInstanceConfigResource() resource.Resource { return &instanceConfigResource{} }

type instanceConfigResource struct {
	client *client.Client
}

type instanceConfigModel struct {
	ID                 types.String `tfsdk:"id"`
	AllowedOrigins     types.Set    `tfsdk:"allowed_origins"`
	AuthConfig         types.String `tfsdk:"auth_config"`
	AuthConfigResolved types.String `tfsdk:"auth_config_resolved"`
	Environment        types.String `tfsdk:"environment"`
	PublishableKey     types.String `tfsdk:"publishable_key"`
	FrontendAPIHost    types.String `tfsdk:"frontend_api_host"`
	CreatedAt          types.Int64  `tfsdk:"created_at"`
}

func (r *instanceConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_config"
}

func (r *instanceConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *instanceConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The singleton configuration of the Atlas instance this secret key belongs to: its " +
			"`allowed_origins` CORS list and its `auth_config`. There is exactly one per instance — manage a single " +
			"`atlas_instance_config` resource. `auth_config` is a PARTIAL patch merged server-side (never a wholesale " +
			"replacement), so manage the whole object you intend to set and expect `auth_config` to read back the merged " +
			"result. Destroying this resource only stops Terraform managing the config; it does not reset the instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The Atlas instance id (singleton key).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"allowed_origins": schema.SetAttribute{
				MarkdownDescription: "Exact origins (`scheme://host[:port]`, no wildcards, no path) permitted for the " +
					"widget / JS SDK and OAuth redirect validation. Native-app webview origins (`tauri://`, " +
					"`capacitor://`, `ionic://`) are accepted.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"auth_config": schema.StringAttribute{
				MarkdownDescription: "The instance auth configuration as a JSON object string (use `jsonencode(...)`). " +
					"A PARTIAL patch merged into the current config server-side; reads back the server-merged result. " +
					"The provider suppresses the diff when your config is a subset of that merged value, so a partial " +
					"patch does not thrash the plan.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{authConfigSubsetModifier{}},
			},
			"auth_config_resolved": schema.StringAttribute{
				MarkdownDescription: "The full effective configuration Atlas resolved from your `auth_config` plus " +
					"defaults (read-only).",
				Computed: true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The instance environment (e.g. `production`, `development`).",
				Computed:            true,
			},
			"publishable_key": schema.StringAttribute{
				MarkdownDescription: "The instance publishable key (`pk_...`).",
				Computed:            true,
			},
			// Read-only/informational: auto-assigned by Atlas and populated by
			// Read. It is not a diff source — never set from config.
			"frontend_api_host": schema.StringAttribute{
				MarkdownDescription: "The instance Frontend API host (read-only; auto-assigned by Atlas).",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Instance creation time (epoch ms).",
				Computed:            true,
			},
		},
	}
}

func (r *instanceConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan instanceConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *instanceConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetInstance(ctx)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read instance config", err.Error())
		return
	}
	r.mapToState(ctx, fetched, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *instanceConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan instanceConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op: an instance's config is a singleton that cannot be deleted.
// Removing the resource simply stops Terraform managing it; the live config is
// left exactly as it is.
func (r *instanceConfigResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *instanceConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The id is the instance id, but Read always fetches the singleton, so any
	// placeholder works; passthrough keeps the convention of the other resources.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// write sends the plan as a PATCH, re-reads the full projection, and maps it back.
func (r *instanceConfigResource) write(ctx context.Context, plan *instanceConfigModel, diags *diag.Diagnostics) {
	body := client.InstanceUpdate{}
	if !plan.AllowedOrigins.IsNull() && !plan.AllowedOrigins.IsUnknown() {
		origins := setToStringSlice(ctx, plan.AllowedOrigins, diags)
		if origins == nil {
			origins = []string{}
		}
		body.AllowedOrigins = origins
	}
	raw, err := stringToJSONRaw(plan.AuthConfig)
	if err != nil {
		diags.AddError("Invalid auth_config", "auth_config must be a JSON object string: "+err.Error())
		return
	}
	body.AuthConfig = raw

	if _, err := r.client.UpdateInstance(ctx, body); err != nil {
		diags.AddError("Unable to update instance config", err.Error())
		return
	}
	// The PATCH response omits the read-only fields (environment, keys, …); re-read
	// the full projection so every computed attribute is populated.
	fetched, err := r.client.GetInstance(ctx)
	if err != nil {
		diags.AddError("Unable to read instance config", err.Error())
		return
	}
	r.mapToState(ctx, fetched, plan, diags)
}

func (r *instanceConfigResource) mapToState(ctx context.Context, in *client.Instance, m *instanceConfigModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(in.ID)
	m.AllowedOrigins = stringSliceToSet(ctx, in.AllowedOrigins, diags)
	m.AuthConfig = jsonRawToValue(in.AuthConfig)
	m.AuthConfigResolved = jsonRawToValue(in.AuthConfigResolved)
	m.Environment = types.StringValue(in.Environment)
	m.PublishableKey = types.StringValue(in.PublishableKey)
	m.FrontendAPIHost = types.StringValue(in.FrontendAPIHost)
	m.CreatedAt = types.Int64Value(in.CreatedAt)
}
