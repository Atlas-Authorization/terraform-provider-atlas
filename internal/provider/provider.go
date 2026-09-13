package provider

import (
	"context"
	"os"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure AtlasProvider satisfies the framework interface.
var _ provider.Provider = &AtlasProvider{}

// AtlasProvider is the top-level provider. It holds only the build version; the
// configured *client.Client is created in Configure and handed to every
// resource and data source via their Configure hooks.
type AtlasProvider struct {
	version string
}

// AtlasProviderModel maps the provider configuration block.
type AtlasProviderModel struct {
	APIURL    types.String `tfsdk:"api_url"`
	SecretKey types.String `tfsdk:"secret_key"`
}

// New returns a provider factory for the given build version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AtlasProvider{version: version}
	}
}

func (p *AtlasProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "atlas"
	resp.Version = p.version
}

func (p *AtlasProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Atlas instance configuration as code via the secret-key Backend API.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Base URL of the Atlas Backend API. Defaults to `https://api.atlas.dev`, " +
					"or the `ATLAS_API_URL` environment variable when set.",
				Optional: true,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "Atlas instance secret key (`sk_...`). Prefer the `ATLAS_SECRET_KEY` " +
					"environment variable so the key never lands in state or config.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *AtlasProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config AtlasProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Config value wins over the environment; the env var is the fallback.
	apiURL := os.Getenv("ATLAS_API_URL")
	if !config.APIURL.IsNull() && config.APIURL.ValueString() != "" {
		apiURL = config.APIURL.ValueString()
	}
	if apiURL == "" {
		apiURL = client.DefaultAPIURL
	}

	secretKey := os.Getenv("ATLAS_SECRET_KEY")
	if !config.SecretKey.IsNull() && config.SecretKey.ValueString() != "" {
		secretKey = config.SecretKey.ValueString()
	}
	if secretKey == "" {
		resp.Diagnostics.AddAttributeError(
			pathSecretKey(),
			"Missing Atlas secret key",
			"The provider needs an instance secret key (sk_...). Set it in the provider block's "+
				"secret_key attribute or, preferably, the ATLAS_SECRET_KEY environment variable.",
		)
		return
	}

	c := client.New(apiURL, secretKey)
	c.UserAgent = "terraform-provider-atlas/" + p.version

	// Both resources and data sources receive the same configured client.
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *AtlasProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOAuthClientResource,
		NewOAuthProviderResource,
		NewSsoConnectionResource,
		NewResourceServerResource,
		NewJwtTemplateResource,
		NewWebhookEndpointResource,
		NewRoleResource,
		NewOrganizationResource,
		NewDomainResource,
	}
}

func (p *AtlasProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOAuthClientDataSource,
		NewOAuthProviderDataSource,
		NewResourceServerDataSource,
	}
}
