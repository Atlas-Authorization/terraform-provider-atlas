package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBeforeUserCreateHook_apply(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 200, map[string]interface{}{
			"object": "before_user_create_hook", "enabled": true, "url": "https://app.example.com/hook",
			"signing_secret_set": true, "timeout_ms": 2000, "fail_closed": true,
		}
	})
	r := &beforeUserCreateHookResource{client: c}
	if typeNameOf(t, r) != "atlas_before_user_create_hook" {
		t.Fatalf("type name %s", typeNameOf(t, r))
	}
	const p = "/v1/instance/hooks/before_user_create"
	cfg := beforeUserCreateHookModel{
		Enabled: types.BoolValue(true), URL: types.StringValue("https://app.example.com/hook"),
		SigningSecretWO: types.StringValue("0123456789abcdef0123"), SigningSecretWOVersion: types.Int64Value(1),
		TimeoutMs: types.Int64Value(2000), FailClosed: types.BoolValue(true),
	}
	plan := cfg
	var diags diag.Diagnostics
	r.apply(ctx, &plan, &cfg, nil, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	b := api.last(http.MethodPatch, p).Body
	if b["enabled"] != true || b["fail_closed"] != true || b["url"] != "https://app.example.com/hook" ||
		b["signing_secret"] != "0123456789abcdef0123" || b["timeout_ms"] != float64(2000) {
		t.Errorf("create PATCH = %v", b)
	}
	if !plan.SigningSecretSet.ValueBool() || !plan.SigningSecretWO.IsNull() || plan.ID.ValueString() != "before_user_create_hook" {
		t.Errorf("state = %+v", plan)
	}

	// Same secret version: the secret is not re-sent.
	prior := plan
	plan = cfg
	r.apply(ctx, &plan, &cfg, &prior, &diags)
	if _, sent := api.last(http.MethodPatch, p).Body["signing_secret"]; sent {
		t.Error("unchanged secret version must not send the secret")
	}

	// New version without a secret clears it.
	cleared := cfg
	cleared.SigningSecretWO, cleared.SigningSecretWOVersion = types.StringNull(), types.Int64Value(2)
	plan = cleared
	r.apply(ctx, &plan, &cleared, &prior, &diags)
	if v, sent := api.last(http.MethodPatch, p).Body["signing_secret"]; !sent || v != nil {
		t.Errorf("version bump without a secret must clear it, got %v", v)
	}

	// Destroy switches the hook off and clears url and secret.
	if d := doDelete(t, r, stateFrom(t, r, prior)); d.Diagnostics.HasError() {
		t.Fatal(d.Diagnostics)
	}
	b = api.last(http.MethodPatch, p).Body
	if b["enabled"] != false || b["url"] != nil || b["signing_secret"] != nil {
		t.Errorf("delete PATCH = %v", b)
	}
}
