package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &oauthClientDataSource{}
	_ datasource.DataSourceWithConfigure = &oauthClientDataSource{}
)

// NewOAuthClientDataSource is the factory registered with the provider.
func NewOAuthClientDataSource() datasource.DataSource { return &oauthClientDataSource{} }

type oauthClientDataSource struct {
	client *client.Client
}

// oauthClientDataModel is the read-only projection (no secret is ever returned).
type oauthClientDataModel struct {
	ID                      types.String `tfsdk:"id"`
	ClientID                types.String `tfsdk:"client_id"`
	Name                    types.String `tfsdk:"name"`
	RedirectURIs            types.List   `tfsdk:"redirect_uris"`
	AllowedScopes           types.List   `tfsdk:"allowed_scopes"`
	GrantTypes              types.List   `tfsdk:"grant_types"`
	TokenEndpointAuthMethod types.String `tfsdk:"token_endpoint_auth_method"`
	LogoURI                 types.String `tfsdk:"logo_uri"`
	FirstParty              types.Bool   `tfsdk:"first_party"`
	SecretPrefix            types.String `tfsdk:"secret_prefix"`
	IsPublic                types.Bool   `tfsdk:"is_public"`
	CreatedAt               types.Int64  `tfsdk:"created_at"`
	UpdatedAt               types.Int64  `tfsdk:"updated_at"`
}

func (d *oauthClientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_client"
}

func (d *oauthClientDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *oauthClientDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an existing OAuth client by id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id to look up.",
				Required:            true,
			},
			"client_id":                  schema.StringAttribute{Computed: true, MarkdownDescription: "The public client_id."},
			"name":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Client name."},
			"redirect_uris":              schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Allowed redirect URIs."},
			"allowed_scopes":             schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Allowed scopes."},
			"grant_types":                schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Allowed grant types."},
			"token_endpoint_auth_method": schema.StringAttribute{Computed: true, MarkdownDescription: "Token endpoint auth method."},
			"logo_uri":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Logo URL."},
			"first_party":                schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the client is first-party."},
			"secret_prefix":              schema.StringAttribute{Computed: true, MarkdownDescription: "Non-secret prefix of the live secret."},
			"is_public":                  schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the client is public (PKCE)."},
			"created_at":                 schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time (epoch ms)."},
			"updated_at":                 schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

func (d *oauthClientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config oauthClientDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	c, err := d.client.GetOAuthClient(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read OAuth client", err.Error())
		return
	}
	config.ClientID = types.StringValue(c.ClientID)
	config.Name = types.StringValue(c.Name)
	config.RedirectURIs = stringSliceToList(ctx, c.RedirectURIs, &resp.Diagnostics)
	config.AllowedScopes = stringSliceToList(ctx, c.AllowedScopes, &resp.Diagnostics)
	config.GrantTypes = stringSliceToList(ctx, c.GrantTypes, &resp.Diagnostics)
	config.TokenEndpointAuthMethod = types.StringValue(c.TokenEndpointAuthMethod)
	config.LogoURI = stringPtrToValue(c.LogoURL)
	config.FirstParty = types.BoolValue(c.FirstParty)
	config.SecretPrefix = stringPtrToValue(c.SecretPrefix)
	config.IsPublic = types.BoolValue(c.IsPublic)
	config.CreatedAt = types.Int64Value(c.CreatedAt)
	config.UpdatedAt = types.Int64Value(c.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
