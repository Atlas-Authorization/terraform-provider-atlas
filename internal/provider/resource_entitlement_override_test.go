package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEntitlementOverride_lifecycle(t *testing.T) {
	ctx := context.Background()
	row := map[string]interface{}{
		"object": "entitlement_override", "id": "eo_1", "subject_type": "org", "subject_id": "org_1",
		"plan": "enterprise", "features": []string{"sso"}, "limits": map[string]interface{}{"seats": 500, "note": "x"},
		"expires_at": nil, "created_by": "ak_1", "created_at": 10,
	}
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodPost {
			return 201, row
		}
		if call.Method == http.MethodGet {
			return 200, map[string]interface{}{"object": "list", "data": []interface{}{row}, "has_more": false}
		}
		return 200, map[string]interface{}{"deleted": true}
	})
	r := NewEntitlementOverrideResource()
	configureResource(t, r, c)
	features, _ := types.SetValueFrom(ctx, types.StringType, []string{"sso"})
	limits, _ := types.MapValueFrom(ctx, types.Float64Type, map[string]float64{"seats": 500})
	plan := planFrom(t, r, entitlementOverrideModel{
		ID: types.StringUnknown(), SubjectType: types.StringValue("org"), SubjectID: types.StringValue("org_1"),
		Plan: types.StringValue("enterprise"), Features: features, Limits: limits,
		ExpiresAt: types.Int64Null(), CreatedAt: types.Int64Unknown(),
	})
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	post := api.last(http.MethodPost, "/v1/entitlement_overrides")
	if post.Body["subject_type"] != "org" || post.Body["plan"] != "enterprise" {
		t.Errorf("POST body = %v", post.Body)
	}
	if _, sent := post.Body["expires_at"]; sent {
		t.Error("unset expires_at must be omitted")
	}
	if l, _ := post.Body["limits"].(map[string]interface{}); l["seats"] != float64(500) {
		t.Errorf("limits = %v", post.Body["limits"])
	}

	_, rr := doRead(t, r, state)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	get := api.last(http.MethodGet, "/v1/entitlement_overrides")
	if get.Query != "subject_id=org_1&subject_type=org" {
		t.Errorf("read must filter the list by subject, query=%q", get.Query)
	}
	var m entitlementOverrideModel
	rr.State.Get(ctx, &m)
	var gotLimits map[string]float64
	m.Limits.ElementsAs(ctx, &gotLimits, false)
	if len(gotLimits) != 1 || gotLimits["seats"] != 500 {
		t.Errorf("non-numeric limits must be skipped, numeric kept: %v", gotLimits)
	}

	if d := doDelete(t, r, state); d.Diagnostics.HasError() || api.last(http.MethodDelete, "/v1/entitlement_overrides/eo_1") == nil {
		t.Error("delete must DELETE /v1/entitlement_overrides/:id")
	}

	// Every input forces replacement (grant-only API).
	for name, a := range schemaOf(t, r).Attributes {
		if name == "id" || name == "created_at" {
			continue
		}
		if !a.IsRequired() && !a.IsOptional() {
			t.Errorf("%s should be configurable", name)
		}
	}
}

func TestEntitlementOverride_importComposite(t *testing.T) {
	r := NewEntitlementOverrideResource().(resource.ResourceWithImportState)
	st := nullState(t, NewEntitlementOverrideResource())
	resp := resource.ImportStateResponse{State: st}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "user/user_9/eo_7"}, &resp)
	mustOK(t, "import", resp.Diagnostics.HasError(), resp.Diagnostics)
	var m entitlementOverrideModel
	resp.State.Get(context.Background(), &m)
	if m.ID.ValueString() != "eo_7" || m.SubjectType.ValueString() != "user" || m.SubjectID.ValueString() != "user_9" {
		t.Errorf("import = %+v", m)
	}
}
