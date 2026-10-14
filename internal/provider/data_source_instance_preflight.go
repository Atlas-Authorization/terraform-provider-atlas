package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &instancePreflightDataSource{}
	_ datasource.DataSourceWithConfigure = &instancePreflightDataSource{}
)

// NewInstancePreflightDataSource is the factory for atlas_instance_preflight.
func NewInstancePreflightDataSource() datasource.DataSource { return &instancePreflightDataSource{} }

type instancePreflightDataSource struct {
	client *client.Client
}

type instancePreflightModel struct {
	ID        types.String `tfsdk:"id"`
	OK        types.Bool   `tfsdk:"ok"`
	CheckedAt types.Int64  `tfsdk:"checked_at"`
	Errors    types.Int64  `tfsdk:"error_count"`
	Warnings  types.Int64  `tfsdk:"warning_count"`
	Findings  types.List   `tfsdk:"findings"`
}

var preflightFindingType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"severity": types.StringType,
	"code":     types.StringType,
	"message":  types.StringType,
	"fix":      types.StringType,
	"subject":  types.MapType{ElemType: types.StringType},
}}

func (d *instancePreflightDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_preflight"
}

func (d *instancePreflightDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *instancePreflightDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks the instance setup in one call and lists what is broken or worth knowing " +
			"(provider credentials missing, OAuth clients that cannot complete a sign-in, custom domains not " +
			"serving yet, no email provider for email sign-in, …). Pair it with a `check` block or a " +
			"`precondition` to fail a plan on error-severity findings.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Always `preflight`."},
			"ok":            schema.BoolAttribute{Computed: true, MarkdownDescription: "True when there are no error-severity findings."},
			"checked_at":    schema.Int64Attribute{Computed: true, MarkdownDescription: "When the check ran (epoch ms)."},
			"error_count":   schema.Int64Attribute{Computed: true, MarkdownDescription: "Number of error-severity findings."},
			"warning_count": schema.Int64Attribute{Computed: true, MarkdownDescription: "Number of warning-severity findings."},
			"findings": schema.ListNestedAttribute{
				MarkdownDescription: "Findings, ordered error, then warning, then info.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"severity": schema.StringAttribute{Computed: true, MarkdownDescription: "`error`, `warning` or `info`."},
					"code":     schema.StringAttribute{Computed: true, MarkdownDescription: "Stable snake_case identifier, e.g. `oauth_provider_missing_credentials`."},
					"message":  schema.StringAttribute{Computed: true, MarkdownDescription: "What is wrong, in a plain sentence."},
					"fix":      schema.StringAttribute{Computed: true, MarkdownDescription: "What to do about it."},
					"subject": schema.MapAttribute{Computed: true, ElementType: types.StringType,
						MarkdownDescription: "Ids of what the finding is about (provider, client_id, domain_id, host, …)."},
				}},
			},
		},
	}
}

func (d *instancePreflightDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	p, err := d.client.GetPreflight(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to run instance preflight", err.Error())
		return
	}
	var m instancePreflightModel
	mapPreflight(ctx, p, &m, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func mapPreflight(ctx context.Context, p *client.Preflight, m *instancePreflightModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue("preflight")
	m.OK = types.BoolValue(p.OK)
	m.CheckedAt = types.Int64Value(p.CheckedAt)
	var errs, warns int64
	items := make([]attr.Value, 0, len(p.Findings))
	for _, f := range p.Findings {
		switch f.Severity {
		case "error":
			errs++
		case "warning":
			warns++
		}
		obj, d := types.ObjectValue(preflightFindingType.AttrTypes, map[string]attr.Value{
			"severity": types.StringValue(f.Severity),
			"code":     types.StringValue(f.Code),
			"message":  types.StringValue(f.Message),
			"fix":      types.StringValue(f.Fix),
			"subject":  stringMapToMap(ctx, f.Subject, diags),
		})
		diags.Append(d...)
		items = append(items, obj)
	}
	m.Errors = types.Int64Value(errs)
	m.Warnings = types.Int64Value(warns)
	list, d := types.ListValue(preflightFindingType, items)
	diags.Append(d...)
	m.Findings = list
}
