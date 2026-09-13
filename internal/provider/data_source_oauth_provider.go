package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &oauthProviderDataSource{}
	_ datasource.DataSourceWithConfigure = &oauthProviderDataSource{}
)

// NewOAuthProviderDataSource is the factory registered with the provider.
func NewOAuthProviderDataSource() datasource.DataSource { return &oauthProviderDataSource{} }

type oauthProviderDataSource struct {
	client *client.Client
}

// oauthProviderDataModel is the read-only projection (the client secret is never
// returned; only has_secret reports its presence).
type oauthProviderDataModel struct {
	Provider      types.String `tfsdk:"provider_key"`
	DisplayName   types.String `tfsdk:"display_name"`
	Category      types.String `tfsdk:"category"`
	Tier          types.String `tfsdk:"tier"`
	RedirectURI   types.String `tfsdk:"redirect_uri"`
	DefaultScopes types.List   `tfsdk:"default_scopes"`
	Configured    types.Bool   `tfsdk:"configured"`
	ClientID      types.String `tfsdk:"client_id"`
	HasSecret     types.Bool   `tfsdk:"has_secret"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	AllowSignIn   types.Bool   `tfsdk:"allow_sign_in"`
	AllowSignUp   types.Bool   `tfsdk:"allow_sign_up"`
	Scopes        types.List   `tfsdk:"scopes"`
	UpdatedAt     types.Int64  `tfsdk:"updated_at"`
}

func (d *oauthProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_provider"
}

func (d *oauthProviderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *oauthProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a social sign-in provider by its catalog key. The client secret is never returned.",
		Attributes: map[string]schema.Attribute{
			"provider_key": schema.StringAttribute{
				MarkdownDescription: "Catalog key of the provider to look up (e.g. `google`). (Named `provider_key` " +
					"because `provider` is a reserved Terraform meta-argument.)",
				Required: true,
			},
			"display_name":   schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable provider name."},
			"category":       schema.StringAttribute{Computed: true, MarkdownDescription: "Catalog category."},
			"tier":           schema.StringAttribute{Computed: true, MarkdownDescription: "Catalog tier."},
			"redirect_uri":   schema.StringAttribute{Computed: true, MarkdownDescription: "Redirect/callback URI derived from your instance host."},
			"default_scopes": schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "The provider's default scopes."},
			"configured":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether credentials are configured."},
			"client_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "The configured client id (not a secret)."},
			"has_secret":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether a client secret is stored."},
			"enabled":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the provider is enabled."},
			"allow_sign_in":  schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether existing users may sign in with this provider."},
			"allow_sign_up":  schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether new users may sign up with this provider."},
			"scopes":         schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "The configured scopes."},
			"updated_at":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Last update time (epoch ms)."},
		},
	}
}

func (d *oauthProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config oauthProviderDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	c, err := d.client.GetOAuthProvider(ctx, config.Provider.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read OAuth provider", err.Error())
		return
	}
	config.Provider = types.StringValue(c.Provider)
	config.DisplayName = types.StringValue(c.DisplayName)
	config.Category = types.StringValue(c.Category)
	config.Tier = types.StringValue(c.Tier)
	config.RedirectURI = types.StringValue(c.RedirectURI)
	config.DefaultScopes = stringSliceToList(ctx, c.DefaultScopes, &resp.Diagnostics)
	config.Configured = types.BoolValue(c.Configured)
	config.ClientID = stringPtrToValue(c.ClientID)
	config.HasSecret = types.BoolValue(c.HasSecret)
	config.Enabled = types.BoolValue(c.Enabled)
	config.AllowSignIn = types.BoolValue(c.AllowSignIn)
	config.AllowSignUp = types.BoolValue(c.AllowSignUp)
	config.Scopes = stringSliceToList(ctx, c.Scopes, &resp.Diagnostics)
	config.UpdatedAt = int64PtrToValue(c.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
