package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPolicySettings_lifecycle(t *testing.T) {
	ctx := context.Background()
	claim := false
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodPatch {
			claim, _ = call.Body["session_claim"].(bool)
		}
		return 200, map[string]interface{}{"object": "policy_settings", "session_claim": claim}
	})
	r := NewPolicySettingsResource()
	configureResource(t, r, c)
	if typeNameOf(t, r) != "atlas_policy_settings" {
		t.Fatalf("type name %s", typeNameOf(t, r))
	}
	const path = "/v1/instance/policy_settings"

	state, resp := doCreate(t, r, planFrom(t, r, policySettingsModel{ID: types.StringUnknown(), SessionClaim: types.BoolValue(true)}))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	if p := api.last(http.MethodPatch, path); p == nil || p.Body["session_claim"] != true {
		t.Fatalf("PATCH = %+v", p)
	}
	var m policySettingsModel
	state.Get(ctx, &m)
	if m.ID.ValueString() != "policy_settings" || !m.SessionClaim.ValueBool() {
		t.Errorf("state = %+v", m)
	}

	// An out-of-band change shows as drift on refresh.
	claim = false
	drifted, rr := doRead(t, r, state)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	drifted.Get(ctx, &m)
	if m.SessionClaim.ValueBool() {
		t.Error("refresh must read the live value")
	}

	state2, ur := doUpdate(t, r, planFrom(t, r, policySettingsModel{ID: types.StringValue("policy_settings"), SessionClaim: types.BoolValue(true)}), drifted)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	state2.Get(ctx, &m)
	if !m.SessionClaim.ValueBool() || !claim {
		t.Error("update must turn the claim on")
	}

	// Destroy turns the claim off.
	if d := doDelete(t, r, state2); d.Diagnostics.HasError() || claim {
		t.Errorf("delete must turn the claim off: %v", d.Diagnostics)
	}
}
