package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func designResponse(design map[string]interface{}, org bool) map[string]interface{} {
	out := map[string]interface{}{
		"object": "design", "design": design,
		"resolved": map[string]interface{}{"version": 1, "preset": "minimal"},
		"warnings": []interface{}{}, "legacy": false, "legacy_design": nil,
	}
	if org {
		out["object"] = "organization_design"
		out["organization_id"] = "org_1"
	}
	return out
}

func TestInstanceDesign_putReadDelete(t *testing.T) {
	ctx := context.Background()
	stored := map[string]interface{}{}
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodPut {
			stored, _ = call.Body["design"].(map[string]interface{})
		}
		return 200, designResponse(stored, false)
	})
	r := NewInstanceDesignResource()
	configureResource(t, r, c)
	if typeNameOf(t, r) != "atlas_instance_design" {
		t.Fatalf("type name = %s", typeNameOf(t, r))
	}

	plan := planFrom(t, r, instanceDesignModel{
		ID:       types.StringUnknown(),
		Design:   newJSONObjectValue(`{"preset":"midnight","palette":{"primary":"#112233"}}`),
		Resolved: types.StringUnknown(), Warnings: types.StringUnknown(),
	})
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	put := api.last(http.MethodPut, "/v1/instance/design")
	if put == nil || put.Body["design"] == nil {
		t.Fatalf("create must PUT { design }, got %+v", put)
	}
	var m instanceDesignModel
	state.Get(ctx, &m)
	if m.ID.ValueString() != "instance" || m.Resolved.IsNull() {
		t.Errorf("computed fields not set: %+v", m)
	}
	if eq, _ := newJSONObjectValue(`{"palette":{"primary":"#112233"},"preset":"midnight"}`).StringSemanticEquals(ctx, m.Design); !eq {
		t.Errorf("design read-back mismatch: %s", m.Design.ValueString())
	}

	if _, rr := doRead(t, r, state); rr.Diagnostics.HasError() {
		t.Fatalf("read: %v", rr.Diagnostics)
	}

	del := doDelete(t, r, state)
	mustOK(t, "delete", del.Diagnostics.HasError(), del.Diagnostics)
	last := api.last(http.MethodPut, "/v1/instance/design")
	if d, ok := last.Body["design"].(map[string]interface{}); !ok || len(d) != 0 {
		t.Errorf("destroy must PUT { design: {} }, got %v", last.Body)
	}
}

func TestOrganizationDesign_pathAndImport(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 200, designResponse(map[string]interface{}{"preset": "glass"}, true)
	})
	r := NewOrganizationDesignResource()
	configureResource(t, r, c)
	plan := planFrom(t, r, designModel{
		ID: types.StringUnknown(), OrganizationID: types.StringValue("org_1"),
		Design:   newJSONObjectValue(`{"preset":"glass"}`),
		Resolved: types.StringUnknown(), Warnings: types.StringUnknown(),
	})
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	if api.last(http.MethodPut, "/v1/organizations/org_1/design") == nil {
		t.Fatal("org design must PUT /v1/organizations/:id/design")
	}
	var m designModel
	state.Get(ctx, &m)
	if m.ID.ValueString() != "org_1" {
		t.Errorf("id = %s, want org_1", m.ID.ValueString())
	}
	s := schemaOf(t, r)
	if a := s.Attributes["organization_id"]; a == nil || !a.IsRequired() {
		t.Error("organization_id should be required")
	}
}
