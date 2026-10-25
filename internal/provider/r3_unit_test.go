package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const r3SecretToken = "pt_live_do_not_leak_0123456789"

func integrationAPIBody(perCall interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id": "support-desk", "name": "Support", "role": "other", "enabled": true, "connected": true, "audience": "desk:one",
		"principal_claim": "audience_hint", "ext_claim": "external_id", "ttl_seconds": 120, "issuer": "https://i", "jwks_url": "https://i/jwks",
		"atlas_id_users": false, "metadata": map[string]interface{}{}, "cmp": nil, "per_call_audience": perCall,
	}
}

func TestIntegration_perCallAudienceSentAndReadBack(t *testing.T) {
	obj := map[string]interface{}{
		"param": "tenant", "template": "desk:{tenant}", "allowed": []interface{}{"t1", "t2"}, "any_matching": false, "default": false,
	}
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) { return 200, integrationAPIBody(obj) })
	r := NewIntegrationResource().(*integrationResource)
	configureResource(t, r, c)
	ctx := context.Background()
	var d diag.Diagnostics
	plan := integrationPlan(nil)
	plan.PerCallAudience = &perCallAudienceModel{
		Param: strVal("tenant"), Template: strVal("desk:{tenant}"),
		Allowed: stringSliceToSet(ctx, []string{"t1", "t2"}, &d), AnyMatching: types.BoolValue(false),
	}
	state, resp := doCreate(t, r, planFrom(t, r, plan))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	pc, _ := api.last(http.MethodPut, "/v1/instance/integrations/support-desk").Body["per_call_audience"].(map[string]interface{})
	if pc["param"] != "tenant" || pc["template"] != "desk:{tenant}" || pc["any_matching"] != false {
		t.Errorf("PUT body per_call_audience: %+v", pc)
	}
	if allowed, _ := pc["allowed"].([]interface{}); len(allowed) != 2 {
		t.Errorf("allowed: %+v", pc["allowed"])
	}
	var got integrationModel
	mustOK(t, "get state", state.Get(ctx, &got).HasError(), nil)
	if got.PerCallAudience == nil || got.PerCallAudience.Param.ValueString() != "tenant" || len(got.PerCallAudience.Allowed.Elements()) != 2 {
		t.Errorf("state per_call_audience: %+v", got.PerCallAudience)
	}
}

func TestIntegration_perCallAudienceOmittedKeepsDefaultAndDisabledSendsNull(t *testing.T) {
	def := map[string]interface{}{"param": "cbid", "template": "cookiemunch:{cbid}", "allowed": []interface{}{}, "any_matching": false, "default": true}
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) { return 200, integrationAPIBody(def) })
	r := NewIntegrationResource().(*integrationResource)
	configureResource(t, r, c)
	ctx := context.Background()
	state, resp := doCreate(t, r, planFrom(t, r, integrationPlan(nil)))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	if _, ok := api.last(http.MethodPut, "/v1/instance/integrations/support-desk").Body["per_call_audience"]; ok {
		t.Error("an omitted block must not be sent (the preset default stays)")
	}
	var got integrationModel
	state.Get(ctx, &got)
	if got.PerCallAudience != nil {
		t.Errorf("the preset's default must stay out of state: %+v", got.PerCallAudience)
	}

	// Disabled sends an explicit null.
	off := integrationPlan(nil)
	off.PerCallAudienceDisabled = types.BoolValue(true)
	c2, api2 := newFakeAPI(t, func(call apiCall) (int, interface{}) { return 200, integrationAPIBody(nil) })
	r2 := NewIntegrationResource().(*integrationResource)
	configureResource(t, r2, c2)
	state2, resp2 := doCreate(t, r2, planFrom(t, r2, off))
	mustOK(t, "create off", resp2.Diagnostics.HasError(), resp2.Diagnostics)
	body := api2.last(http.MethodPut, "/v1/instance/integrations/support-desk").Body
	if v, ok := body["per_call_audience"]; !ok || v != nil {
		t.Errorf("disabled must send per_call_audience: null, got %v (present %v)", v, ok)
	}
	var got2 integrationModel
	state2.Get(ctx, &got2)
	if got2.PerCallAudience != nil || !got2.PerCallAudienceDisabled.ValueBool() {
		t.Errorf("state: %+v disabled=%v", got2.PerCallAudience, got2.PerCallAudienceDisabled)
	}
}

func TestPerCallAudience_blockAndDisabledConflict(t *testing.T) {
	var d diag.Diagnostics
	validatePerCallAudience(&perCallAudienceModel{}, types.BoolValue(true), &d)
	if !d.HasError() {
		t.Error("a block together with per_call_audience_disabled must be refused")
	}
	d = diag.Diagnostics{}
	validatePerCallAudience(&perCallAudienceModel{}, types.BoolValue(false), &d)
	validatePerCallAudience(nil, types.BoolValue(true), &d)
	if d.HasError() {
		t.Errorf("valid combinations refused: %v", d)
	}
}

func TestPerCallAudience_dataSourceShowsTheDefault(t *testing.T) {
	var d diag.Diagnostics
	def := map[string]interface{}{"param": "cbid", "template": "cookiemunch:{cbid}", "allowed": []interface{}{"a"}, "any_matching": true, "default": true}
	if perCallAudienceFromAPI(context.Background(), nil, def, false, &d) != nil {
		t.Error("resource state hides the default")
	}
	got := perCallAudienceFromAPI(context.Background(), nil, def, true, &d)
	if got == nil || got.Param.ValueString() != "cbid" || !got.AnyMatching.ValueBool() {
		t.Errorf("data source: %+v", got)
	}
}

func orchestratorPlan() privacyOrchestratorModel {
	return privacyOrchestratorModel{
		ID: unkS(), ProviderID: strVal("tfacc"), Name: strVal("Orchestrator"), Enabled: types.BoolValue(true),
		CaseURL: strVal("https://example.com/cases"), CancelURL: nullS(), AuthHeader: nullS(), AuthScheme: nullS(),
		APIKeyWO: nullS(), APIKeyWOVersion: types.Int64Null(), UsePartnerLink: types.BoolValue(false),
		PartnerLinkID: strVal("ppl_123"), SiteID: strVal("dv-site-1"), CancelDeletionOnSignIn: types.BoolValue(true),
		SignRequests: types.BoolValue(false), OpenClosureCases: types.BoolValue(true), CompletionWebhookURL: nullS(),
		APIKeySet: unkB(), SigningSecret: unkS(), CompletionWebhookSecret: unkS(), CompletionWebhookEndpoint: unkS(),
		LastCaseAt: unkI(), LastError: unkS(),
	}
}

func unkB() types.Bool { return types.BoolUnknown() }

func TestPrivacyOrchestrator_siteAndPartnerLinkSent(t *testing.T) {
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 200, map[string]interface{}{
			"configured": true, "enabled": true, "provider": "tfacc", "name": "Orchestrator", "case_url": "https://example.com/cases",
			"auth_header": "authorization", "auth_scheme": "bearer", "api_key_set": false, "signing_secret_set": false,
			"use_partner_link": true, "partner_link_id": "ppl_123", "site_id": "dv-site-1", "cancel_deletion_on_sign_in": true,
			"open_closure_cases": true, "completion_webhook": nil,
		}
	})
	r := NewPrivacyOrchestratorResource().(*privacyOrchestratorResource)
	configureResource(t, r, c)
	state, resp := doCreate(t, r, planFrom(t, r, orchestratorPlan()))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	body := api.last(http.MethodPut, "/v1/instance/privacy_orchestrator").Body
	if body["site_id"] != "dv-site-1" || body["partner_link_id"] != "ppl_123" || body["cancel_deletion_on_sign_in"] != true {
		t.Errorf("PUT body: %+v", body)
	}
	var got privacyOrchestratorModel
	state.Get(context.Background(), &got)
	if got.SiteID.ValueString() != "dv-site-1" || got.PartnerLinkID.ValueString() != "ppl_123" || !got.CancelDeletionOnSignIn.ValueBool() {
		t.Errorf("state: %+v", got)
	}
}

func TestPrivacyOrchestrator_removingSiteIDClearsIt(t *testing.T) {
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 200, map[string]interface{}{
			"configured": true, "enabled": true, "provider": "tfacc", "name": "Orchestrator", "case_url": "https://example.com/cases",
			"auth_header": "authorization", "auth_scheme": "bearer", "api_key_set": false, "signing_secret_set": false,
			"use_partner_link": false, "partner_link_id": nil, "site_id": nil, "cancel_deletion_on_sign_in": false,
			"open_closure_cases": true, "completion_webhook": nil,
		}
	})
	r := NewPrivacyOrchestratorResource().(*privacyOrchestratorResource)
	configureResource(t, r, c)
	var prior privacyOrchestratorModel = orchestratorPlan()
	prior.SigningSecret, prior.CompletionWebhookSecret = nullS(), nullS()
	plan := orchestratorPlan()
	plan.PartnerLinkID, plan.SiteID = nullS(), nullS()
	plan.CancelDeletionOnSignIn = unkB()
	var diags diag.Diagnostics
	r.apply(context.Background(), &plan, &plan, &prior, &diags)
	mustOK(t, "apply", diags.HasError(), diags)
	body := api.last(http.MethodPut, "/v1/instance/privacy_orchestrator").Body
	for _, k := range []string{"site_id", "partner_link_id"} {
		if v, ok := body[k]; !ok || v != nil {
			t.Errorf("%s must be sent as null to clear it, got %v (present %v)", k, v, ok)
		}
	}
	if _, ok := body["cancel_deletion_on_sign_in"]; ok {
		t.Error("an unset cancel_deletion_on_sign_in must not be sent")
	}
}

func partnerPlatformAPI(tokenSet bool, prefix interface{}) map[string]interface{} {
	return map[string]interface{}{
		"object": "partner_platform", "id": "ppl_abc", "partner": "consentco", "name": "ConsentCo", "preset": false, "platform_id": "atlas",
		"enabled": true, "jwks_url": "https://consentco.example/.well-known/jwks.json", "signature_header": "x-consentco-signature",
		"signature_scheme": "t_kid_ed25519", "tolerance_seconds": 300, "allowed_sources": []string{"site-1"}, "record_consents": true,
		"receiver_url": "https://api.example/v1/partner_platforms/consentco/receive", "receiver_url_template": "",
		"registration": map[string]interface{}{"status": "unregistered", "registered_at": nil},
		"token_set":    tokenSet, "token_prefix": prefix, "last_delivery_at": nil, "last_delivery_error": nil,
	}
}

func partnerPlatformPlan(ctx context.Context, token types.String, version types.Int64) partnerPlatformModel {
	var d diag.Diagnostics
	return partnerPlatformModel{
		ID: unkS(), Partner: strVal("consentco"), Name: strVal("ConsentCo"), PlatformID: strVal("atlas"), Enabled: types.BoolValue(true),
		JWKSURL: strVal("https://consentco.example/.well-known/jwks.json"), SignatureHeader: strVal("x-consentco-signature"),
		SignatureScheme: unkS(), ToleranceSeconds: unkI(), AllowedSources: stringSliceToSet(ctx, []string{"site-1"}, &d),
		RecordConsents: types.BoolValue(true), TokenWO: token, TokenWOVersion: version, APIBaseURL: strVal("https://api.consentco.example"),
		Preset: unkB(), ReceiverURL: unkS(), ReceiverURLTemplate: unkS(), RegistrationStatus: unkS(), RegisteredAt: unkS(),
		TokenSet: unkB(), TokenPrefix: unkS(), LastDeliveryAt: unkS(), LastDeliveryError: unkS(),
	}
}

func TestPartnerPlatform_createStoresTokenOnceAndNeverKeepsIt(t *testing.T) {
	stored := false
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodPost && strings.HasSuffix(call.Path, "/token") {
			stored = true
		}
		if stored {
			return 200, partnerPlatformAPI(true, "pt_live_")
		}
		return 200, partnerPlatformAPI(false, nil)
	})
	r := NewPartnerPlatformResource().(*partnerPlatformResource)
	configureResource(t, r, c)
	ctx := context.Background()
	plan := partnerPlatformPlan(ctx, strVal(r3SecretToken), types.Int64Value(1))
	state, resp := doCreate(t, r, planFrom(t, r, plan))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)

	put := api.last(http.MethodPut, "/v1/partner_platforms/consentco")
	if put == nil {
		t.Fatal("no PUT")
	}
	if _, ok := put.Body["token"]; ok {
		t.Error("the token must not travel in the PUT")
	}
	if put.Body["platform_id"] != "atlas" || put.Body["record_consents"] != true {
		t.Errorf("PUT body: %+v", put.Body)
	}
	post := api.last(http.MethodPost, "/v1/partner_platforms/consentco/token")
	if post == nil || post.Body["token"] != r3SecretToken || post.Body["rotate"] != false || post.Body["api_base_url"] != "https://api.consentco.example" {
		t.Fatalf("token POST: %+v", post)
	}

	var got partnerPlatformModel
	mustOK(t, "get state", state.Get(ctx, &got).HasError(), nil)
	if !got.TokenSet.ValueBool() || got.TokenPrefix.ValueString() != "pt_live_" || got.RegistrationStatus.ValueString() != "unregistered" {
		t.Errorf("state: %+v", got)
	}
	if !got.TokenWO.IsNull() {
		t.Error("token_wo must be null in state")
	}
	raw, _ := json.Marshal(state.Raw.String())
	if strings.Contains(string(raw), r3SecretToken) || strings.Contains(state.Raw.String(), r3SecretToken) {
		t.Error("the token is in state")
	}
}

func TestPartnerPlatform_tokenVersionBumpRotatesAndNeedsToken(t *testing.T) {
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) { return 200, partnerPlatformAPI(true, "pt_live_") })
	r := NewPartnerPlatformResource().(*partnerPlatformResource)
	configureResource(t, r, c)
	ctx := context.Background()
	var diags diag.Diagnostics
	prior := partnerPlatformPlan(ctx, nullS(), types.Int64Value(1))
	fillPartnerPlatform(ctx, &prior, mustPlatform(t, partnerPlatformAPI(true, "pt_live_")), &diags)

	// Same version: no token call.
	same := partnerPlatformPlan(ctx, strVal(r3SecretToken), types.Int64Value(1))
	r.apply(ctx, &same, &same, &prior, &diags)
	mustOK(t, "apply same", diags.HasError(), diags)
	if api.count(http.MethodPost, "/v1/partner_platforms/consentco/token") != 0 {
		t.Error("an unchanged token_wo_version must not send the token")
	}

	// New version: rotate=true.
	bumped := partnerPlatformPlan(ctx, strVal(r3SecretToken), types.Int64Value(2))
	r.apply(ctx, &bumped, &bumped, &prior, &diags)
	mustOK(t, "apply bump", diags.HasError(), diags)
	post := api.last(http.MethodPost, "/v1/partner_platforms/consentco/token")
	if post == nil || post.Body["rotate"] != true {
		t.Errorf("a later send must rotate: %+v", post)
	}

	// New version without a token: refused.
	missing := partnerPlatformPlan(ctx, nullS(), types.Int64Value(3))
	var d2 diag.Diagnostics
	r.apply(ctx, &missing, &missing, &prior, &d2)
	if !d2.HasError() {
		t.Error("a bumped version without token_wo must be refused")
	}
}

func mustPlatform(t *testing.T, body map[string]interface{}) *client.PartnerPlatform {
	t.Helper()
	raw, _ := json.Marshal(body)
	var p client.PartnerPlatform
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	return &p
}

func TestPartnerPlatform_modelNeverCarriesAToken(t *testing.T) {
	// The view type has no token field at all: a response that did carry one
	// would be dropped, not stored.
	body := partnerPlatformAPI(true, "pt_live_")
	body["token"] = r3SecretToken
	p := mustPlatform(t, body)
	var m partnerPlatformModel
	var d diag.Diagnostics
	fillPartnerPlatform(context.Background(), &m, p, &d)
	if strings.Contains(m.TokenPrefix.String(), r3SecretToken) || !m.TokenWO.IsNull() && m.TokenWO.ValueString() != "" {
		t.Errorf("token leaked into the model: %+v", m)
	}
}

func TestPartnerPlatform_readDropsAMissingPlatform(t *testing.T) {
	c, _ := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 404, map[string]interface{}{"errors": []map[string]interface{}{{"code": "NOT_FOUND", "message": "gone"}}}
	})
	r := NewPartnerPlatformResource().(*partnerPlatformResource)
	configureResource(t, r, c)
	ctx := context.Background()
	m := partnerPlatformPlan(ctx, nullS(), types.Int64Null())
	var d diag.Diagnostics
	fillPartnerPlatform(ctx, &m, mustPlatform(t, partnerPlatformAPI(false, nil)), &d)
	_, rr := doRead(t, r, stateFrom(t, r, m))
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	if !rr.State.Raw.IsNull() {
		t.Error("a missing platform must leave state")
	}
}
