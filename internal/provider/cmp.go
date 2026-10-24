package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The `cmp` object: a consent manager's script (and, optionally, the consent
// API it exposes on the page) that the hosted pages load. It is set on an
// identity integration and on an outbound partner link, with the same shape.

type consentAPIModel struct {
	Global        types.String `tfsdk:"global"`
	HasMethod     types.String `tfsdk:"has_method"`
	RequestMethod types.String `tfsdk:"request_method"`
	CategoryMap   types.Map    `tfsdk:"category_map"`
}

type cmpModel struct {
	ScriptURL      types.String     `tfsdk:"script_url"`
	ScriptID       types.String     `tfsdk:"script_id"`
	SiteID         types.String     `tfsdk:"site_id"`
	BlockingMode   types.String     `tfsdk:"blocking_mode"`
	ExtraOrigins   types.List       `tfsdk:"extra_origins"`
	SettingsPreset types.String     `tfsdk:"settings_preset"`
	ConsentAPI     *consentAPIModel `tfsdk:"consent_api"`
}

const cmpBlockDesc = "A consent manager (any CMP) the hosted sign-in pages load from this object when the instance has no " +
	"`cookie_consent` of its own. Omit the block to bring none (the Cookie Munch preset then keeps its derived default, " +
	"which is not shown in state). Removing the block clears it."

const consentAPIBlockDesc = "The consent API the CMP exposes on the page, so the hosted pages read and record consent through it " +
	"and not only through the banner."

func cmpBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: cmpBlockDesc,
		Attributes: map[string]schema.Attribute{
			"script_url": schema.StringAttribute{
				MarkdownDescription: "The CMP script (https). Required when the block is set.",
				Optional:            true,
			},
			"script_id": schema.StringAttribute{MarkdownDescription: "The script element id, if the CMP needs one.", Optional: true},
			"site_id":   schema.StringAttribute{MarkdownDescription: "The `data-cbid` / site id the script needs.", Optional: true},
			"blocking_mode": schema.StringAttribute{
				MarkdownDescription: "The CMP's blocking mode, if it has one.",
				Optional:            true,
			},
			"extra_origins": schema.ListAttribute{
				MarkdownDescription: "Extra https origins the CMP calls (CSP), at most 10. Atlas stores each as a bare origin (no path or trailing slash).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"settings_preset": schema.StringAttribute{
				MarkdownDescription: "`cookiemunch`, `cookiebot`, `onetrust` or `custom`.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"consent_api": schema.SingleNestedBlock{
				MarkdownDescription: consentAPIBlockDesc,
				Attributes: map[string]schema.Attribute{
					"global": schema.StringAttribute{
						MarkdownDescription: "The `window` property holding the API (a plain identifier). Required when the block is set.",
						Optional:            true,
					},
					"has_method": schema.StringAttribute{
						MarkdownDescription: "The method answering `true`, `false` or `null` (no decision yet). Default `has`.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("has"),
					},
					"request_method": schema.StringAttribute{
						MarkdownDescription: "The method that asks the visitor and records the answer. Unset: Atlas never asks.",
						Optional:            true,
					},
					"category_map": schema.MapAttribute{
						MarkdownDescription: "Atlas category name to the CMP's own name, when they differ.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
		},
	}
}

// cmpDataSourceAttribute is the read-only form of the block, for data sources.
func cmpDataSourceAttribute() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		Computed:            true,
		MarkdownDescription: "The consent manager the hosted pages load from this object, or null. For the Cookie Munch preset this includes the derived default.",
		Attributes: map[string]dschema.Attribute{
			"script_url":      dschema.StringAttribute{Computed: true, MarkdownDescription: "The CMP script (https)."},
			"script_id":       dschema.StringAttribute{Computed: true, MarkdownDescription: "The script element id."},
			"site_id":         dschema.StringAttribute{Computed: true, MarkdownDescription: "The site id the script needs."},
			"blocking_mode":   dschema.StringAttribute{Computed: true, MarkdownDescription: "The CMP's blocking mode."},
			"extra_origins":   dschema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Extra https origins the CMP calls."},
			"settings_preset": dschema.StringAttribute{Computed: true, MarkdownDescription: "The settings preset."},
			"consent_api": dschema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: consentAPIBlockDesc,
				Attributes: map[string]dschema.Attribute{
					"global":         dschema.StringAttribute{Computed: true, MarkdownDescription: "The `window` property holding the API."},
					"has_method":     dschema.StringAttribute{Computed: true, MarkdownDescription: "The method answering true, false or null."},
					"request_method": dschema.StringAttribute{Computed: true, MarkdownDescription: "The method that asks the visitor."},
					"category_map":   dschema.MapAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Atlas category name to the CMP's own name."},
				},
			},
		},
	}
}

func cmpSet(v types.String) bool { return !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" }

// cmpToAPI builds the `cmp` request object. Absent optional fields are sent as
// null so a changed block replaces the stored one.
func cmpToAPI(ctx context.Context, m *cmpModel, diags *diag.Diagnostics) map[string]interface{} {
	if m == nil {
		return nil
	}
	if !cmpSet(m.ScriptURL) {
		diags.AddAttributeError(pathCmp("script_url"), "script_url is required", "A `cmp` block needs `script_url` (an https URL).")
		return nil
	}
	str := func(v types.String) interface{} {
		if v.IsNull() || v.IsUnknown() {
			return nil
		}
		return v.ValueString()
	}
	origins := listToStringSlice(ctx, m.ExtraOrigins, diags)
	if origins == nil {
		origins = []string{}
	}
	out := map[string]interface{}{
		"script_url":      m.ScriptURL.ValueString(),
		"script_id":       str(m.ScriptID),
		"site_id":         str(m.SiteID),
		"blocking_mode":   str(m.BlockingMode),
		"extra_origins":   origins,
		"settings_preset": str(m.SettingsPreset),
	}
	if m.ConsentAPI != nil {
		if !cmpSet(m.ConsentAPI.Global) {
			diags.AddAttributeError(pathCmp("consent_api", "global"), "global is required", "A `consent_api` block needs `global` (the `window` property holding the API).")
			return nil
		}
		ca := map[string]interface{}{
			"global":         m.ConsentAPI.Global.ValueString(),
			"has_method":     str(m.ConsentAPI.HasMethod),
			"request_method": str(m.ConsentAPI.RequestMethod),
			"category_map":   nil,
		}
		if cm := mapToStringMap(ctx, m.ConsentAPI.CategoryMap, diags); cm != nil {
			ca["category_map"] = cm
		}
		out["consent_api"] = ca
	}
	return out
}

func apiString(v interface{}) types.String {
	if s, ok := v.(string); ok {
		return types.StringValue(s)
	}
	return types.StringNull()
}

// cmpFromAPI reads the API's `cmp` object back into the block model; nil in,
// nil out.
func cmpFromAPI(ctx context.Context, in map[string]interface{}, diags *diag.Diagnostics) *cmpModel {
	if in == nil {
		return nil
	}
	origins := []string{}
	if raw, ok := in["extra_origins"].([]interface{}); ok {
		for _, o := range raw {
			if s, ok := o.(string); ok {
				origins = append(origins, s)
			}
		}
	}
	out := &cmpModel{
		ScriptURL:      apiString(in["script_url"]),
		ScriptID:       apiString(in["script_id"]),
		SiteID:         apiString(in["site_id"]),
		BlockingMode:   apiString(in["blocking_mode"]),
		ExtraOrigins:   stringSliceToList(ctx, origins, diags),
		SettingsPreset: apiString(in["settings_preset"]),
	}
	if ca, ok := in["consent_api"].(map[string]interface{}); ok && ca != nil {
		cm := map[string]string{}
		hasMap := false
		if raw, ok := ca["category_map"].(map[string]interface{}); ok {
			hasMap = true
			for k, v := range raw {
				if s, ok := v.(string); ok {
					cm[k] = s
				}
			}
		}
		m := &consentAPIModel{
			Global:        apiString(ca["global"]),
			HasMethod:     apiString(ca["has_method"]),
			RequestMethod: apiString(ca["request_method"]),
			CategoryMap:   types.MapNull(types.StringType),
		}
		if hasMap {
			m.CategoryMap = stringMapToMap(ctx, cm, diags)
		}
		out.ConsentAPI = m
	}
	return out
}

func pathCmp(parts ...string) path.Path {
	p := path.Root("cmp")
	for _, s := range parts {
		p = p.AtName(s)
	}
	return p
}
