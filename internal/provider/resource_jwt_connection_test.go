package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func jwtConnectionFake(t *testing.T) (*jwtConnectionResource, *fakeAPI) {
	conn := map[string]interface{}{
		"object": "jwt_connection", "id": "jwt_1", "organization_id": "org_1", "status": "active",
		"jwks_uri": "https://idp.example.com/.well-known/jwks.json", "issuer": "https://idp.example.com",
		"audience": "atlas", "email_claim": "email", "max_age_seconds": 300, "default_role_id": nil,
		"login_url": "https://acme-fapi.example.com/v1/sso/jwt/jwt_1", "created_at": 1, "updated_at": 2,
	}
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		switch {
		case call.Method == http.MethodDelete:
			return 200, map[string]interface{}{"object": "jwt_connection", "id": "jwt_1", "deleted": true}
		case call.Method == http.MethodPatch:
			if v, ok := call.Body["organization_id"]; ok {
				conn["organization_id"] = v
			}
			if v, ok := call.Body["max_age_seconds"]; ok {
				conn["max_age_seconds"] = v
			}
		case call.Method == http.MethodPost:
			return 201, conn
		}
		return 200, conn
	})
	r := NewJWTConnectionResource().(*jwtConnectionResource)
	configureResource(t, r, c)
	return r, api
}

func TestJWTConnection_lifecycle(t *testing.T) {
	ctx := context.Background()
	r, api := jwtConnectionFake(t)
	plan := jwtConnectionModel{
		ID: types.StringUnknown(), OrganizationID: types.StringValue("org_1"), Status: types.StringValue("active"),
		JWKSURI: types.StringValue("https://idp.example.com/.well-known/jwks.json"), Issuer: types.StringValue("https://idp.example.com"),
		Audience: types.StringValue("atlas"), EmailClaim: types.StringUnknown(), MaxAgeSeconds: types.Int64Unknown(),
		DefaultRoleID: types.StringNull(), LoginURL: types.StringUnknown(), CreatedAt: types.Int64Unknown(), UpdatedAt: types.Int64Unknown(),
	}
	state, resp := doCreate(t, r, planFrom(t, r, plan))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	post := api.last(http.MethodPost, "/v1/jwt_connections")
	if post == nil || post.Body["issuer"] != "https://idp.example.com" || post.Body["status"] != "active" || post.Body["organization_id"] != "org_1" {
		t.Fatalf("create body = %+v", post)
	}
	if _, sent := post.Body["email_claim"]; sent {
		t.Error("an unset email_claim must be left to Atlas's default")
	}
	var m jwtConnectionModel
	state.Get(ctx, &m)
	if m.ID.ValueString() != "jwt_1" || m.LoginURL.ValueString() == "" || m.EmailClaim.ValueString() != "email" || m.MaxAgeSeconds.ValueInt64() != 300 {
		t.Fatalf("state after create = %+v", m)
	}

	// Removing organization_id clears it (sent as null); max_age_seconds changes in place.
	next := m
	next.OrganizationID = types.StringNull()
	next.MaxAgeSeconds = types.Int64Value(600)
	updated, uresp := doUpdate(t, r, planFrom(t, r, next), state)
	mustOK(t, "update", uresp.Diagnostics.HasError(), uresp.Diagnostics)
	patch := api.last(http.MethodPatch, "/v1/jwt_connections/jwt_1")
	if v, sent := patch.Body["organization_id"]; !sent || v != nil {
		t.Errorf("a removed organization_id must be sent as null, got %v", patch.Body)
	}
	if patch.Body["max_age_seconds"] != float64(600) {
		t.Errorf("max_age_seconds sent = %v", patch.Body["max_age_seconds"])
	}
	var um jwtConnectionModel
	updated.Get(ctx, &um)
	if !um.OrganizationID.IsNull() || um.MaxAgeSeconds.ValueInt64() != 600 {
		t.Errorf("state after update = %+v", um)
	}

	if d := doDelete(t, r, updated); d.Diagnostics.HasError() || api.last(http.MethodDelete, "/v1/jwt_connections/jwt_1") == nil {
		t.Errorf("delete must DELETE /v1/jwt_connections/jwt_1: %v", d.Diagnostics)
	}
}

func TestJWTConnection_readGoneDropsState(t *testing.T) {
	c, _ := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 404, map[string]interface{}{"errors": []map[string]string{{"code": "NOT_FOUND", "message": "Unknown JWT connection."}}}
	})
	r := NewJWTConnectionResource()
	configureResource(t, r, c)
	m := jwtConnectionModel{
		ID: types.StringValue("jwt_gone"), OrganizationID: types.StringNull(), Status: types.StringValue("draft"),
		JWKSURI: types.StringNull(), Issuer: types.StringNull(), Audience: types.StringNull(), EmailClaim: types.StringValue("email"),
		MaxAgeSeconds: types.Int64Value(300), DefaultRoleID: types.StringNull(), LoginURL: types.StringValue("https://x.example.com"),
		CreatedAt: types.Int64Value(1), UpdatedAt: types.Int64Value(1),
	}
	_, rr := doRead(t, r, stateFrom(t, r, m))
	if rr.Diagnostics.HasError() || !rr.State.Raw.IsNull() {
		t.Errorf("a deleted connection must drop from state: %v", rr.Diagnostics)
	}
}
