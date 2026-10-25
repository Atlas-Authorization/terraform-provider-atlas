package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The `per_call_audience` object on an identity integration: lets one
// integration mint a token whose `aud` is chosen per call (for example one
// audience per site), from a template with one `{param}` placeholder.

type perCallAudienceModel struct {
	Param       types.String `tfsdk:"param"`
	Template    types.String `tfsdk:"template"`
	Allowed     types.Set    `tfsdk:"allowed"`
	AnyMatching types.Bool   `tfsdk:"any_matching"`
}

const perCallAudienceBlockDesc = "Lets a token mint name its own audience per call (`getToken({ for, cbid })`, `?audience=`, or " +
	"`?<param>=`). A preset already allows its own parameter for the configured value only (chatly `channel`, forumulate and " +
	"referalli `org_id`, cookiemunch `cbid`); set this block to allow more values, or to change them. Omit the block to keep the " +
	"preset's default (a custom integration then mints its configured audience only); set `per_call_audience_disabled = true` to turn it off."

func perCallAudienceBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: perCallAudienceBlockDesc,
		Attributes: map[string]schema.Attribute{
			"param": schema.StringAttribute{
				MarkdownDescription: "The query parameter that carries the value (1-32 lowercase letters, digits or `_`, starting with a letter). " +
					"A preset has its own and needs none here.",
				Optional: true,
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Custom integrations only (required there): the audience with `{param}` in it once, e.g. `acme:{tenant}`. " +
					"A preset's audience is always `<preset>:{param}`; leave this out.",
				Optional: true,
			},
			"allowed": schema.SetAttribute{
				MarkdownDescription: "Values of `param` (not whole audiences) a call may name, besides the configured one. At most 50 values, each 1-128 letters, digits, `.`, `_` or `-`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
			"any_matching": schema.BoolAttribute{
				MarkdownDescription: "Allow any value that fits the audience format. A custom integration needs `audience_format` for this. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

const perCallAudienceDisabledDesc = "Turn the per-call audience off, including a preset's default (sends `per_call_audience: null`). " +
	"Cannot be combined with a `per_call_audience` block. Default `false`."

func perCallAudienceDataSourceAttribute() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		Computed:            true,
		MarkdownDescription: "The per-call audience in effect (a preset's default included), or null when off.",
		Attributes: map[string]dschema.Attribute{
			"param":        dschema.StringAttribute{Computed: true, MarkdownDescription: "The query parameter that carries the value."},
			"template":     dschema.StringAttribute{Computed: true, MarkdownDescription: "The audience template, with `{param}` in it."},
			"allowed":      dschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Values a call may name besides the configured one."},
			"any_matching": dschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether any value that fits the audience format is allowed."},
		},
	}
}

// perCallAudienceToAPI builds the request object; nil when the block is absent.
func perCallAudienceToAPI(ctx context.Context, m *perCallAudienceModel, diags *diag.Diagnostics) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := map[string]interface{}{}
	if cmpSet(m.Param) {
		out["param"] = m.Param.ValueString()
	}
	if cmpSet(m.Template) {
		out["template"] = m.Template.ValueString()
	}
	allowed := setToStringSlice(ctx, m.Allowed, diags)
	if allowed == nil {
		allowed = []string{}
	}
	out["allowed"] = allowed
	out["any_matching"] = !m.AnyMatching.IsNull() && !m.AnyMatching.IsUnknown() && m.AnyMatching.ValueBool()
	return out
}

// perCallAudienceFromAPI reads the API's object back. prior is what the plan or
// state held: a param or template the configuration left out stays null (the
// API fills them in for a preset). The preset's own default is shown only when
// showDefault is set (data sources) or the block is already configured.
func perCallAudienceFromAPI(ctx context.Context, prior *perCallAudienceModel, in map[string]interface{}, showDefault bool, diags *diag.Diagnostics) *perCallAudienceModel {
	if in == nil {
		return nil
	}
	if isDefault, _ := in["default"].(bool); isDefault && prior == nil && !showDefault {
		return nil
	}
	allowed := []string{}
	if raw, ok := in["allowed"].([]interface{}); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				allowed = append(allowed, s)
			}
		}
	}
	out := &perCallAudienceModel{
		Param:       apiString(in["param"]),
		Template:    apiString(in["template"]),
		Allowed:     stringSliceToSet(ctx, allowed, diags),
		AnyMatching: types.BoolValue(false),
	}
	if b, ok := in["any_matching"].(bool); ok {
		out.AnyMatching = types.BoolValue(b)
	}
	if prior != nil && !showDefault {
		if prior.Param.IsNull() {
			out.Param = types.StringNull()
		}
		if prior.Template.IsNull() {
			out.Template = types.StringNull()
		}
	}
	return out
}

func validatePerCallAudience(block *perCallAudienceModel, disabled types.Bool, diags *diag.Diagnostics) {
	if block != nil && !disabled.IsNull() && !disabled.IsUnknown() && disabled.ValueBool() {
		diags.AddAttributeError(path.Root("per_call_audience_disabled"), "Conflicting settings",
			"per_call_audience_disabled = true turns the per-call audience off; remove the per_call_audience block or set it to false.")
	}
}
