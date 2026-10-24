package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &atlasIDAppDataSource{}
	_ datasource.DataSourceWithConfigure = &atlasIDAppDataSource{}
)

// NewAtlasIDAppDataSource is the factory registered with the provider.
func NewAtlasIDAppDataSource() datasource.DataSource { return &atlasIDAppDataSource{} }

type atlasIDAppDataSource struct {
	client *client.Client
}

var buttonOriginObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"origin":       types.StringType,
	"last_seen_at": types.Int64Type,
	"requests":     types.Int64Type,
}}

type atlasIDAppDataModel struct {
	ID                    types.String `tfsdk:"id"`
	ClientID              types.String `tfsdk:"client_id"`
	Name                  types.String `tfsdk:"name"`
	Type                  types.String `tfsdk:"type"`
	RedirectURIs          types.List   `tfsdk:"redirect_uris"`
	AllowedScopes         types.List   `tfsdk:"allowed_scopes"`
	HomepageURL           types.String `tfsdk:"homepage_url"`
	SecretPrefix          types.String `tfsdk:"secret_prefix"`
	ManagedByInstanceID   types.String `tfsdk:"managed_by_instance_id"`
	ManagedVia            types.String `tfsdk:"managed_via"`
	ButtonRenderedOrigins types.List   `tfsdk:"button_rendered_origins"`
	CreatedAt             types.Int64  `tfsdk:"created_at"`
	UpdatedAt             types.Int64  `tfsdk:"updated_at"`
}

func (d *atlasIDAppDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_id_app"
}

func (d *atlasIDAppDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *atlasIDAppDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an existing **Sign in with Atlas** app by id (needs the provider's `platform_key`). " +
			"Never returns the client secret.",
		Attributes: map[string]schema.Attribute{
			"id":                     schema.StringAttribute{Required: true, MarkdownDescription: "The app id."},
			"client_id":              schema.StringAttribute{Computed: true, MarkdownDescription: "The OAuth client id."},
			"name":                   schema.StringAttribute{Computed: true, MarkdownDescription: "App name."},
			"type":                   schema.StringAttribute{Computed: true, MarkdownDescription: "`confidential` or `public`."},
			"redirect_uris":          schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Redirect URIs."},
			"allowed_scopes":         schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Allowed scopes."},
			"homepage_url":           schema.StringAttribute{Computed: true, MarkdownDescription: "Homepage URL."},
			"secret_prefix":          schema.StringAttribute{Computed: true, MarkdownDescription: "Non-secret prefix of the live secret."},
			"managed_by_instance_id": schema.StringAttribute{Computed: true, MarkdownDescription: "The instance that manages the app's grants (toggle or adoption)."},
			"managed_via":            schema.StringAttribute{Computed: true, MarkdownDescription: "`instance_toggle`, `adopted`, or unset."},
			"button_rendered_origins": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Page origins the Sign in with Atlas button script ran on in the last 7 days.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"origin":       schema.StringAttribute{Computed: true, MarkdownDescription: "The page origin."},
					"last_seen_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Last seen (epoch ms)."},
					"requests":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Authorize requests seen in the window."},
				}},
			},
			"created_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time (epoch ms)."},
			"updated_at": schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

func (d *atlasIDAppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg atlasIDAppDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := cfg.ID.ValueString()
	app, err := d.client.GetAtlasIDApp(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sign in with Atlas app", err.Error())
		return
	}
	pre, err := d.client.GetAtlasIDAppPreflight(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sign in with Atlas app preflight", err.Error())
		return
	}
	cfg.ClientID = types.StringValue(app.ClientID)
	cfg.Name = types.StringValue(app.Name)
	cfg.Type = types.StringValue(app.Type)
	cfg.RedirectURIs = stringSliceToList(ctx, app.RedirectURIs, &resp.Diagnostics)
	cfg.AllowedScopes = stringSliceToList(ctx, app.AllowedScopes, &resp.Diagnostics)
	cfg.HomepageURL = stringPtrToValue(nonEmpty(app.HomepageURL))
	cfg.SecretPrefix = stringPtrToValue(app.SecretPrefix)
	cfg.ManagedByInstanceID = stringPtrToValue(app.ManagedByInstanceID)
	cfg.ManagedVia = stringPtrToValue(app.ManagedVia)
	cfg.CreatedAt = types.Int64Value(app.CreatedAt)
	cfg.UpdatedAt = types.Int64Value(app.UpdatedAt)
	elems := make([]attr.Value, 0, len(pre.ButtonRenderedOrigins))
	for _, o := range pre.ButtonRenderedOrigins {
		obj, diags := types.ObjectValue(buttonOriginObjectType.AttrTypes, map[string]attr.Value{
			"origin":       types.StringValue(o.Origin),
			"last_seen_at": types.Int64Value(o.LastSeenAt),
			"requests":     types.Int64Value(o.Requests),
		})
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	list, diags := types.ListValue(buttonOriginObjectType, elems)
	resp.Diagnostics.Append(diags...)
	cfg.ButtonRenderedOrigins = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
