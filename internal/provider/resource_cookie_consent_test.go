package provider

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestCookieConsent_lifecycle(t *testing.T) {
	ctx := context.Background()
	enabled := true
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodDelete || !enabled {
			return 200, map[string]interface{}{"object": "cookie_consent", "enabled": false}
		}
		return 200, map[string]interface{}{
			"object": "cookie_consent", "enabled": true,
			"script_url": "https://consent.example/uc.js", "script_id": "Cookiebot",
			"site_id": "abc", "blocking_mode": nil, "extra_origins": []string{},
		}
	})
	r := NewCookieConsentResource()
	configureResource(t, r, c)
	plan := planFrom(t, r, cookieConsentModel{
		ID: types.StringUnknown(), ScriptURL: types.StringValue("https://consent.example/uc.js"),
		ScriptID: types.StringValue("Cookiebot"), SiteID: types.StringValue("abc"),
		BlockingMode: types.StringNull(), ExtraOrigins: types.ListNull(types.StringType),
		SettingsLink: types.ObjectNull(cookieSettingsLinkAttrTypes),
	})
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	put := api.last(http.MethodPut, "/v1/instance/cookie_consent")
	if put == nil || put.Body["script_url"] != "https://consent.example/uc.js" || put.Body["site_id"] != "abc" {
		t.Fatalf("PUT body = %+v", put)
	}
	if v, sent := put.Body["settings_link"]; !sent || v != nil {
		t.Errorf("an unset settings_link must be sent as null (removes the link), got %v", v)
	}
	if !reflect.DeepEqual(put.Body["extra_origins"], []interface{}{}) {
		t.Errorf("extra_origins should be sent as [] (whole-setting replace), got %v", put.Body["extra_origins"])
	}
	var m cookieConsentModel
	state.Get(ctx, &m)
	if !m.ExtraOrigins.IsNull() || m.ID.ValueString() != "cookie_consent" {
		t.Errorf("unset extra_origins must stay null: %+v", m)
	}

	// Turned off out of band → removed from state.
	enabled = false
	_, rr := doRead(t, r, state)
	if !rr.State.Raw.IsNull() {
		t.Error("a disabled cookie consent must drop from state")
	}

	if d := doDelete(t, r, state); d.Diagnostics.HasError() || api.last(http.MethodDelete, "/v1/instance/cookie_consent") == nil {
		t.Errorf("delete must DELETE /v1/instance/cookie_consent: %v", d.Diagnostics)
	}
}

func settingsLinkObject(t *testing.T, label string, preset types.String, attrs map[string]string) types.Object {
	t.Helper()
	attributes := types.MapNull(types.StringType)
	if attrs != nil {
		attributes, _ = types.MapValueFrom(context.Background(), types.StringType, attrs)
	}
	obj, d := types.ObjectValue(cookieSettingsLinkAttrTypes, map[string]attr.Value{
		"label": types.StringValue(label), "preset": preset, "attributes": attributes,
		"rendered_attributes": types.MapUnknown(types.StringType),
	})
	if d.HasError() {
		t.Fatal(d)
	}
	return obj
}

func cookieConsentWithLink(link types.Object) cookieConsentModel {
	return cookieConsentModel{
		ID: types.StringUnknown(), ScriptURL: types.StringValue("https://consent.example/uc.js"),
		ScriptID: types.StringNull(), SiteID: types.StringNull(), BlockingMode: types.StringNull(),
		ExtraOrigins: types.ListNull(types.StringType), SettingsLink: link,
	}
}

func TestCookieConsent_settingsLink(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 200, map[string]interface{}{
			"object": "cookie_consent", "enabled": true, "script_url": "https://consent.example/uc.js",
			"script_id": nil, "site_id": nil, "blocking_mode": nil, "extra_origins": []string{},
			"settings_link": map[string]interface{}{
				"label": "Cookie settings", "preset": "cookiemunch", "attributes": map[string]string{},
				"rendered_attributes": map[string]string{"data-cc": "show-settings"},
			},
		}
	})
	r := NewCookieConsentResource()
	configureResource(t, r, c)
	plan := planFrom(t, r, cookieConsentWithLink(settingsLinkObject(t, "Cookie settings", types.StringValue("cookiemunch"), nil)))
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	put := api.last(http.MethodPut, "/v1/instance/cookie_consent")
	link, _ := put.Body["settings_link"].(map[string]interface{})
	if link["label"] != "Cookie settings" || link["preset"] != "cookiemunch" {
		t.Fatalf("settings_link sent = %v", put.Body["settings_link"])
	}
	if _, sent := link["attributes"]; sent {
		t.Error("unset attributes must be omitted")
	}

	var m cookieConsentModel
	state.Get(ctx, &m)
	var got cookieSettingsLinkModel
	m.SettingsLink.As(ctx, &got, basetypes.ObjectAsOptions{})
	rendered := mapToStringMap(ctx, got.RenderedAttributes, &resp.Diagnostics)
	if rendered["data-cc"] != "show-settings" {
		t.Errorf("rendered_attributes = %v", rendered)
	}
	if !got.Attributes.IsNull() {
		t.Error("an unset attributes map must stay null when the API returns {}")
	}

	refreshed, rr := doRead(t, r, state)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	var again cookieConsentModel
	refreshed.Get(ctx, &again)
	if !again.SettingsLink.Equal(m.SettingsLink) {
		t.Errorf("read changed settings_link: %v -> %v", m.SettingsLink, again.SettingsLink)
	}
}

func TestCookieConsent_settingsLinkValidation(t *testing.T) {
	ctx := context.Background()
	r := NewCookieConsentResource()
	validate := func(link types.Object) []string {
		p := planFrom(t, r, cookieConsentWithLink(link))
		var resp resource.ValidateConfigResponse
		r.(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: p.Schema, Raw: p.Raw}}, &resp)
		var out []string
		for _, d := range resp.Diagnostics.Errors() {
			out = append(out, d.Summary()+": "+d.Detail())
		}
		return out
	}
	none := types.StringNull()
	for name, tc := range map[string]struct {
		link types.Object
		want string // "" = valid
	}{
		"no link":            {types.ObjectNull(cookieSettingsLinkAttrTypes), ""},
		"preset only":        {settingsLinkObject(t, "Cookie settings", types.StringValue("onetrust"), nil), ""},
		"custom attributes":  {settingsLinkObject(t, "Cookies", types.StringValue("custom"), map[string]string{"data-cc": "show", "aria-label": "Open", "href": "#prefs"}), ""},
		"https href":         {settingsLinkObject(t, "Cookies", none, map[string]string{"href": "https://example.com/cookies"}), ""},
		"empty label":        {settingsLinkObject(t, "", types.StringValue("cookiebot"), nil), "label must be 1-60"},
		"long label":         {settingsLinkObject(t, strings.Repeat("x", 61), types.StringValue("cookiebot"), nil), "label must be 1-60"},
		"padded label":       {settingsLinkObject(t, " Cookies", types.StringValue("cookiebot"), nil), "leading or trailing"},
		"unknown preset":     {settingsLinkObject(t, "Cookies", types.StringValue("iubenda"), nil), "preset must be one of"},
		"neither":            {settingsLinkObject(t, "Cookies", none, nil), "needs a preset or attributes"},
		"event handler":      {settingsLinkObject(t, "Cookies", none, map[string]string{"onclick": "x()"}), "event handler"},
		"reserved namespace": {settingsLinkObject(t, "Cookies", none, map[string]string{"data-atlas-cmp-open": "x"}), "reserved"},
		"style":              {settingsLinkObject(t, "Cookies", none, map[string]string{"style": "color:red"}), "not an allowed attribute"},
		"http href":          {settingsLinkObject(t, "Cookies", none, map[string]string{"href": "http://example.com"}), "https URL or a #fragment"},
		"upper case":         {settingsLinkObject(t, "Cookies", none, map[string]string{"Data-CC": "x"}), "lower-case"},
		"long value":         {settingsLinkObject(t, "Cookies", none, map[string]string{"data-x": strings.Repeat("v", 201)}), "at most 200"},
	} {
		errs := validate(tc.link)
		joined := strings.Join(errs, " | ")
		if tc.want == "" && len(errs) > 0 {
			t.Errorf("%s: want valid, got %s", name, joined)
		}
		if tc.want != "" && !strings.Contains(joined, tc.want) {
			t.Errorf("%s: want an error containing %q, got %q", name, tc.want, joined)
		}
	}
	many := map[string]string{}
	for i := 0; i < 11; i++ {
		many["data-k"+string(rune('a'+i))] = "v"
	}
	if errs := validate(settingsLinkObject(t, "Cookies", none, many)); !strings.Contains(strings.Join(errs, ""), "At most 10") {
		t.Errorf("11 attributes must be refused, got %v", errs)
	}
}
