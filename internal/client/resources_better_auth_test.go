package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBetterAuthBridgeWriteSecretIsTriState(t *testing.T) {
	url := "https://app.example.com/api/auth"
	secret := "0123456789abcdef"
	decode := func(w BetterAuthBridgeWrite) map[string]any {
		raw, err := json.Marshal(w)
		if err != nil {
			t.Fatal(err)
		}
		var out map[string]any
		_ = json.Unmarshal(raw, &out)
		return out
	}
	kept := decode(BetterAuthBridgeWrite{Enabled: true, CookiePrefix: "ba", LiveFallback: BetterAuthBridgeLiveFallbackWrite{Enabled: true, BaseURL: &url}})
	live := kept["live_fallback"].(map[string]any)
	if _, sent := live["secret"]; sent {
		t.Error("SendSecret false must omit the secret (the stored one is kept)")
	}
	if v, sent := kept["ends_at"]; !sent || v != nil {
		t.Errorf("a nil EndsAt must be sent as null: %v", kept)
	}
	set := decode(BetterAuthBridgeWrite{LiveFallback: BetterAuthBridgeLiveFallbackWrite{SendSecret: true, Secret: &secret}})
	if set["live_fallback"].(map[string]any)["secret"] != secret {
		t.Errorf("set = %v", set)
	}
	cleared := decode(BetterAuthBridgeWrite{LiveFallback: BetterAuthBridgeLiveFallbackWrite{SendSecret: true}})
	if v, sent := cleared["live_fallback"].(map[string]any)["secret"]; !sent || v != nil {
		t.Errorf("SendSecret with a nil secret must send null: %v", cleared)
	}
}

func TestBetterAuthBridgeRoutes(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		calls = append(calls, r.Method+" "+r.URL.Path+" "+string(body))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "better_auth_bridge", "enabled": true, "ends_at": 1798675200000, "cookie_prefix": "ba",
			"live_fallback": map[string]any{"enabled": true, "base_url": "https://app.example.com/api/auth", "has_secret": true},
			"exchange_path": "/v1/client/migrations/better_auth/exchange", "created_at": 1, "updated_at": 2,
		})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	got, err := c.GetBetterAuthBridge(context.Background())
	if err != nil || !got.Enabled || got.EndsAt == nil || *got.EndsAt != 1798675200000 || !got.LiveFallback.HasSecret || got.CreatedAt == nil {
		t.Fatalf("get = %+v %v", got, err)
	}
	if _, err := c.PutBetterAuthBridge(context.Background(), BetterAuthBridgeWrite{Enabled: true, CookiePrefix: "ba"}); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteBetterAuthBridge(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /v1/migrations/better_auth/bridge ",
		`PUT /v1/migrations/better_auth/bridge {"cookie_prefix":"ba","enabled":true,"ends_at":null,"live_fallback":{"base_url":null,"enabled":false}}`,
		"DELETE /v1/migrations/better_auth/bridge ",
	}
	if len(calls) != len(want) {
		t.Fatalf("calls = %q", calls)
	}
	for i, w := range want {
		if calls[i] != w {
			t.Errorf("call %d = %q, want %q", i, calls[i], w)
		}
	}
}
