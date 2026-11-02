package provider

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestInstanceConfig_accountPortalPanels(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		switch {
		case call.Path == "/v1/instance/branding" && call.Method == http.MethodGet:
			return 200, map[string]interface{}{"object": "branding", "accountPortal": map[string]interface{}{
				"panels": map[string]interface{}{"delete_account": false, "ciba": nil},
			}}
		case call.Path == "/v1/instance/branding":
			return 200, map[string]interface{}{"object": "branding"}
		}
		return 200, map[string]interface{}{"id": "ins_1", "environment": "production", "publishable_key": "pk_live_x",
			"frontend_api_host": "fapi.example.com", "allowed_origins": []string{}, "auth_config": map[string]interface{}{}, "created_at": 1}
	})
	r := &instanceConfigResource{client: c}
	panels := types.MapValueMust(types.BoolType, map[string]attr.Value{"delete_account": types.BoolValue(false), "mfa": types.BoolValue(true)})
	prior := types.MapValueMust(types.BoolType, map[string]attr.Value{"referrals": types.BoolValue(false)})
	m := &instanceConfigModel{AuthConfig: types.StringNull(), AccountPortalPanels: panels}
	var diags diag.Diagnostics
	r.write(ctx, m, prior, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	patch := api.last(http.MethodPatch, "/v1/instance/branding")
	if patch == nil {
		t.Fatal("account_portal_panels must PATCH /v1/instance/branding")
	}
	want := map[string]interface{}{"accountPortal": map[string]interface{}{"panels": map[string]interface{}{
		"delete_account": false, "mfa": true, "referrals": nil,
	}}}
	if !reflect.DeepEqual(patch.Body, want) {
		t.Errorf("branding patch = %v, want %v (a dropped panel is reset with null)", patch.Body, want)
	}

	// Unmanaged (null on both sides): no branding call.
	before := api.count(http.MethodPatch, "/v1/instance/branding")
	r.write(ctx, &instanceConfigModel{AuthConfig: types.StringNull(), AccountPortalPanels: types.MapNull(types.BoolType)}, types.MapNull(types.BoolType), &diags)
	if api.count(http.MethodPatch, "/v1/instance/branding") != before {
		t.Error("unset account_portal_panels must not touch the branding")
	}

	// Read-back drops null (default) toggles.
	b, err := c.GetBranding(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := accountPortalPanelsFromAPI(ctx, b, &diags)
	if want := types.MapValueMust(types.BoolType, map[string]attr.Value{"delete_account": types.BoolValue(false)}); !got.Equal(want) {
		t.Errorf("panels read back = %v", got)
	}
}
