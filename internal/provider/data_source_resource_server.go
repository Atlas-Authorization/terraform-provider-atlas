package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &resourceServerDataSource{}
	_ datasource.DataSourceWithConfigure = &resourceServerDataSource{}
)

// NewResourceServerDataSource is the factory registered with the provider.
func NewResourceServerDataSource() datasource.DataSource { return &resourceServerDataSource{} }

type resourceServerDataSource struct {
	client *client.Client
}

type resourceServerDataModel struct {
	ID              types.String               `tfsdk:"id"`
	Identifier      types.String               `tfsdk:"identifier"`
	Name            types.String               `tfsdk:"name"`
	Scopes          []resourceServerScopeModel `tfsdk:"scopes"`
	TokenTTLSeconds types.Int64                `tfsdk:"token_ttl_seconds"`
	SigningAlg      types.String               `tfsdk:"signing_alg"`
	CreatedAt       types.Int64                `tfsdk:"created_at"`
	UpdatedAt       types.Int64                `tfsdk:"updated_at"`
}

func (d *resourceServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_server"
}

func (d *resourceServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *resourceServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an existing resource server (API) by id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id to look up.",
				Required:            true,
			},
			"identifier": schema.StringAttribute{Computed: true, MarkdownDescription: "The audience identifier."},
			"name":       schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable name."},
			"scopes": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Scopes this API defines.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value":       schema.StringAttribute{Computed: true, MarkdownDescription: "The scope string."},
						"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Optional description."},
					},
				},
			},
			"token_ttl_seconds": schema.Int64Attribute{Computed: true, MarkdownDescription: "Machine-token lifetime (seconds)."},
			"signing_alg":       schema.StringAttribute{Computed: true, MarkdownDescription: "JWT signing algorithm."},
			"created_at":        schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time (epoch ms)."},
			"updated_at":        schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

func (d *resourceServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resourceServerDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rs, err := d.client.GetResourceServer(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read resource server", err.Error())
		return
	}
	config.Identifier = types.StringValue(rs.Identifier)
	config.Name = types.StringValue(rs.Name)
	config.Scopes = scopesFromAPI(rs.Scopes)
	config.TokenTTLSeconds = types.Int64Value(rs.TokenTTLSeconds)
	config.SigningAlg = types.StringValue(rs.SigningAlg)
	config.CreatedAt = types.Int64Value(rs.CreatedAt)
	config.UpdatedAt = types.Int64Value(rs.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
