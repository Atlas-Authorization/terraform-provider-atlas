package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Generated data sources: the read-only twin of a table-driven resource (or a
// data-only spec such as atlas_user). The object is selected by its id and path
// keys; every other attribute is computed. Write-only secrets are not exposed.

var (
	_ datasource.DataSource              = &genDataSource{}
	_ datasource.DataSourceWithConfigure = &genDataSource{}
)

type genDataSource struct {
	spec   *genSpec
	client *client.Client
}

func genDataSourceFactories() []func() datasource.DataSource {
	specs := genResourceSpecs()
	var out []func() datasource.DataSource
	for i := range specs {
		if !specs[i].DS && specs[i].Kind != "dataonly" {
			continue
		}
		s := &specs[i]
		out = append(out, func() datasource.DataSource { return &genDataSource{spec: s} })
	}
	return out
}

func (d *genDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.spec.Type
}

func (d *genDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *genDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	s := d.spec
	idDesc := "Atlas object id."
	if s.Kind == "coll" || s.Kind == "dataonly" {
		idDesc = "Atlas object id to look up."
	}
	attrs := map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required:            s.Kind == "coll" || s.Kind == "dataonly",
			Computed:            s.Kind != "coll" && s.Kind != "dataonly",
			MarkdownDescription: idDesc,
		},
	}
	for _, a := range s.Attrs {
		if a.WriteOnly {
			continue
		}
		req := a.Param
		comp := !a.Param
		switch a.Kind {
		case "bool":
			attrs[a.Name] = dschema.BoolAttribute{Computed: comp, Sensitive: a.Sens, MarkdownDescription: a.Desc}
		case "int":
			attrs[a.Name] = dschema.Int64Attribute{Computed: comp, Sensitive: a.Sens, MarkdownDescription: a.Desc}
		case "float":
			attrs[a.Name] = dschema.Float64Attribute{Computed: comp, Sensitive: a.Sens, MarkdownDescription: a.Desc}
		case "strlist":
			attrs[a.Name] = dschema.ListAttribute{ElementType: types.StringType, Computed: comp, Sensitive: a.Sens, MarkdownDescription: a.Desc}
		default:
			attrs[a.Name] = dschema.StringAttribute{Required: req, Computed: comp, Sensitive: a.Sens, MarkdownDescription: a.Desc}
		}
	}
	resp.Schema = dschema.Schema{MarkdownDescription: "Reads " + s.Type + " (read-only). " + s.Desc, Attributes: attrs}
}

func (d *genDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	s := d.spec
	id := s.Type
	if s.Kind == "coll" || s.Kind == "dataonly" {
		var v types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &v)...)
		id = v.ValueString()
	}
	vals := s.paramVals(ctx, req.Config, id, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	r := &genResource{spec: s, client: d.client}
	obj, found, err := r.fetch(ctx, vals, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+s.Type, err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Not found", s.Type+" "+id+" does not exist")
		return
	}
	resp.State.Raw = req.Config.Raw
	if s.Kind == "keyed" {
		id = ""
		for i, a := range s.params() {
			if i > 0 {
				id += "/"
			}
			id += vals[a.Name]
		}
	}
	r.apply(ctx, &resp.State, obj, id, false, &resp.Diagnostics)
}
