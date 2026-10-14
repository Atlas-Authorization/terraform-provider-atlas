package provider

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// appResponse is the API projection the fake returns: branding always carries
// every field (nulls for unset ones), the default theme and resolved_design.
func appResponse(extra map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{
		"object": "atlas_id_app", "id": "oac_1", "client_id": "client_abc", "name": "Acme",
		"logo_url": nil, "homepage_url": nil,
		"branding": map[string]interface{}{
			"theme": "auto", "accent_color": "#112233", "background_color": nil,
			"privacy_policy_url": nil, "terms_of_service_url": nil,
			"copy":            map[string]interface{}{"sign_in_title": "Welcome to {app}", "sign_in_subtitle": nil, "sign_up_title": nil, "sign_up_subtitle": nil, "consent_message": nil, "support_url": nil, "support_label": nil},
			"design":          nil,
			"resolved_design": map[string]interface{}{"version": 1, "palette": map[string]interface{}{"primary": "#112233"}},
		},
		"redirect_uris":                        []string{"https://acme.example/cb"},
		"post_logout_redirect_uris":            []string{"https://acme.example/bye"},
		"backchannel_logout_uri":               "https://acme.example/bcl",
		"backchannel_logout_session_required":  true,
		"frontchannel_logout_uri":              nil,
		"frontchannel_logout_session_required": false,
		"allowed_scopes":                       []string{"openid", "profile", "email"},
		"type":                                 "confidential",
		"secret_prefix":                        "oacs_ab",
		"managed_by_instance_id":               nil,
		"created_at":                           1, "updated_at": 2,
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func baseAppModel(ctx context.Context) atlasIDAppModel {
	list := func(v ...string) types.List { l, _ := types.ListValueFrom(ctx, types.StringType, v); return l }
	return atlasIDAppModel{
		ID: types.StringUnknown(), ClientID: types.StringUnknown(), Name: types.StringValue("Acme"),
		Type:                              types.StringUnknown(),
		RedirectURIs:                      list("https://acme.example/cb"),
		PostLogoutRedirectURIs:            list("https://acme.example/bye"),
		BackchannelLogoutURI:              types.StringValue("https://acme.example/bcl"),
		BackchannelLogoutSessionRequired:  types.BoolValue(true),
		FrontchannelLogoutURI:             types.StringNull(),
		FrontchannelLogoutSessionRequired: types.BoolUnknown(),
		AllowedScopes:                     types.ListUnknown(types.StringType),
		LogoURL:                           types.StringValue("https://acme.example/logo.png"),
		HomepageURL:                       types.StringNull(),
		Branding:                          newJSONObjectValue(`{ "copy": {"sign_in_title": "Welcome to {app}"}, "accent_color": "#112233" }`),
		RotateSecretTrigger:               types.StringNull(),
		ClientSecret:                      types.StringUnknown(),
		SecretPrefix:                      types.StringUnknown(),
		ManagedByInstanceID:               types.StringUnknown(),
		CreatedAt:                         types.Int64Unknown(),
		UpdatedAt:                         types.Int64Unknown(),
	}
}

func TestAtlasIDApp_createUpdateRotate(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		switch {
		case call.Method == http.MethodPost && call.Path == "/v1/platform/atlas_id_apps":
			return 201, appResponse(map[string]interface{}{"client_secret": "oacs_first", "logo_url": "https://acme.example/logo.png"})
		case call.Method == http.MethodPatch:
			return 200, appResponse(map[string]interface{}{"logo_url": nil})
		case call.Method == http.MethodPost && call.Path == "/v1/platform/atlas_id_apps/oac_1/rotate_secret":
			return 200, appResponse(map[string]interface{}{"client_secret": "oacs_second", "secret_prefix": "oacs_cd"})
		case call.Method == http.MethodGet:
			return 200, appResponse(map[string]interface{}{"user_count": 3})
		}
		return 404, map[string]interface{}{"errors": []map[string]string{{"code": "NOT_FOUND", "message": "x"}}}
	})
	r := NewAtlasIDAppResource()
	configureResource(t, r, c)

	state, cresp := doCreate(t, r, planFrom(t, r, baseAppModel(ctx)))
	mustOK(t, "create", cresp.Diagnostics.HasError(), cresp.Diagnostics)
	call := api.last(http.MethodPost, "/v1/platform/atlas_id_apps")
	if call == nil || call.Auth != "Bearer psk_test_key" {
		t.Fatalf("create must be signed with the platform key, got %+v", call)
	}
	if call.Body["backchannel_logout_uri"] != "https://acme.example/bcl" || call.Body["backchannel_logout_session_required"] != true {
		t.Errorf("logout fields not sent: %v", call.Body)
	}
	if _, sent := call.Body["homepage_url"]; sent {
		t.Errorf("a null field must be omitted on create: %v", call.Body)
	}
	if got := call.Body["post_logout_redirect_uris"]; !reflect.DeepEqual(got, []interface{}{"https://acme.example/bye"}) {
		t.Errorf("post_logout_redirect_uris = %v", got)
	}
	if b, _ := call.Body["branding"].(map[string]interface{}); b["accent_color"] != "#112233" {
		t.Errorf("branding not sent as an object: %v", call.Body["branding"])
	}

	var m atlasIDAppModel
	mustOK(t, "get state", state.Get(ctx, &m).HasError(), nil)
	if m.ClientSecret.ValueString() != "oacs_first" || m.ClientID.ValueString() != "client_abc" {
		t.Errorf("secret/client id not stored: %+v", m)
	}
	// The API branding (nulls, default theme, resolved_design) must project to
	// the configured document.
	eq, _ := baseAppModel(ctx).Branding.StringSemanticEquals(ctx, m.Branding)
	if !eq {
		t.Errorf("branding read-back %s is not semantically equal to the configuration", m.Branding.ValueString())
	}

	// Refresh keeps the secret (never re-returned) and stays equal.
	refreshed, rresp := doRead(t, r, state)
	mustOK(t, "read", rresp.Diagnostics.HasError(), rresp.Diagnostics)
	var rm atlasIDAppModel
	refreshed.Get(ctx, &rm)
	if rm.ClientSecret.ValueString() != "oacs_first" {
		t.Errorf("read lost the client secret")
	}

	// Update: drop logo_url and accent_color, change rotate trigger.
	next := m
	next.LogoURL = types.StringNull()
	next.Branding = newJSONObjectValue(`{"copy":{"sign_in_title":"Welcome to {app}"}}`)
	next.RotateSecretTrigger = types.StringValue("2026-10")
	next.ClientSecret = types.StringUnknown()
	updated, uresp := doUpdate(t, r, planFrom(t, r, next), state)
	mustOK(t, "update", uresp.Diagnostics.HasError(), uresp.Diagnostics)
	patch := api.last(http.MethodPatch, "/v1/platform/atlas_id_apps/oac_1")
	if patch == nil {
		t.Fatal("update did not PATCH")
	}
	if v, ok := patch.Body["logo_url"]; !ok || v != nil {
		t.Errorf("removed logo_url must be sent as null, body=%v", patch.Body)
	}
	b, _ := patch.Body["branding"].(map[string]interface{})
	if v, ok := b["accent_color"]; !ok || v != nil {
		t.Errorf("removed branding key must be cleared with null, branding=%v", b)
	}
	if api.count(http.MethodPost, "/v1/platform/atlas_id_apps/oac_1/rotate_secret") != 1 {
		t.Error("changing rotate_secret_trigger must rotate once")
	}
	var um atlasIDAppModel
	updated.Get(ctx, &um)
	if um.ClientSecret.ValueString() != "oacs_second" || um.SecretPrefix.ValueString() != "oacs_cd" {
		t.Errorf("rotated secret not stored: %+v", um.ClientSecret)
	}

	// An update without a trigger change keeps the secret and does not rotate.
	same := um
	same.Name = types.StringValue("Acme 2")
	_, uresp = doUpdate(t, r, planFrom(t, r, same), updated)
	mustOK(t, "update 2", uresp.Diagnostics.HasError(), uresp.Diagnostics)
	if api.count(http.MethodPost, "/v1/platform/atlas_id_apps/oac_1/rotate_secret") != 1 {
		t.Error("an update without a trigger change must not rotate")
	}

	del := doDelete(t, r, updated)
	mustOK(t, "delete", del.Diagnostics.HasError(), del.Diagnostics)
	if api.last(http.MethodDelete, "/v1/platform/atlas_id_apps/oac_1") == nil {
		t.Error("delete did not call DELETE")
	}
}

func TestAtlasIDApp_requiresPlatformKey(t *testing.T) {
	c, _ := newFakeAPI(t, func(apiCall) (int, interface{}) { return 200, appResponse(nil) })
	c.PlatformKey = ""
	r := NewAtlasIDAppResource()
	configureResource(t, r, c)
	_, resp := doCreate(t, r, planFrom(t, r, baseAppModel(context.Background())))
	if !resp.Diagnostics.HasError() {
		t.Fatal("create without a platform key must fail")
	}
}

func TestJSONObjectSemanticEquality(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		a, b string
		want bool
	}{
		{`{"a":1,"b":{"c":"x"}}`, "{\n  \"b\": {\"c\": \"x\"},\n  \"a\": 1\n}", true},
		{`{"a":1}`, `{"a":1,"b":null,"c":{"d":null}}`, true},
		{`{"a":1}`, `{"a":2}`, false},
		{`{"a":[1,2]}`, `{"a":[2,1]}`, false},
		{`{"a":{"b":1}}`, `{"a":{"b":1,"c":2}}`, false},
	}
	for _, tc := range cases {
		got, _ := newJSONObjectValue(tc.a).StringSemanticEquals(ctx, basetypes.StringValuable(newJSONObjectValue(tc.b)))
		if got != tc.want {
			t.Errorf("semantic(%s, %s) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestMergePatchFor(t *testing.T) {
	prior := map[string]interface{}{"theme": "dark", "copy": map[string]interface{}{"sign_in_title": "x", "support_url": "https://s"}, "accent_color": "#000000"}
	desired := map[string]interface{}{"theme": "light", "copy": map[string]interface{}{"sign_in_title": "y"}}
	got := mergePatchFor(prior, desired)
	want := map[string]interface{}{
		"theme":        "light",
		"copy":         map[string]interface{}{"sign_in_title": "y", "support_url": nil},
		"accent_color": nil,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergePatchFor = %v, want %v", got, want)
	}
}

func TestProjectAppBrandingKeepsExplicitAutoTheme(t *testing.T) {
	raw := []byte(`{"theme":"auto","accent_color":null,"resolved_design":{"x":1}}`)
	if v := projectAppBranding(raw, newJSONObjectNull()); !v.IsNull() {
		t.Errorf("default theme with no config should project to null, got %s", v.ValueString())
	}
	if v := projectAppBranding(raw, newJSONObjectValue(`{"theme":"auto"}`)); v.ValueString() != `{"theme":"auto"}` {
		t.Errorf("an explicitly configured theme must be kept, got %s", v.ValueString())
	}
}
