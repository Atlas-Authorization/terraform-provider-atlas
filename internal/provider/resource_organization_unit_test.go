package provider

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestOrganization_backendCreatedWithExternalIDAndRoleMappings(t *testing.T) {
	ctx := context.Background()
	org := map[string]interface{}{
		"object": "organization", "id": "org_1", "name": "Acme", "slug": "acme", "image_url": nil,
		"max_allowed_memberships": nil, "created_by": nil, "external_id": "tenant_42",
		"public_metadata": map[string]interface{}{"tier": "gold"}, "role_mappings": map[string]string{"org:admin": "owner"},
		"created_at": 1, "updated_at": 2,
	}
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodPatch {
			if v, ok := call.Body["external_id"]; ok {
				if v == "" {
					v = nil
				}
				org["external_id"] = v
			}
			if v, ok := call.Body["role_mappings"].(map[string]interface{}); ok {
				if len(v) == 0 {
					org["role_mappings"] = nil
				} else {
					org["role_mappings"] = v
				}
			}
		}
		return 200, org
	})
	r := NewOrganizationResource()
	configureResource(t, r, c)
	plan := organizationModel{
		ID: types.StringUnknown(), Name: types.StringValue("Acme"), Slug: types.StringValue("acme"),
		CreatedBy: types.StringUnknown(), MaxAllowedMemberships: types.Int64Unknown(), ImageURL: types.StringUnknown(),
		CreatedAt: types.Int64Unknown(), UpdatedAt: types.Int64Unknown(),
	}
	plan.ExternalID = types.StringValue("tenant_42")
	plan.PublicMetadata = newJSONObjectValue(`{"tier":"gold"}`)
	plan.PrivateMetadata = newJSONObjectValue(`{"crm":"c_1"}`)
	plan.RoleMappings = types.MapValueMust(types.StringType, map[string]attr.Value{"org:admin": types.StringValue("owner")})
	state, resp := doCreate(t, r, planFrom(t, r, plan))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	post := api.last(http.MethodPost, "/v1/organizations")
	if _, sent := post.Body["created_by"]; sent {
		t.Error("an unset created_by must be omitted (backend-created organization)")
	}
	if post.Body["external_id"] != "tenant_42" ||
		!reflect.DeepEqual(post.Body["role_mappings"], map[string]interface{}{"org:admin": "owner"}) ||
		!reflect.DeepEqual(post.Body["private_metadata"], map[string]interface{}{"crm": "c_1"}) ||
		!reflect.DeepEqual(post.Body["public_metadata"], map[string]interface{}{"tier": "gold"}) {
		t.Fatalf("create body = %v", post.Body)
	}
	var m organizationModel
	state.Get(ctx, &m)
	if !m.CreatedBy.IsNull() || m.ExternalID.ValueString() != "tenant_42" || m.PrivateMetadata.IsNull() {
		t.Fatalf("state after create = %+v", m)
	}

	// Removing external_id and role_mappings clears both in place.
	next := m
	next.ExternalID = types.StringNull()
	next.RoleMappings = types.MapNull(types.StringType)
	updated, uresp := doUpdate(t, r, planFrom(t, r, next), state)
	mustOK(t, "update", uresp.Diagnostics.HasError(), uresp.Diagnostics)
	patch := api.last(http.MethodPatch, "/v1/organizations/org_1")
	if patch.Body["external_id"] != "" || !reflect.DeepEqual(patch.Body["role_mappings"], map[string]interface{}{}) {
		t.Errorf("clearing must send external_id \"\" and role_mappings {}, got %v", patch.Body)
	}
	if _, sent := patch.Body["private_metadata"]; sent {
		t.Error("an unchanged private_metadata must not be resent")
	}
	var um organizationModel
	updated.Get(ctx, &um)
	if !um.ExternalID.IsNull() || !um.RoleMappings.IsNull() || um.PrivateMetadata.IsNull() {
		t.Errorf("state after update = %+v", um)
	}
}
