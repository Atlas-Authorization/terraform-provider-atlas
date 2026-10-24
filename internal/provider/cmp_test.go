package provider

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func cmpFixture(t *testing.T) *cmpModel {
	t.Helper()
	ctx := context.Background()
	var d diag.Diagnostics
	m := &cmpModel{
		ScriptURL:      strVal("https://cmp.example.com/banner.js"),
		ScriptID:       nullS(),
		SiteID:         strVal("site-1"),
		BlockingMode:   nullS(),
		ExtraOrigins:   stringSliceToList(ctx, []string{"https://cdn.example.com"}, &d),
		SettingsPreset: strVal("custom"),
		ConsentAPI: &consentAPIModel{
			Global:        strVal("__cmp"),
			HasMethod:     strVal("has"),
			RequestMethod: strVal("request"),
			CategoryMap:   stringMapToMap(ctx, map[string]string{"analytics": "statistics"}, &d),
		},
	}
	if d.HasError() {
		t.Fatal(d)
	}
	return m
}

// cmpAPIObject is what the API answers for cmpFixture (cmpView, snake_case).
func cmpAPIObject() map[string]interface{} {
	return map[string]interface{}{
		"script_url": "https://cmp.example.com/banner.js", "script_id": nil, "site_id": "site-1", "blocking_mode": nil,
		"extra_origins": []interface{}{"https://cdn.example.com"}, "settings_preset": "custom",
		"consent_api": map[string]interface{}{
			"global": "__cmp", "has_method": "has", "request_method": "request",
			"category_map": map[string]interface{}{"analytics": "statistics"},
		},
	}
}

func TestCmp_toAPIUsesTheAPIFieldNames(t *testing.T) {
	var d diag.Diagnostics
	got := cmpToAPI(context.Background(), cmpFixture(t), &d)
	if d.HasError() {
		t.Fatal(d)
	}
	if got["script_url"] != "https://cmp.example.com/banner.js" || got["site_id"] != "site-1" || got["script_id"] != nil || got["settings_preset"] != "custom" {
		t.Errorf("cmp body: %+v", got)
	}
	ca, _ := got["consent_api"].(map[string]interface{})
	if ca["global"] != "__cmp" || ca["has_method"] != "has" || ca["request_method"] != "request" {
		t.Errorf("consent_api body: %+v", ca)
	}
	if cm, _ := ca["category_map"].(map[string]string); cm["analytics"] != "statistics" {
		t.Errorf("category_map: %+v", ca["category_map"])
	}
}

func TestCmp_requiredFields(t *testing.T) {
	var d diag.Diagnostics
	m := cmpFixture(t)
	m.ScriptURL = nullS()
	if cmpToAPI(context.Background(), m, &d) != nil || !d.HasError() {
		t.Error("a cmp block without script_url must be refused")
	}
	d = diag.Diagnostics{}
	m = cmpFixture(t)
	m.ConsentAPI.Global = nullS()
	if cmpToAPI(context.Background(), m, &d) != nil || !d.HasError() {
		t.Error("a consent_api block without global must be refused")
	}
	if cmpToAPI(context.Background(), nil, &d) != nil {
		t.Error("no block, no cmp")
	}
}

func TestCmp_roundTrip(t *testing.T) {
	var d diag.Diagnostics
	got := cmpFromAPI(context.Background(), cmpAPIObject(), &d)
	if d.HasError() || got == nil || got.ConsentAPI == nil {
		t.Fatalf("fromAPI: %v %+v", d, got)
	}
	if got.ScriptURL.ValueString() != "https://cmp.example.com/banner.js" || !got.ScriptID.IsNull() || got.SiteID.ValueString() != "site-1" {
		t.Errorf("fields: %+v", got)
	}
	if got.ConsentAPI.Global.ValueString() != "__cmp" || got.ConsentAPI.CategoryMap.Elements()["analytics"].String() != `"statistics"` {
		t.Errorf("consent_api: %+v", got.ConsentAPI)
	}
	if cmpFromAPI(context.Background(), nil, &d) != nil {
		t.Error("nil in, nil out")
	}
	// No consent_api / null category_map.
	obj := cmpAPIObject()
	obj["consent_api"] = nil
	if g := cmpFromAPI(context.Background(), obj, &d); g.ConsentAPI != nil {
		t.Error("null consent_api must stay absent")
	}
	obj["consent_api"] = map[string]interface{}{"global": "x", "has_method": "has", "request_method": nil, "category_map": nil}
	if g := cmpFromAPI(context.Background(), obj, &d); !g.ConsentAPI.CategoryMap.IsNull() || !g.ConsentAPI.RequestMethod.IsNull() {
		t.Errorf("null category_map / request_method: %+v", g.ConsentAPI)
	}
}

func integrationPlan(cmp *cmpModel) integrationModel {
	return integrationModel{
		ID: unkS(), Integration: strVal("support-desk"), Name: strVal("Support"), Role: strVal("other"), Audience: strVal("desk:one"),
		AudienceFormat: nullS(), Enabled: types.BoolValue(true), PrincipalClaim: strVal("audience_hint"), ExtClaim: strVal("external_id"),
		TTLSeconds: types.Int64Unknown(), EventsURL: nullS(), BackchannelLogoutURL: nullS(), VerifyErrorReportURL: nullS(),
		OAuthClientID: nullS(), AtlasIDUsers: types.BoolValue(false), ExternalRef: nullS(),
		Metadata: types.MapValueMust(types.StringType, nil), Preset: unkS(), Issuer: unkS(), JWKSURL: unkS(), BackchannelLogoutSource: unkS(),
		Cmp: cmp,
	}
}

func TestIntegration_cmpSentAndReadBack(t *testing.T) {
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodPut || call.Method == http.MethodGet {
			return 200, map[string]interface{}{
				"id": "support-desk", "name": "Support", "role": "other", "enabled": true, "connected": true, "audience": "desk:one",
				"principal_claim": "audience_hint", "ext_claim": "external_id", "ttl_seconds": 120, "issuer": "https://i", "jwks_url": "https://i/jwks",
				"atlas_id_users": false, "metadata": map[string]interface{}{}, "cmp": cmpAPIObject(),
			}
		}
		return 200, map[string]interface{}{}
	})
	r := NewIntegrationResource().(*integrationResource)
	configureResource(t, r, c)
	state, resp := doCreate(t, r, planFrom(t, r, integrationPlan(cmpFixture(t))))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	call := api.last(http.MethodPut, "/v1/instance/integrations/support-desk")
	cmp, _ := call.Body["cmp"].(map[string]interface{})
	ca, _ := cmp["consent_api"].(map[string]interface{})
	if cmp["script_url"] != "https://cmp.example.com/banner.js" || ca["global"] != "__cmp" || ca["request_method"] != "request" {
		t.Errorf("PUT body cmp: %+v", call.Body["cmp"])
	}
	var got integrationModel
	mustOK(t, "get state", state.Get(context.Background(), &got).HasError(), nil)
	if got.Cmp == nil || got.Cmp.ConsentAPI == nil || got.Cmp.ConsentAPI.Global.ValueString() != "__cmp" {
		t.Errorf("state cmp: %+v", got.Cmp)
	}
	rs, rr := doRead(t, r, state)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	var again integrationModel
	rs.Get(context.Background(), &again)
	if again.Cmp == nil || again.Cmp.SiteID.ValueString() != "site-1" {
		t.Errorf("read cmp: %+v", again.Cmp)
	}
}

func TestIntegration_noCmpBlockSendsNoCmp(t *testing.T) {
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 200, map[string]interface{}{
			"id": "support-desk", "name": "Support", "role": "other", "enabled": true, "connected": true, "audience": "desk:one",
			"principal_claim": "audience_hint", "ext_claim": "external_id", "issuer": "https://i", "jwks_url": "https://i/jwks",
			"atlas_id_users": false, "metadata": map[string]interface{}{}, "cmp": nil,
		}
	})
	r := NewIntegrationResource().(*integrationResource)
	configureResource(t, r, c)
	state, resp := doCreate(t, r, planFrom(t, r, integrationPlan(nil)))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	if _, ok := api.last(http.MethodPut, "/v1/instance/integrations/support-desk").Body["cmp"]; ok {
		t.Error("an omitted cmp block must not be sent")
	}
	var got integrationModel
	state.Get(context.Background(), &got)
	if got.Cmp != nil {
		t.Errorf("state cmp must be null: %+v", got.Cmp)
	}
}

// The Cookie Munch preset answers a derived default cmp; it stays out of state
// unless the block is configured.
func TestIntegration_cookiemunchDerivedDefaultStaysOutOfState(t *testing.T) {
	var d diag.Diagnostics
	api := &client.Integration{ID: "cookiemunch", Cmp: cmpAPIObject()}
	if got := integrationCmpState(context.Background(), nil, api, &d); got != nil {
		t.Errorf("derived default leaked into state: %+v", got)
	}
	if got := integrationCmpState(context.Background(), &cmpModel{}, api, &d); got == nil {
		t.Error("a configured block must be read back")
	}
}

func outboundPlan(cmp *cmpModel) outboundPartnerLinkModel {
	var d diag.Diagnostics
	return outboundPartnerLinkModel{
		ID: unkS(), Partner: strVal("consentco"), Name: strVal("ConsentCo"), ClientID: strVal("cid"), ClientSecret: strVal("secret"),
		Scopes: stringSliceToSet(context.Background(), []string{"consent:read"}, &d), Resource: nullS(), BaseURL: strVal("https://example.com"),
		TokenURL: unkS(), AuthMethod: unkS(), Status: unkS(), LastError: unkS(), Cmp: cmp,
	}
}

func outboundAPILink(cmp interface{}, isDefault bool) map[string]interface{} {
	return map[string]interface{}{
		"id": "pcr_1", "partner": "consentco", "name": "ConsentCo", "preset": false, "scopes": []string{"consent:read"},
		"base_url": "https://example.com", "token_url": "https://example.com/v1/oauth/token", "client_id": "cid",
		"auth_method": "client_credentials", "source": "manual", "status": "active", "cmp": cmp, "cmp_default": isDefault,
	}
}

func TestOutboundPartnerLink_cmpSetThenCleared(t *testing.T) {
	var current interface{} = cmpAPIObject()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodPut {
			if v, ok := call.Body["cmp"]; ok {
				current = v
			}
		}
		return 200, outboundAPILink(current, false)
	})
	r := NewOutboundPartnerLinkResource().(*outboundPartnerLinkResource)
	configureResource(t, r, c)
	state, resp := doCreate(t, r, planFrom(t, r, outboundPlan(cmpFixture(t))))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	cmp, _ := api.last(http.MethodPut, "/v1/partner_links/outbound/consentco").Body["cmp"].(map[string]interface{})
	if ca, _ := cmp["consent_api"].(map[string]interface{}); cmp["script_url"] == nil || ca["global"] != "__cmp" {
		t.Errorf("PUT cmp: %+v", cmp)
	}
	var got outboundPartnerLinkModel
	state.Get(context.Background(), &got)
	if got.Cmp == nil || got.Cmp.ConsentAPI == nil {
		t.Fatalf("state cmp: %+v", got.Cmp)
	}

	// Removing the block sends cmp: null (absent would keep the stored one).
	current = nil
	state2, up := doUpdate(t, r, planFrom(t, r, outboundPlan(nil)), state)
	mustOK(t, "update", up.Diagnostics.HasError(), up.Diagnostics)
	calls := 0
	for _, c := range api.all() {
		if c.Method == http.MethodPut {
			calls++
			if calls == 2 {
				if v, ok := c.Body["cmp"]; !ok || v != nil {
					t.Errorf("clearing must send cmp:null, body: %+v", c.Body)
				}
			}
		}
	}
	var after outboundPartnerLinkModel
	state2.Get(context.Background(), &after)
	if after.Cmp != nil {
		t.Errorf("cmp must be gone: %+v", after.Cmp)
	}
}

func TestOutboundPartnerLink_derivedCmpIsNotState(t *testing.T) {
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) { return 200, outboundAPILink(cmpAPIObject(), true) })
	r := NewOutboundPartnerLinkResource().(*outboundPartnerLinkResource)
	configureResource(t, r, c)
	state, resp := doCreate(t, r, planFrom(t, r, outboundPlan(nil)))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	if _, ok := api.last(http.MethodPut, "/v1/partner_links/outbound/consentco").Body["cmp"]; ok {
		t.Error("no cmp block, nothing sent (the API keeps the stored cmp)")
	}
	var got outboundPartnerLinkModel
	state.Get(context.Background(), &got)
	if got.Cmp != nil {
		t.Errorf("derived default leaked into state: %+v", got.Cmp)
	}
}

// The data sources are derived from the resource schemas: cmp becomes a
// read-only nested attribute, and the derived default is included.
func TestCmp_dataSourcesExposeCmp(t *testing.T) {
	ctx := context.Background()
	c, _ := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if strings.Contains(call.Path, "/partner_links/outbound") {
			return 200, map[string]interface{}{"data": []interface{}{outboundAPILink(cmpAPIObject(), true)}}
		}
		return 200, map[string]interface{}{
			"id": "support-desk", "name": "Support", "role": "other", "enabled": true, "connected": true, "audience": "desk:one",
			"principal_claim": "audience_hint", "ext_claim": "external_id", "issuer": "https://i", "jwks_url": "https://i/jwks",
			"atlas_id_users": false, "metadata": map[string]interface{}{}, "cmp": cmpAPIObject(),
		}
	})
	for _, spec := range dataSourceSpecs() {
		if spec.name != "integration" && spec.name != "outbound_partner_link" {
			continue
		}
		id := "support-desk"
		if spec.name == "outbound_partner_link" {
			id = "consentco"
		}
		var d diag.Diagnostics
		st, ok := spec.load(ctx, c, id, &d)
		if !ok || d.HasError() {
			t.Fatalf("%s load: %v", spec.name, d)
		}
		var all map[string]tftypes.Value
		if err := st.Raw.As(&all); err != nil {
			t.Fatal(err)
		}
		if all["cmp"].IsNull() {
			t.Errorf("%s: cmp missing from the data source", spec.name)
		}
		ds := (&specDataSource{spec: spec}).dataSourceSchema()
		if _, ok := ds.Attributes["cmp"]; !ok {
			t.Errorf("%s data source schema has no cmp", spec.name)
		}
	}
}

func TestCmp_presetResourcesAcceptTheBlock(t *testing.T) {
	for _, p := range integrationPresets {
		s := integrationSchema(p.ID)
		if _, ok := s.Blocks["cmp"]; !ok {
			t.Errorf("atlas_%s_integration has no cmp block", p.ID)
		}
	}
	if _, ok := outboundPartnerLinkSchema().Blocks["cmp"]; !ok {
		t.Error("atlas_outbound_partner_link has no cmp block")
	}
}
