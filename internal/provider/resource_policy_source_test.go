package provider

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fakePolicySource is an in-memory policy source with the API's merge
// semantics: a PUT merges onto the stored source and the secret is write-only.
type fakePolicySource struct {
	mu         sync.Mutex
	configured bool
	enabled    bool
	endpoint   interface{}
	sync       string
	poll       int64
	secret     string
}

func (f *fakePolicySource) view() map[string]interface{} {
	syncMode := f.sync
	if syncMode == "" {
		syncMode = "webhook"
	}
	poll := f.poll
	if poll == 0 {
		poll = 60
	}
	return map[string]interface{}{
		"object": "policy_source", "enabled": f.enabled, "endpoint_url": f.endpoint, "signing_secret_set": f.secret != "",
		"sync": syncMode, "poll_interval_minutes": poll, "webhook_url": "https://api.example.com/v1/webhooks/policies/ins_1",
		"protocol_version": "2026-10-24", "last_synced_at": nil, "last_sync_ok": nil, "last_sync_error": nil,
	}
}

func (f *fakePolicySource) handle(call apiCall) (int, interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch call.Method {
	case http.MethodDelete:
		f.configured, f.enabled, f.endpoint, f.sync, f.poll, f.secret = false, false, nil, "", 0, ""
		return 200, f.view()
	case http.MethodPut:
		f.configured = true
		if v, ok := call.Body["enabled"].(bool); ok {
			f.enabled = v
		}
		if v, ok := call.Body["endpoint_url"].(string); ok {
			f.endpoint = v
		}
		if v, ok := call.Body["sync"].(string); ok {
			f.sync = v
		}
		if v, ok := call.Body["poll_interval_minutes"].(float64); ok {
			f.poll = int64(v)
		}
		if v, ok := call.Body["signing_secret"]; ok {
			s, _ := v.(string)
			f.secret = s
		}
	}
	return 200, f.view()
}

func sourcePlan(enabled bool, endpoint, syncMode string, poll types.Int64, secret types.String, version types.Int64) policySourceModel {
	return policySourceModel{
		ID: types.StringUnknown(), Enabled: types.BoolValue(enabled), EndpointURL: types.StringValue(endpoint),
		Sync: types.StringValue(syncMode), PollIntervalMinutes: poll, SigningSecretWO: secret, SigningSecretWOVersion: version,
		SigningSecretSet: types.BoolUnknown(), WebhookURL: types.StringUnknown(), ProtocolVersion: types.StringUnknown(),
		LastSyncedAt: types.Int64Unknown(), LastSyncOK: types.BoolUnknown(), LastSyncError: types.StringUnknown(),
	}
}

func TestPolicySource_lifecycleAndWriteOnlySecret(t *testing.T) {
	ctx := context.Background()
	fake := &fakePolicySource{}
	c, api := newFakeAPI(t, fake.handle)
	r := NewPolicySourceResource()
	configureResource(t, r, c)
	if typeNameOf(t, r) != "atlas_policy_source" {
		t.Fatalf("type name %s", typeNameOf(t, r))
	}
	const path = "/v1/instance/policy_source"
	const endpoint = "https://cms.example.com/atlas"

	state, resp := doCreate(t, r, planFrom(t, r, sourcePlan(true, endpoint, "webhook", types.Int64Unknown(), types.StringValue("whsec_0123456789abcdef"), types.Int64Value(1))))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	put := api.last(http.MethodPut, path)
	if put.Body["enabled"] != true || put.Body["endpoint_url"] != endpoint || put.Body["sync"] != "webhook" || put.Body["signing_secret"] != "whsec_0123456789abcdef" {
		t.Fatalf("PUT body = %v", put.Body)
	}
	if _, sent := put.Body["poll_interval_minutes"]; sent {
		t.Error("an unset poll interval must be omitted")
	}
	var m policySourceModel
	state.Get(ctx, &m)
	if m.ID.ValueString() != "policy_source" || !m.SigningSecretSet.ValueBool() || !m.SigningSecretWO.IsNull() || m.SigningSecretWOVersion.ValueInt64() != 1 ||
		m.WebhookURL.ValueString() == "" || m.PollIntervalMinutes.ValueInt64() != 60 || m.ProtocolVersion.ValueString() != "2026-10-24" {
		t.Errorf("state = %+v", m)
	}

	// Refresh is stable.
	refreshed, rr := doRead(t, r, state)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	if !refreshed.Raw.Equal(state.Raw) {
		t.Errorf("refresh drifted:\n%v\n%v", refreshed.Raw, state.Raw)
	}

	// Same version: the secret is not resent.
	plan2 := sourcePlan(true, endpoint, "poll", types.Int64Value(30), types.StringValue("whsec_0123456789abcdef"), types.Int64Value(1))
	state2, ur := doUpdate(t, r, planFrom(t, r, plan2), state)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	put = api.last(http.MethodPut, path)
	if _, sent := put.Body["signing_secret"]; sent || put.Body["sync"] != "poll" || put.Body["poll_interval_minutes"] != float64(30) {
		t.Errorf("PUT body = %v", put.Body)
	}

	// A new version with a secret rotates it.
	plan3 := sourcePlan(true, endpoint, "poll", types.Int64Value(30), types.StringValue("whsec_fedcba9876543210"), types.Int64Value(2))
	state3, ur := doUpdate(t, r, planFrom(t, r, plan3), state2)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	if api.last(http.MethodPut, path).Body["signing_secret"] != "whsec_fedcba9876543210" || fake.secret != "whsec_fedcba9876543210" {
		t.Error("a bumped version must send the new secret")
	}

	// A new version without a secret clears it.
	plan4 := sourcePlan(false, endpoint, "poll", types.Int64Value(30), types.StringNull(), types.Int64Value(3))
	state4, ur := doUpdate(t, r, planFrom(t, r, plan4), state3)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	if v, sent := api.last(http.MethodPut, path).Body["signing_secret"]; !sent || v != nil {
		t.Errorf("a bumped version without the secret must send null, got %v", v)
	}
	state4.Get(ctx, &m)
	if m.SigningSecretSet.ValueBool() || m.Enabled.ValueBool() {
		t.Errorf("state = %+v", m)
	}

	// Destroy disconnects; a refresh then drops the resource.
	if d := doDelete(t, r, state4); d.Diagnostics.HasError() || api.count(http.MethodDelete, path) != 1 {
		t.Errorf("delete: %v", d.Diagnostics)
	}
	_, rr = doRead(t, r, state4)
	if !rr.State.Raw.IsNull() {
		t.Error("a disconnected source must drop from state")
	}
	for _, call := range api.all() {
		if call.Auth != "Bearer sk_test_key" {
			t.Errorf("every call uses the secret key, got %q", call.Auth)
		}
	}
}

func TestPolicySource_apiErrorSurfaces(t *testing.T) {
	failing, _ := newFakeAPI(t, func(apiCall) (int, interface{}) {
		return 422, map[string]interface{}{"errors": []map[string]interface{}{{"code": "VALIDATION", "message": "enabling needs a signing_secret", "param": "signing_secret"}}}
	})
	r := NewPolicySourceResource()
	configureResource(t, r, failing)
	_, resp := doCreate(t, r, planFrom(t, r, sourcePlan(true, "https://cms.example.com/atlas", "webhook", types.Int64Unknown(), types.StringNull(), types.Int64Null())))
	if !resp.Diagnostics.HasError() || !strings.Contains(resp.Diagnostics[0].Detail(), "signing_secret") {
		t.Errorf("the API's refusal must surface: %v", resp.Diagnostics)
	}
}

func TestPolicySource_validation(t *testing.T) {
	ctx := context.Background()
	r := NewPolicySourceResource()
	validate := func(m policySourceModel) []string {
		p := planFrom(t, r, m)
		var resp resource.ValidateConfigResponse
		r.(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: p.Schema, Raw: p.Raw}}, &resp)
		var out []string
		for _, d := range resp.Diagnostics {
			out = append(out, d.Summary())
		}
		return out
	}
	ok := sourcePlan(true, "https://cms.example.com/atlas", "poll", types.Int64Value(15), types.StringValue("whsec_0123456789abcdef"), types.Int64Value(1))
	if got := validate(ok); len(got) != 0 {
		t.Errorf("a valid config must pass: %v", got)
	}
	cases := map[string]policySourceModel{
		"Invalid endpoint_url":          sourcePlan(true, "cms.example.com", "webhook", types.Int64Null(), types.StringNull(), types.Int64Null()),
		"Invalid sync":                  sourcePlan(true, "https://cms.example.com", "push", types.Int64Null(), types.StringNull(), types.Int64Null()),
		"Invalid poll_interval_minutes": sourcePlan(true, "https://cms.example.com", "poll", types.Int64Value(1), types.StringNull(), types.Int64Null()),
		"Invalid signing_secret_wo":     sourcePlan(true, "https://cms.example.com", "webhook", types.Int64Null(), types.StringValue("short"), types.Int64Null()),
	}
	for want, m := range cases {
		got := validate(m)
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s: got %v", want, got)
		}
	}
}
