package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetWebhookEndpointKeepsInactive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": []map[string]any{
			{"id": "wh_off", "url": "https://a", "active": false, "failing_since": 5, "created_at": 1},
		}})
	}))
	defer srv.Close()
	got, err := newTestClient(srv).GetWebhookEndpoint(context.Background(), "wh_off")
	if err != nil {
		t.Fatalf("an auto-disabled endpoint still exists: %v", err)
	}
	if got.Active || got.FailingSince == nil || *got.FailingSince != 5 {
		t.Errorf("decoded = %+v", got)
	}
}

func TestRevealAndRotateWebhookSecret(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "webhook_secret", "secret": "whsec_" + r.URL.Path[len(r.URL.Path)-6:]})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	if s, err := c.RevealWebhookSecret(context.Background(), "wh_1"); err != nil || s == "" {
		t.Fatalf("reveal: %q %v", s, err)
	}
	if _, err := c.RotateWebhookSecret(context.Background(), "wh_1"); err != nil {
		t.Fatal(err)
	}
	if paths[0] != "POST /v1/webhook_endpoints/wh_1/reveal_secret" || paths[1] != "POST /v1/webhook_endpoints/wh_1/rotate_secret" {
		t.Errorf("paths = %v", paths)
	}
}

func TestPlatformCallsUsePlatformKey(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "oac_1", "client_id": "client_1", "type": "public"})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	if _, err := c.GetAtlasIDApp(context.Background(), "oac_1"); err == nil {
		t.Fatal("a platform call without a platform key must fail before the request")
	}
	c.PlatformKey = "psk_x"
	if _, err := c.GetAtlasIDApp(context.Background(), "oac_1"); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer psk_x" {
		t.Errorf("auth = %q, want the platform key", auth)
	}
	// BAPI calls without a secret key fail with a clear error, too.
	c.SecretKey = ""
	if _, err := c.GetPreflight(context.Background()); err == nil {
		t.Error("a BAPI call without a secret key must fail")
	}
}

func TestMessagingProviderChannelAbsentIsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "messaging_providers", "email": map[string]any{
			"channel": "email", "transport": "resend", "status": "active", "config": map[string]string{},
			"secrets_set": map[string]bool{"api_key": true}, "updated_at": 1,
		}, "sms": nil})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	if p, err := c.GetMessagingProvider(context.Background(), "email"); err != nil || p.Transport != "resend" || !p.SecretsSet["api_key"] {
		t.Fatalf("email = %+v %v", p, err)
	}
	if _, err := c.GetMessagingProvider(context.Background(), "sms"); !IsNotFound(err) {
		t.Errorf("an unconfigured channel must be NOT_FOUND, got %v", err)
	}
}

func TestDeleteMessagingProviderForceAndInUse(t *testing.T) {
	var uris []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uris = append(uris, r.Method+" "+r.URL.RequestURI())
		if r.URL.Query().Get("force") != "true" {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"errors": []map[string]any{{
				"code": "EMAIL_PROVIDER_IN_USE", "message": "in use",
				"meta": map[string]any{"features": []string{"Password reset"}},
			}}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "messaging_provider", "channel": "email", "deleted": true})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	err := c.DeleteMessagingProvider(context.Background(), "email", false)
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Status != http.StatusConflict || !apiErr.HasCode("EMAIL_PROVIDER_IN_USE") {
		t.Fatalf("want a 409 EMAIL_PROVIDER_IN_USE, got %v", err)
	}
	if features, _ := apiErr.Errors[0].Meta["features"].([]interface{}); len(features) != 1 {
		t.Errorf("meta.features not decoded: %v", apiErr.Errors[0].Meta)
	}
	if err := c.DeleteMessagingProvider(context.Background(), "email", true); err != nil {
		t.Fatal(err)
	}
	if uris[0] != "DELETE /v1/messaging_providers/email" || uris[1] != "DELETE /v1/messaging_providers/email?force=true" {
		t.Errorf("uris = %v", uris)
	}
}

func TestCookieConsentSettingsLinkRoundTrip(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "cookie_consent", "enabled": true, "script_url": "https://consent.example/uc.js", "extra_origins": []string{},
			"settings_link": map[string]any{
				"label": "Cookie settings", "preset": "cookiemunch", "attributes": map[string]string{"id": "cs"},
				"rendered_attributes": map[string]string{"id": "cs", "data-cc": "show-settings"},
			},
		})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	preset := "cookiemunch"
	got, err := c.PutCookieConsent(context.Background(), CookieConsentWrite{
		ScriptURL:    "https://consent.example/uc.js",
		SettingsLink: &CookieSettingsLinkWrite{Label: "Cookie settings", Preset: &preset, Attributes: map[string]string{"id": "cs"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	link, _ := body["settings_link"].(map[string]any)
	if link["label"] != "Cookie settings" || link["preset"] != "cookiemunch" {
		t.Errorf("settings_link sent = %v", body["settings_link"])
	}
	if got.SettingsLink == nil || got.SettingsLink.RenderedAttributes["data-cc"] != "show-settings" {
		t.Errorf("decoded = %+v", got.SettingsLink)
	}

	// No link configured → settings_link: null (the PUT replaces the whole setting).
	if _, err := c.PutCookieConsent(context.Background(), CookieConsentWrite{ScriptURL: "https://consent.example/uc.js"}); err != nil {
		t.Fatal(err)
	}
	if v, present := body["settings_link"]; !present || v != nil {
		t.Errorf("settings_link must be sent as null, got %v (present=%v)", v, present)
	}
}

func TestOAuthClientSkipLogoutConfirmation(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "oac_1", "first_party": true, "skip_logout_confirmation": true})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	got, err := c.UpdateOAuthClient(context.Background(), "oac_1", OAuthClientWrite{SkipLogoutConfirmation: Ptr(false)})
	if err != nil {
		t.Fatal(err)
	}
	if v, present := body["skip_logout_confirmation"]; !present || v != false {
		t.Errorf("an explicit false must be sent, got %v", body)
	}
	if !got.SkipLogoutConfirmation {
		t.Error("skip_logout_confirmation not decoded")
	}
}

func TestOAuthClientDelegatedUI(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "oac_1", "login_url": "https://a.example/signin", "consent_url": nil})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	got, err := c.UpdateOAuthClient(context.Background(), "oac_1", OAuthClientWrite{LoginURL: Ptr("https://a.example/signin"), ConsentURL: Ptr("")})
	if err != nil {
		t.Fatal(err)
	}
	if body["login_url"] != "https://a.example/signin" || body["consent_url"] != "" {
		t.Errorf("urls not sent (an empty string clears): %v", body)
	}
	if got.LoginURL == nil || *got.LoginURL != "https://a.example/signin" || got.ConsentURL != nil {
		t.Errorf("urls not decoded: %+v", got)
	}
}

func TestVcCredentialTypeRoundTrip(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "vct_1", "vct": "org.iso.18013.5.1.mDL", "format": "mso_mdoc", "namespace": "org.iso.18013.5.1", "name": "mDL",
			"lifetime_seconds": 3600, "claims": []map[string]any{{"name": "birth_date", "source": "public_metadata.dob", "disclosable": true, "value_type": "full-date"}},
		})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	got, err := c.CreateVcCredentialType(context.Background(), VcCredentialTypeWrite{
		Vct: "org.iso.18013.5.1.mDL", Name: "mDL", Format: "mso_mdoc",
		Claims: []VcClaim{{Name: "birth_date", Source: "public_metadata.dob", ValueType: "full-date"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["format"] != "mso_mdoc" || body["vct"] != "org.iso.18013.5.1.mDL" {
		t.Errorf("payload: %v", body)
	}
	if got.ID != "vct_1" || got.Claims[0].ValueType != "full-date" || *got.Namespace != "org.iso.18013.5.1" {
		t.Errorf("decode: %+v", got)
	}
}

func TestMcpGatewayCrudPaths(t *testing.T) {
	var calls []string
	var lastBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		lastBody = nil
		_ = json.NewDecoder(r.Body).Decode(&lastBody)
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "mcp_gateway", "id": "mgw_1", "name": "g", "upstream_url": "https://u/", "has_upstream_authorization": true, "default_policy": "deny", "tools": map[string]any{"t": map[string]any{"approval": true}}, "enabled": true, "url": "https://fapi/v1/mcp_gateways/mgw_1/mcp"})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	ctx := context.Background()
	g, err := c.CreateMcpGateway(ctx, McpGatewayWrite{Name: "g", UpstreamURL: "https://u", Tools: map[string]interface{}{}})
	if err != nil || g.ID != "mgw_1" || !g.HasUpstreamAuthorization || string(g.Tools) == "" {
		t.Fatalf("create = %+v, %v", g, err)
	}
	// A null upstream_authorization is sent so an update can clear it.
	if v, ok := lastBody["upstream_authorization"]; !ok || v != nil {
		t.Errorf("upstream_authorization in body = %v (present %v)", v, ok)
	}
	if _, err := c.UpdateMcpGateway(ctx, "mgw_1", McpGatewayWrite{Name: "g", UpstreamURL: "https://u", Tools: map[string]interface{}{}}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetMcpGateway(ctx, "mgw_1"); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteMcpGateway(ctx, "mgw_1"); err != nil {
		t.Fatal(err)
	}
	want := []string{"POST /v1/mcp_gateways", "PATCH /v1/mcp_gateways/mgw_1", "GET /v1/mcp_gateways/mgw_1", "DELETE /v1/mcp_gateways/mgw_1"}
	for i, w := range want {
		if calls[i] != w {
			t.Errorf("call %d = %s, want %s", i, calls[i], w)
		}
	}
}
