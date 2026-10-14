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
