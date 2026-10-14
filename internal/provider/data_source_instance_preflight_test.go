package provider

import (
	"context"
	"testing"
)

func TestInstancePreflightDataSource(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(apiCall) (int, interface{}) {
		return 200, map[string]interface{}{
			"object": "preflight", "ok": false, "checked_at": 99,
			"findings": []map[string]interface{}{
				{"severity": "error", "code": "oauth_provider_missing_credentials", "message": "m1", "fix": "f1", "subject": map[string]string{"provider": "google"}},
				{"severity": "warning", "code": "domain_not_serving", "message": "m2", "fix": "f2"},
				{"severity": "info", "code": "shared_dev_sender", "message": "m3", "fix": "f3"},
			},
		}
	})
	state, resp := readDataSource(t, NewInstancePreflightDataSource(), c)
	mustOK(t, "read", resp.Diagnostics.HasError(), resp.Diagnostics)
	if api.last("GET", "/v1/instance/preflight") == nil {
		t.Fatal("must GET /v1/instance/preflight")
	}
	var m instancePreflightModel
	mustOK(t, "get", state.Get(ctx, &m).HasError(), nil)
	if m.OK.ValueBool() || m.Errors.ValueInt64() != 1 || m.Warnings.ValueInt64() != 1 || len(m.Findings.Elements()) != 3 {
		t.Errorf("preflight model = ok:%v errors:%d warnings:%d findings:%d", m.OK.ValueBool(), m.Errors.ValueInt64(), m.Warnings.ValueInt64(), len(m.Findings.Elements()))
	}
	if m.CheckedAt.ValueInt64() != 99 || m.ID.ValueString() != "preflight" {
		t.Errorf("checked_at/id = %d/%s", m.CheckedAt.ValueInt64(), m.ID.ValueString())
	}
}
