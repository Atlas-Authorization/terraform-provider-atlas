package provider

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
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
	})
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	put := api.last(http.MethodPut, "/v1/instance/cookie_consent")
	if put == nil || put.Body["script_url"] != "https://consent.example/uc.js" || put.Body["site_id"] != "abc" {
		t.Fatalf("PUT body = %+v", put)
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
