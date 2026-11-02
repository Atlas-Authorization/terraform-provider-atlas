package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIgaCrudPaths(t *testing.T) {
	var calls []string
	var lastBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		lastBody = nil
		_ = json.NewDecoder(r.Body).Decode(&lastBody)
		switch r.URL.Path {
		case "/v1/iga/sod_rules", "/v1/iga/birthright_rules":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": []map[string]any{{"id": "igs_1", "name": "n", "entitlement_a": "a", "entitlement_b": "b", "mode": "block", "enabled": true, "match": map[string]any{"group": "g"}, "entitlement_ids": []string{"ige_1"}, "revoke_on_unmatch": true}}})
				return
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "ige_1", "key": "k", "name": "n", "kind": "custom", "target": map[string]any{}, "risk_level": "low", "approval_rules": map[string]any{"steps": []any{}}, "requestable": true, "require_justification": true})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	ctx := context.Background()

	e, err := c.CreateIgaEntitlement(ctx, IgaEntitlementWrite{Key: "k", Kind: "custom", Name: "n", Target: map[string]interface{}{}, ApprovalRules: map[string]interface{}{"steps": []interface{}{}}})
	if err != nil || e.ID != "ige_1" {
		t.Fatalf("create = %+v, %v", e, err)
	}
	if _, err := c.UpdateIgaEntitlement(ctx, "ige_1", IgaEntitlementWrite{Key: "k", Kind: "custom", Name: "n2"}); err != nil {
		t.Fatal(err)
	}
	// key and kind are create-only: an update must not send them.
	if _, ok := lastBody["key"]; ok {
		t.Errorf("update body carries key: %v", lastBody)
	}
	if _, ok := lastBody["kind"]; ok {
		t.Errorf("update body carries kind: %v", lastBody)
	}
	if err := c.DeleteIgaEntitlement(ctx, "ige_1"); err != nil {
		t.Fatal(err)
	}
	rule, err := c.GetIgaSodRule(ctx, "igs_1")
	if err != nil || rule.Mode != "block" {
		t.Fatalf("sod get = %+v, %v", rule, err)
	}
	if _, err := c.GetIgaSodRule(ctx, "nope"); !IsNotFound(err) {
		t.Errorf("unknown sod rule should be not found, got %v", err)
	}
	if _, err := c.GetIgaBirthrightRule(ctx, "nope"); !IsNotFound(err) {
		t.Errorf("unknown birthright rule should be not found, got %v", err)
	}
	want := []string{"POST /v1/iga/entitlements", "PATCH /v1/iga/entitlements/ige_1", "DELETE /v1/iga/entitlements/ige_1", "GET /v1/iga/sod_rules"}
	for i, w := range want {
		if calls[i] != w {
			t.Errorf("call %d = %s, want %s", i, calls[i], w)
		}
	}
}
