package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient points a Client at an httptest server so no network is touched.
func newTestClient(srv *httptest.Server) *Client {
	c := New(srv.URL, "sk_test_key")
	c.HTTP = srv.Client()
	return c
}

func TestRequestSendsBearerAndDecodes(t *testing.T) {
	var gotAuth, gotAccept, gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                "rs_123",
			"identifier":        "https://api.example.com",
			"name":              "Example API",
			"token_ttl_seconds": 3600,
			"signing_alg":       "RS256",
			"created_at":        1700000000000,
			"updated_at":        1700000000000,
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	rs, err := c.GetResourceServer(context.Background(), "rs_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer sk_test_key" {
		t.Errorf("Authorization = %q, want Bearer sk_test_key", gotAuth)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q, want application/json", gotAccept)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/v1/resource_servers/rs_123" {
		t.Errorf("path = %q, want /v1/resource_servers/rs_123", gotPath)
	}
	if rs.ID != "rs_123" || rs.Identifier != "https://api.example.com" || rs.TokenTTLSeconds != 3600 {
		t.Errorf("decoded resource server mismatch: %+v", rs)
	}
}

func TestRequestSendsJSONBody(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "wh_1", "url": "https://x.example.com/hook", "active": true, "created_at": 1})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.CreateWebhookEndpoint(context.Background(), "https://x.example.com/hook", []string{"user.created"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["url"] != "https://x.example.com/hook" {
		t.Errorf("body url = %v", received["url"])
	}
	events, ok := received["enabled_events"].([]any)
	if !ok || len(events) != 1 || events[0] != "user.created" {
		t.Errorf("body enabled_events = %v", received["enabled_events"])
	}
}

func TestErrorEnvelopeMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"code":"VALIDATION_FAILED","message":"A name is required.","param":"name"}]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.GetResourceServer(context.Background(), "rs_1")
	if err == nil {
		t.Fatal("expected an error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", apiErr.Status)
	}
	if !apiErr.HasCode("VALIDATION_FAILED") {
		t.Errorf("HasCode(VALIDATION_FAILED) = false; errors = %+v", apiErr.Errors)
	}
	if apiErr.Errors[0].Param != "name" {
		t.Errorf("param = %q, want name", apiErr.Errors[0].Param)
	}
}

func TestNotFoundMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"code":"NOT_FOUND","message":"Unknown resource server."}]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.GetResourceServer(context.Background(), "missing")
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound = false for err %v", err)
	}
}

func TestNonJSONErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream proxy exploded"))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.GetOAuthClient(context.Background(), "x")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Status != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", apiErr.Status)
	}
	if apiErr.Errors[0].Message != "upstream proxy exploded" {
		t.Errorf("message = %q", apiErr.Errors[0].Message)
	}
}

func TestGetWebhookEndpointFiltersList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "wh_a", "url": "https://a", "active": true, "created_at": 1},
				{"id": "wh_b", "url": "https://b", "active": true, "created_at": 2},
			},
			"has_more": false,
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	got, err := c.GetWebhookEndpoint(context.Background(), "wh_b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.URL != "https://b" {
		t.Errorf("url = %q, want https://b", got.URL)
	}

	if _, err := c.GetWebhookEndpoint(context.Background(), "wh_missing"); !IsNotFound(err) {
		t.Errorf("expected NotFound for missing endpoint, got %v", err)
	}
}

func TestGetOAuthProviderDecodes(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":         "oauth_provider",
			"provider":       "google",
			"display_name":   "Google",
			"category":       "consumer",
			"tier":           "popular",
			"redirect_uri":   "https://acme.atlas.dev/v1/oauth_callbacks/google",
			"default_scopes": []string{"openid", "email", "profile"},
			"configured":     true,
			"client_id":      "gid-123",
			"has_secret":     true,
			"enabled":        true,
			"allow_sign_in":  true,
			"allow_sign_up":  false,
			"scopes":         []string{"openid", "email"},
			"updated_at":     1700000000000,
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	p, err := c.GetOAuthProvider(context.Background(), "google")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/v1/oauth_providers/google" {
		t.Errorf("request = %s %s, want GET /v1/oauth_providers/google", gotMethod, gotPath)
	}
	if p.Provider != "google" || !p.Configured || !p.HasSecret || p.ClientID == nil || *p.ClientID != "gid-123" {
		t.Errorf("decoded provider mismatch: %+v", p)
	}
	if !p.Enabled || !p.AllowSignIn || p.AllowSignUp {
		t.Errorf("scope/enabled flags mismatch: %+v", p)
	}
	if len(p.Scopes) != 2 || p.Scopes[0] != "openid" {
		t.Errorf("scopes = %v", p.Scopes)
	}
	if p.UpdatedAt == nil || *p.UpdatedAt != 1700000000000 {
		t.Errorf("updated_at = %v", p.UpdatedAt)
	}
}

func TestPutOAuthProviderSendsValues(t *testing.T) {
	var received map[string]any
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "oauth_provider", "provider": "github", "configured": true})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	err := c.PutOAuthProvider(context.Background(), "github", OAuthProviderWrite{
		Values:   map[string]string{"client_id": "cid", "client_secret": "shh"},
		Settings: map[string]string{"team": "acme"},
		Scopes:   []string{"read:user"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut || gotPath != "/v1/oauth_providers/github" {
		t.Errorf("request = %s %s, want PUT /v1/oauth_providers/github", gotMethod, gotPath)
	}
	values, ok := received["values"].(map[string]any)
	if !ok || values["client_id"] != "cid" || values["client_secret"] != "shh" {
		t.Errorf("values = %v", received["values"])
	}
	settings, ok := received["settings"].(map[string]any)
	if !ok || settings["team"] != "acme" {
		t.Errorf("settings = %v", received["settings"])
	}
	scopes, ok := received["scopes"].([]any)
	if !ok || len(scopes) != 1 || scopes[0] != "read:user" {
		t.Errorf("scopes = %v", received["scopes"])
	}
}

func TestSetOAuthProviderEnabledAndScope(t *testing.T) {
	var enabledBody, scopeBody map[string]any
	var enabledPath, scopePath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/oauth_providers/google/enabled":
			enabledPath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&enabledBody)
		case r.URL.Path == "/v1/oauth_providers/google/scope":
			scopePath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&scopeBody)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"object":"oauth_provider","provider":"google"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	if err := c.SetOAuthProviderEnabled(context.Background(), "google", true); err != nil {
		t.Fatalf("enabled: %v", err)
	}
	if err := c.SetOAuthProviderScope(context.Background(), "google", true, false); err != nil {
		t.Fatalf("scope: %v", err)
	}
	if enabledPath == "" || enabledBody["enabled"] != true {
		t.Errorf("enabled request = %s %v", enabledPath, enabledBody)
	}
	if scopePath == "" || scopeBody["allow_sign_in"] != true || scopeBody["allow_sign_up"] != false {
		t.Errorf("scope request = %s %v", scopePath, scopeBody)
	}
}

func TestDeleteOAuthProvider(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"object":"oauth_provider","provider":"google","deleted":true}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	if err := c.DeleteOAuthProvider(context.Background(), "google"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/v1/oauth_providers/google" {
		t.Errorf("request = %s %s, want DELETE /v1/oauth_providers/google", gotMethod, gotPath)
	}
}
