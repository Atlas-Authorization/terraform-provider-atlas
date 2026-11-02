package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestOrganizationPolicy_endUserPoolAndClearOnDelete(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 200, map[string]interface{}{
			"object": "organization_policy", "organization_id": "org_1",
			"policy": map[string]interface{}{"requireMfa": true, "endUserPool": map[string]interface{}{"enabled": true, "requireApproval": true}},
		}
	})
	r := NewOrganizationPolicyResource()
	configureResource(t, r, c)
	pool := types.ObjectValueMust(endUserPoolAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(true), "require_approval": types.BoolValue(true), "allow_hosted_signup": types.BoolNull(),
	})
	plan := planFrom(t, r, organizationPolicyModel{
		ID: types.StringUnknown(), OrganizationID: types.StringValue("org_1"), RequireMfa: types.BoolValue(true),
		SsoRequired: types.BoolUnknown(), SessionIdleOverrideMs: types.Int64Unknown(),
		AllowedSignInMethods: types.SetUnknown(types.StringType), IPAllowlist: types.SetUnknown(types.StringType),
		SsoBypass: types.SetUnknown(types.StringType), ItContacts: types.SetUnknown(types.StringType),
		MaxSessionAgeSeconds: types.Int64Unknown(), EndUserPool: pool,
	})
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	patch := api.last(http.MethodPatch, "/v1/organizations/org_1/policy")
	eup, _ := patch.Body["endUserPool"].(map[string]interface{})
	if eup["enabled"] != true || eup["requireApproval"] != true {
		t.Errorf("endUserPool must be sent camelCase: %v", patch.Body)
	}
	if _, sent := eup["allowHostedSignup"]; sent {
		t.Error("unset allowHostedSignup must be omitted")
	}
	var m organizationPolicyModel
	state.Get(ctx, &m)
	if m.EndUserPool.IsNull() {
		t.Error("end_user_pool not read back")
	}

	d := doDelete(t, r, state)
	mustOK(t, "delete", d.Diagnostics.HasError(), d.Diagnostics)
	clear := api.last(http.MethodPatch, "/v1/organizations/org_1/policy")
	for _, k := range client.OrgPolicyKeys {
		if v, ok := clear.Body[k]; !ok || v != nil {
			t.Errorf("destroy must clear %s with null, body=%v", k, clear.Body)
		}
	}
}

func TestProviderSchema_platformKeyAndRegistrations(t *testing.T) {
	ctx := context.Background()
	p := New("test")()
	var sr = providerSchema(t, p)
	pk, ok := sr.Attributes["platform_key"]
	if !ok || !pk.IsSensitive() || !pk.IsOptional() {
		t.Error("platform_key must be an optional sensitive provider attribute")
	}
	want := map[string]bool{
		"atlas_id_app": false, "atlas_instance_design": false, "atlas_organization_design": false,
		"atlas_cookie_consent": false, "atlas_entitlement_override": false,
		"atlas_email_provider": false, "atlas_sms_provider": false,
	}
	for _, f := range p.Resources(ctx) {
		if _, ok := want[typeNameOf(t, f())]; ok {
			want[typeNameOf(t, f())] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("%s not registered", name)
		}
	}
	found := false
	for _, f := range p.DataSources(ctx) {
		if dataSourceTypeName(f()) == "atlas_instance_preflight" {
			found = true
		}
	}
	if !found {
		t.Error("atlas_instance_preflight data source not registered")
	}
}
