package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPolicyWriteMarshalsNullGraceDays(t *testing.T) {
	decode := func(w PolicyWrite) map[string]any {
		raw, err := json.Marshal(w)
		if err != nil {
			t.Fatal(err)
		}
		var out map[string]any
		_ = json.Unmarshal(raw, &out)
		return out
	}
	days := int64(14)
	set := decode(PolicyWrite{Key: "tos", Title: "Terms", Display: "scroll", Required: true, AppliesTo: "sign_in_and_sign_up", Reaccept: "grace", GraceDays: &days})
	if set["grace_days"] != float64(14) || set["key"] != "tos" {
		t.Errorf("set = %v", set)
	}
	if _, sent := set["status"]; sent {
		t.Error("status must be omitted unless set")
	}
	cleared := decode(PolicyWrite{Title: "Terms", Display: "checkbox", AppliesTo: "sign_up", Reaccept: "none", Status: "active"})
	if v, sent := cleared["grace_days"]; !sent || v != nil {
		t.Errorf("a nil GraceDays must be sent as null: %v", cleared)
	}
	if _, sent := cleared["key"]; sent {
		t.Error("key must be omitted on update")
	}
	if cleared["status"] != "active" {
		t.Errorf("cleared = %v", cleared)
	}
}

func TestPolicySourceWriteSecretIsTriState(t *testing.T) {
	secret := "whsec_0123456789abcdef"
	poll := int64(30)
	decode := func(w PolicySourceWrite) map[string]any {
		raw, _ := json.Marshal(w)
		var out map[string]any
		_ = json.Unmarshal(raw, &out)
		return out
	}
	kept := decode(PolicySourceWrite{Enabled: true, EndpointURL: "https://cms.example.com/atlas", Sync: "webhook"})
	if _, sent := kept["signing_secret"]; sent {
		t.Error("SendSecret false must omit the secret")
	}
	if _, sent := kept["poll_interval_minutes"]; sent {
		t.Error("a nil poll interval must be omitted")
	}
	set := decode(PolicySourceWrite{Sync: "poll", PollIntervalMinutes: &poll, SendSecret: true, Secret: &secret})
	if set["signing_secret"] != secret || set["poll_interval_minutes"] != float64(30) {
		t.Errorf("set = %v", set)
	}
	cleared := decode(PolicySourceWrite{SendSecret: true})
	if v, sent := cleared["signing_secret"]; !sent || v != nil {
		t.Errorf("SendSecret with a nil secret must send null: %v", cleared)
	}
}

func TestPolicyRoutes(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		calls = append(calls, r.Method+" "+r.URL.EscapedPath()+" "+string(body))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "pol_1", "key": "tos", "title": "Terms", "display": "checkbox", "required": true,
			"applies_to": "sign_up", "reaccept": "none", "grace_days": nil, "status": "active",
			"session_claim": true, "enabled": true, "endpoint_url": "https://cms.example.com/atlas",
			"signing_secret_set": true, "sync": "webhook", "poll_interval_minutes": 60,
		})
	}))
	defer srv.Close()
	c := newTestClient(srv)
	ctx := context.Background()
	if p, err := c.GetPolicy(ctx, "a/b"); err != nil || p.ID != "pol_1" || p.GraceDays != nil {
		t.Fatalf("get = %+v %v", p, err)
	}
	if _, err := c.CreatePolicy(ctx, PolicyWrite{Key: "tos", Title: "Terms"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UpdatePolicy(ctx, "pol_1", PolicyWrite{Title: "Terms"}); err != nil {
		t.Fatal(err)
	}
	if err := c.DeletePolicy(ctx, "pol_1"); err != nil {
		t.Fatal(err)
	}
	if s, err := c.GetPolicySettings(ctx); err != nil || !s.SessionClaim {
		t.Fatalf("settings = %+v %v", s, err)
	}
	if _, err := c.SetPolicySettings(ctx, false); err != nil {
		t.Fatal(err)
	}
	if s, err := c.GetPolicySource(ctx); err != nil || !s.SigningSecretSet || s.PollIntervalMinutes != 60 {
		t.Fatalf("source = %+v %v", s, err)
	}
	if _, err := c.PutPolicySource(ctx, PolicySourceWrite{EndpointURL: "https://cms.example.com/atlas", Sync: "webhook"}); err != nil {
		t.Fatal(err)
	}
	if err := c.DeletePolicySource(ctx); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /v1/policies/a%2Fb ",
		`POST /v1/policies {"applies_to":"","display":"","grace_days":null,"key":"tos","reaccept":"","required":false,"title":"Terms"}`,
		`PATCH /v1/policies/pol_1 {"applies_to":"","display":"","grace_days":null,"reaccept":"","required":false,"title":"Terms"}`,
		"DELETE /v1/policies/pol_1 ",
		"GET /v1/instance/policy_settings ",
		`PATCH /v1/instance/policy_settings {"session_claim":false}`,
		"GET /v1/instance/policy_source ",
		`PUT /v1/instance/policy_source {"enabled":false,"endpoint_url":"https://cms.example.com/atlas","sync":"webhook"}`,
		"DELETE /v1/instance/policy_source ",
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

func TestIsConflict(t *testing.T) {
	if !IsConflict(&APIError{Status: 409}) || !IsConflict(&APIError{Status: 422, Errors: []ErrorItem{{Code: "CONFLICT"}}}) {
		t.Error("409 and CONFLICT are conflicts")
	}
	if IsConflict(&APIError{Status: 404}) || IsConflict(nil) {
		t.Error("404 and nil are not conflicts")
	}
}
