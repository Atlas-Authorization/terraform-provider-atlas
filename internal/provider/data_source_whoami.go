package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &whoamiDataSource{}
	_ datasource.DataSourceWithConfigure = &whoamiDataSource{}
)

// NewWhoamiDataSource is the factory for the atlas_whoami data source.
func NewWhoamiDataSource() datasource.DataSource { return &whoamiDataSource{} }

type whoamiDataSource struct {
	client *client.Client
}

type whoamiModel struct {
	ID                  types.String `tfsdk:"id"`
	KeyType             types.String `tfsdk:"key_type"`
	KeyPrefix           types.String `tfsdk:"key_prefix"`
	Environment         types.String `tfsdk:"environment"`
	Scopes              types.Set    `tfsdk:"scopes"`
	EffectiveScopes     types.Set    `tfsdk:"effective_scopes"`
	Owner               types.String `tfsdk:"owner"`
	AllowedIPs          types.Set    `tfsdk:"allowed_ips"`
	ExpiresAt           types.Int64  `tfsdk:"expires_at"`
	InstanceID          types.String `tfsdk:"instance_id"`
	InstanceEnvironment types.String `tfsdk:"instance_environment"`
	AccountID           types.String `tfsdk:"account_id"`
}

func (d *whoamiDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_whoami"
}

func (d *whoamiDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *whoamiDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "What the provider's own secret key is: its id, scopes (and the scopes presets expand to), " +
			"owner, expiry, IP lock and the instance and account it belongs to (`GET /v1/whoami`). Useful to assert a key " +
			"holds the scopes a configuration needs. Needs no scope.",
		Attributes: map[string]schema.Attribute{
			"id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "The key id."},
			"key_type":             schema.StringAttribute{Computed: true, MarkdownDescription: "`secret`, `partner_link` or `access_token`."},
			"key_prefix":           schema.StringAttribute{Computed: true, MarkdownDescription: "The non-secret start of the key."},
			"environment":          schema.StringAttribute{Computed: true, MarkdownDescription: "`live` or `test`."},
			"scopes":               schema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "The scopes the key was granted (presets as granted)."},
			"effective_scopes":     schema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "The scopes after presets are expanded."},
			"owner":                schema.StringAttribute{Computed: true, MarkdownDescription: "Who is responsible for the key."},
			"allowed_ips":          schema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "The IP lock, empty for none."},
			"expires_at":           schema.Int64Attribute{Computed: true, MarkdownDescription: "When the key expires (epoch ms), if it does."},
			"instance_id":          schema.StringAttribute{Computed: true, MarkdownDescription: "The instance the key belongs to."},
			"instance_environment": schema.StringAttribute{Computed: true, MarkdownDescription: "The instance's environment."},
			"account_id":           schema.StringAttribute{Computed: true, MarkdownDescription: "The account that owns the instance."},
		},
	}
}

func (d *whoamiDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	w, err := d.client.GetWhoami(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read whoami", err.Error())
		return
	}
	m := whoamiModel{
		ID:                  types.StringValue(w.Key.ID),
		KeyType:             types.StringValue(w.Key.Type),
		KeyPrefix:           types.StringValue(w.Key.Prefix),
		Environment:         stringPtrToValue(w.Key.Environment),
		Scopes:              stringSliceToSet(ctx, w.Key.Scopes, &resp.Diagnostics),
		EffectiveScopes:     stringSliceToSet(ctx, w.Key.EffectiveScopes, &resp.Diagnostics),
		Owner:               stringPtrToValue(w.Key.Owner),
		AllowedIPs:          stringSliceToSet(ctx, w.Key.AllowedIPs, &resp.Diagnostics),
		ExpiresAt:           int64PtrToValue(w.Key.ExpiresAt),
		InstanceID:          types.StringValue(w.Instance.ID),
		InstanceEnvironment: stringPtrToValue(w.Instance.Environment),
		AccountID:           stringPtrToValue(w.Account.ID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
