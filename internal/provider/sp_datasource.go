package provider

import (
	"context"
	"fmt"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// The 2026-10 data sources reuse their resource's schema and mapping: a data
// source is the resource's attributes, read-only, keyed by one lookup attribute
// (or none, for a singleton), minus the secrets and write-only inputs.

// dataSourceSpec describes one data source built from a resource.
type dataSourceSpec struct {
	// name is the type-name suffix: atlas_<name>.
	name string
	// doc is the data source's description.
	doc string
	// schema returns the resource schema the data source derives from.
	schema func() rschema.Schema
	// lookup names the attribute that selects the object ("" for a singleton).
	lookup string
	// skip names resource attributes left out of the data source.
	skip []string
	// load reads the object (for id when lookup is set) and returns it shaped
	// like the resource's state.
	load func(ctx context.Context, c *client.Client, id string, diags *diag.Diagnostics) (tfsdk.State, bool)
}

type specDataSource struct {
	spec   dataSourceSpec
	client *client.Client
}

var (
	_ datasource.DataSource              = &specDataSource{}
	_ datasource.DataSourceWithConfigure = &specDataSource{}
)

func newSpecDataSource(spec dataSourceSpec) func() datasource.DataSource {
	return func() datasource.DataSource { return &specDataSource{spec: spec} }
}

func (d *specDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.spec.name
}

func (d *specDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *specDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.dataSourceSchema()
}

func (d *specDataSource) dataSourceSchema() dschema.Schema {
	skip := map[string]bool{}
	for _, s := range d.spec.skip {
		skip[s] = true
	}
	rs := d.spec.schema()
	attrs := map[string]dschema.Attribute{}
	for name, a := range rs.Attributes {
		if skip[name] || a.IsWriteOnly() {
			continue
		}
		required := name == d.spec.lookup
		switch v := a.(type) {
		case rschema.StringAttribute:
			attrs[name] = dschema.StringAttribute{Required: required, Computed: !required, Sensitive: v.Sensitive, MarkdownDescription: v.MarkdownDescription}
		case rschema.BoolAttribute:
			attrs[name] = dschema.BoolAttribute{Required: required, Computed: !required, Sensitive: v.Sensitive, MarkdownDescription: v.MarkdownDescription}
		case rschema.Int64Attribute:
			attrs[name] = dschema.Int64Attribute{Required: required, Computed: !required, Sensitive: v.Sensitive, MarkdownDescription: v.MarkdownDescription}
		case rschema.SetAttribute:
			attrs[name] = dschema.SetAttribute{Computed: true, ElementType: v.ElementType, Sensitive: v.Sensitive, MarkdownDescription: v.MarkdownDescription}
		case rschema.ListAttribute:
			attrs[name] = dschema.ListAttribute{Computed: true, ElementType: v.ElementType, Sensitive: v.Sensitive, MarkdownDescription: v.MarkdownDescription}
		case rschema.MapAttribute:
			attrs[name] = dschema.MapAttribute{Computed: true, ElementType: v.ElementType, Sensitive: v.Sensitive, MarkdownDescription: v.MarkdownDescription}
		default:
			panic(fmt.Sprintf("data source %s: unsupported attribute type %T for %s", d.spec.name, a, name))
		}
	}
	return dschema.Schema{MarkdownDescription: d.spec.doc, Attributes: attrs}
}

func (d *specDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	id := ""
	if d.spec.lookup != "" {
		var v types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(d.spec.lookup), &v)...)
		if resp.Diagnostics.HasError() {
			return
		}
		id = v.ValueString()
	}
	loaded, ok := d.spec.load(ctx, d.client, id, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	// Copy the attributes the data source keeps from the resource-shaped state.
	var all map[string]tftypes.Value
	if err := loaded.Raw.As(&all); err != nil {
		resp.Diagnostics.AddError("Unable to read "+d.spec.name, err.Error())
		return
	}
	objType, ok := resp.State.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		resp.Diagnostics.AddError("Unable to read "+d.spec.name, "the data source schema is not an object")
		return
	}
	kept := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name := range objType.AttributeTypes {
		kept[name] = all[name]
	}
	resp.State.Raw = tftypes.NewValue(objType, kept)
}

// stateOf shapes a populated resource model as the resource's state, so a data
// source can project it. Every collection attribute of the model must be set.
func stateOf(ctx context.Context, rs rschema.Schema, model interface{}, diags *diag.Diagnostics) tfsdk.State {
	st := tfsdk.State{Schema: rs, Raw: tftypes.NewValue(rs.Type().TerraformType(ctx), nil)}
	diags.Append(st.Set(ctx, model)...)
	return st
}

// notFound turns a 404 into a data source error that names the object.
func lookupFailed(diags *diag.Diagnostics, what string, err error) {
	if client.IsNotFound(err) {
		diags.AddError("Not found", what+" does not exist.")
		return
	}
	diags.AddError("Unable to read "+what, err.Error())
}

// dataSourceSpecs lists every derived data source.
func dataSourceSpecs() []dataSourceSpec {
	return []dataSourceSpec{
		{
			name:   "secret_key",
			doc:    "Look up a secret key by id. The plaintext secret is never available.",
			schema: secretKeySchema,
			lookup: "id",
			skip:   []string{"secret"},
			load: func(ctx context.Context, c *client.Client, id string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				k, err := c.GetSecretKey(ctx, id)
				if err != nil {
					lookupFailed(diags, "secret key "+id, err)
					return tfsdk.State{}, false
				}
				var m secretKeyModel
				fillSecretKey(ctx, &m, k, diags)
				return stateOf(ctx, secretKeySchema(), &m, diags), true
			},
		},
		{
			name:   "partner_link",
			doc:    "Look up an inbound partner link by id. The client secret is never available.",
			schema: partnerLinkSchema,
			lookup: "id",
			skip:   []string{"client_secret", "rotate_secret_trigger"},
			load: func(ctx context.Context, c *client.Client, id string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				l, err := c.GetPartnerLink(ctx, id)
				if err != nil {
					lookupFailed(diags, "partner link "+id, err)
					return tfsdk.State{}, false
				}
				var m partnerLinkModel
				fillPartnerLink(ctx, &m, l, diags)
				return stateOf(ctx, partnerLinkSchema(), &m, diags), true
			},
		},
		{
			name:   "outbound_partner_link",
			doc:    "Look up the outbound partner link to a partner. The client secret is never available.",
			schema: outboundPartnerLinkSchema,
			lookup: "partner",
			skip:   []string{"client_secret", "id"},
			load: func(ctx context.Context, c *client.Client, id string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				l, err := c.GetOutboundPartnerLink(ctx, id)
				if err != nil {
					lookupFailed(diags, "outbound partner link "+id, err)
					return tfsdk.State{}, false
				}
				var m outboundPartnerLinkModel
				fillOutbound(ctx, &m, l, diags)
				return stateOf(ctx, outboundPartnerLinkSchema(), &m, diags), true
			},
		},
		{
			name:   "integration",
			doc:    "Look up an identity integration (a preset or a custom one) by id. An integration that is not connected is an error.",
			schema: func() rschema.Schema { return integrationSchema("") },
			lookup: "integration",
			skip:   []string{"audience_format", "id"},
			load: func(ctx context.Context, c *client.Client, id string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				i, err := c.GetIntegration(ctx, id)
				if err != nil {
					lookupFailed(diags, "integration "+id, err)
					return tfsdk.State{}, false
				}
				if !i.Connected {
					diags.AddError("Not found", "The "+id+" integration is not connected.")
					return tfsdk.State{}, false
				}
				var m integrationModel
				fillIntegration(ctx, &m, i, diags)
				return stateOf(ctx, integrationSchema(""), &m, diags), true
			},
		},
		{
			name:   "webhook_endpoint",
			doc:    "Look up a webhook endpoint by id, with its delivery health. The signing secret is never available.",
			schema: webhookEndpointSchemaOnly,
			lookup: "id",
			skip:   []string{"secret", "rotate_secret_trigger", "destination_credentials"},
			load: func(ctx context.Context, c *client.Client, id string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				e, err := c.GetWebhookEndpointFull(ctx, id)
				if err != nil {
					lookupFailed(diags, "webhook endpoint "+id, err)
					return tfsdk.State{}, false
				}
				r := &webhookEndpointResource{}
				var m webhookEndpointModel
				m.DestinationCreds = types.MapNull(types.StringType)
				r.mapToState(ctx, e, &m, diags, false)
				return stateOf(ctx, webhookEndpointSchemaOnly(), &m, diags), true
			},
		},
		{
			name:   "audit_export",
			doc:    "Look up an audit export by id. The signing secret and S3 secret key are never available.",
			schema: auditExportSchema,
			lookup: "id",
			skip:   []string{"secret", "secret_access_key", "rotate_secret_trigger", "backfill_hours"},
			load: func(ctx context.Context, c *client.Client, id string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				e, err := c.GetAuditExport(ctx, id)
				if err != nil {
					lookupFailed(diags, "audit export "+id, err)
					return tfsdk.State{}, false
				}
				var m auditExportModel
				fillAuditExport(ctx, &m, e, diags)
				return stateOf(ctx, auditExportSchema(), &m, diags), true
			},
		},
		{
			name:   "status_subscription",
			doc:    "The instance's status subscription. The private feed URLs are marked sensitive.",
			schema: statusSubscriptionSchema,
			skip:   []string{"rotate_feed_trigger"},
			load: func(ctx context.Context, c *client.Client, _ string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				s, err := c.GetStatusSubscription(ctx)
				if err != nil {
					lookupFailed(diags, "the status subscription", err)
					return tfsdk.State{}, false
				}
				var m statusSubscriptionModel
				fillStatusSubscription(ctx, &m, s, diags)
				return stateOf(ctx, statusSubscriptionSchema(), &m, diags), true
			},
		},
		{
			name:   "status_dependency",
			doc:    "Look up a declared upstream status dependency by id, with its current state.",
			schema: statusDependencySchema,
			lookup: "id",
			load: func(ctx context.Context, c *client.Client, id string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				dep, err := c.GetStatusDependency(ctx, id)
				if err != nil {
					lookupFailed(diags, "status dependency "+id, err)
					return tfsdk.State{}, false
				}
				var m statusDependencyModel
				fillStatusDependency(ctx, &m, dep, diags)
				return stateOf(ctx, statusDependencySchema(), &m, diags), true
			},
		},
		{
			name:   "privacy_orchestrator",
			doc:    "The instance's data-rights orchestrator. Secrets are never available.",
			schema: privacyOrchestratorSchema,
			skip:   []string{"signing_secret", "completion_webhook_secret", "api_key_wo", "api_key_wo_version"},
			load: func(ctx context.Context, c *client.Client, _ string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				o, err := c.GetPrivacyOrchestrator(ctx)
				if err != nil {
					lookupFailed(diags, "the privacy orchestrator", err)
					return tfsdk.State{}, false
				}
				if !o.Configured {
					diags.AddError("Not found", "This instance has no privacy orchestrator configured.")
					return tfsdk.State{}, false
				}
				var m privacyOrchestratorModel
				fillPrivacyOrchestrator(&m, o)
				return stateOf(ctx, privacyOrchestratorSchema(), &m, diags), true
			},
		},
		{
			name:   "referral_settings",
			doc:    "The instance's referral capture settings. Secrets are never available.",
			schema: referralSettingsSchema,
			skip:   []string{"signing_secret_wo", "signing_secret_wo_version", "api_key_wo", "api_key_wo_version"},
			load: func(ctx context.Context, c *client.Client, _ string, diags *diag.Diagnostics) (tfsdk.State, bool) {
				s, err := c.GetReferralSettings(ctx)
				if err != nil {
					lookupFailed(diags, "the referral settings", err)
					return tfsdk.State{}, false
				}
				var m referralSettingsModel
				fillReferralSettings(ctx, &m, s, diags)
				return stateOf(ctx, referralSettingsSchema(), &m, diags), true
			},
		},
	}
}

// webhookEndpointSchemaOnly returns the webhook endpoint resource's schema.
func webhookEndpointSchemaOnly() rschema.Schema {
	var resp resource.SchemaResponse
	(&webhookEndpointResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp.Schema
}

// dataSourceFactories builds the factory list for the provider.
func specDataSourceFactories() []func() datasource.DataSource {
	specs := dataSourceSpecs()
	out := make([]func() datasource.DataSource, 0, len(specs)+1)
	for _, s := range specs {
		out = append(out, newSpecDataSource(s))
	}
	return out
}
