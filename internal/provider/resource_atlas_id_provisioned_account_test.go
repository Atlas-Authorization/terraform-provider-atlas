package provider

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const provBase = "/v1/platform/atlas_id_apps/oac_1/provisioned_accounts"

func provisionedResponse(extra map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{
		"object": "atlas_id_provisioned_account", "id": "iacct_1", "sub": "usr_1",
		"email": "reviewer@example.com", "email_verified": true, "label": "reviewer",
		"kind": "provisioned", "app_ids": []string{"oac_1"}, "two_factor_exempt": true,
		"expires_at": nil, "expired": false, "last_sign_in_at": nil, "created_at": 5, "updated_at": 5,
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func baseProvisionedModel() atlasIDProvisionedAccountModel {
	return atlasIDProvisionedAccountModel{
		ID: types.StringUnknown(), AppID: types.StringValue("oac_1"), Email: types.StringValue("reviewer@example.com"),
		Label: types.StringValue("reviewer"), EmailVerified: types.BoolUnknown(), TwoFactorExempt: types.BoolValue(true),
		ExpiresAt: types.StringNull(), PasswordWO: types.StringNull(), PasswordWOVersion: types.Int64Null(),
		GeneratedPassword: types.StringUnknown(), Sub: types.StringUnknown(), Kind: types.StringUnknown(),
		Expired: types.BoolUnknown(), CreatedAt: types.Int64Unknown(),
	}
}

func TestProvisionedAccount_typeNameAndWriteOnlyPassword(t *testing.T) {
	r := NewAtlasIDProvisionedAccountResource()
	if got := typeNameOf(t, r); got != "atlas_id_provisioned_account" {
		t.Fatalf("type name = %s", got)
	}
	s := schemaOf(t, r)
	pw, ok := s.Attributes["password_wo"]
	if !ok || !pw.IsWriteOnly() || !pw.IsSensitive() {
		t.Fatalf("password_wo must be a sensitive write-only attribute")
	}
	for _, name := range []string{"generated_password"} {
		if !s.Attributes[name].IsSensitive() {
			t.Errorf("%s must be sensitive", name)
		}
	}
	if _, ok := s.Attributes["password"]; ok {
		t.Error("there must be no plain password attribute")
	}
}

func TestProvisionedAccount_createWithSuppliedPasswordKeepsItOutOfState(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodPost && call.Path == provBase {
			return 201, provisionedResponse(map[string]interface{}{"password": "Chosen-Password-123"})
		}
		return 404, map[string]interface{}{"errors": []map[string]string{{"code": "NOT_FOUND", "message": "x"}}}
	})
	r := NewAtlasIDProvisionedAccountResource()
	configureResource(t, r, c)

	m := baseProvisionedModel()
	m.PasswordWO = types.StringValue("Chosen-Password-123")
	m.PasswordWOVersion = types.Int64Value(1)
	m.ExpiresAt = types.StringValue("2030-01-01T00:00:00Z")
	state, resp := doCreate(t, r, planFrom(t, r, m))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)

	call := api.last(http.MethodPost, provBase)
	if call == nil || call.Auth != "Bearer psk_test_key" {
		t.Fatalf("create must use the platform key, got %+v", call)
	}
	if call.Body["password"] != "Chosen-Password-123" || call.Body["email"] != "reviewer@example.com" ||
		call.Body["label"] != "reviewer" || call.Body["two_factor_exempt"] != true {
		t.Errorf("body = %v", call.Body)
	}
	if call.Body["expires_at"] != float64(1893456000000) {
		t.Errorf("expires_at must be sent as epoch ms, got %v", call.Body["expires_at"])
	}
	if _, sent := call.Body["email_verified"]; sent {
		t.Errorf("an unset email_verified must be omitted: %v", call.Body)
	}
	var got atlasIDProvisionedAccountModel
	mustOK(t, "state", state.Get(ctx, &got).HasError(), nil)
	if !got.PasswordWO.IsNull() || !got.GeneratedPassword.IsNull() {
		t.Errorf("a supplied password must not be kept in state: %+v", got)
	}
	if got.ID.ValueString() != "iacct_1" || got.Sub.ValueString() != "usr_1" || got.Kind.ValueString() != "provisioned" || !got.EmailVerified.ValueBool() {
		t.Errorf("computed fields not stored: %+v", got)
	}
}

func TestProvisionedAccount_createGeneratedPasswordLandsInSensitiveAttribute(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 201, provisionedResponse(map[string]interface{}{"password": "Generated-By-Atlas-9"})
	})
	r := NewAtlasIDProvisionedAccountResource()
	configureResource(t, r, c)
	state, resp := doCreate(t, r, planFrom(t, r, baseProvisionedModel()))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	if _, sent := api.last(http.MethodPost, provBase).Body["password"]; sent {
		t.Error("no password must be sent when none is configured")
	}
	var got atlasIDProvisionedAccountModel
	state.Get(ctx, &got)
	if got.GeneratedPassword.ValueString() != "Generated-By-Atlas-9" {
		t.Errorf("generated password not stored: %+v", got.GeneratedPassword)
	}
}

func TestProvisionedAccount_rejectsShortPasswordAndBadExpiry(t *testing.T) {
	c, api := newFakeAPI(t, func(apiCall) (int, interface{}) { return 201, provisionedResponse(nil) })
	r := NewAtlasIDProvisionedAccountResource()
	configureResource(t, r, c)

	short := baseProvisionedModel()
	short.PasswordWO = types.StringValue("short")
	if _, resp := doCreate(t, r, planFrom(t, r, short)); !resp.Diagnostics.HasError() {
		t.Error("a password under 12 characters must be refused before any request")
	}
	bad := baseProvisionedModel()
	bad.ExpiresAt = types.StringValue("next tuesday")
	if _, resp := doCreate(t, r, planFrom(t, r, bad)); !resp.Diagnostics.HasError() {
		t.Error("a non-RFC3339 expires_at must be refused")
	}
	if len(api.all()) != 0 {
		t.Errorf("no request may be sent: %+v", api.all())
	}
}

func TestProvisionedAccount_rotateOnVersionChangeOnly(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		switch {
		case call.Method == http.MethodPost && call.Path == provBase:
			return 201, provisionedResponse(map[string]interface{}{"password": "First-Password-12345"})
		case call.Method == http.MethodPost && call.Path == provBase+"/iacct_1/password":
			return 200, provisionedResponse(map[string]interface{}{"password": "Generated-Again-77"})
		case call.Method == http.MethodGet:
			return 200, provisionedResponse(nil)
		}
		return 404, map[string]interface{}{"errors": []map[string]string{{"code": "NOT_FOUND", "message": "x"}}}
	})
	r := NewAtlasIDProvisionedAccountResource()
	configureResource(t, r, c)

	first := baseProvisionedModel()
	first.PasswordWO = types.StringValue("First-Password-12345")
	first.PasswordWOVersion = types.Int64Value(1)
	state, resp := doCreate(t, r, planFrom(t, r, first))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)

	var cur atlasIDProvisionedAccountModel
	state.Get(ctx, &cur)

	// Same version: no rotation.
	same := cur
	same.PasswordWO = types.StringValue("First-Password-12345")
	st2, uresp := doUpdate(t, r, planFrom(t, r, same), state)
	mustOK(t, "update", uresp.Diagnostics.HasError(), uresp.Diagnostics)
	if api.count(http.MethodPost, provBase+"/iacct_1/password") != 0 {
		t.Error("an unchanged password_wo_version must not rotate")
	}

	// New version, no password supplied: Atlas generates one.
	next := cur
	next.PasswordWOVersion = types.Int64Value(2)
	next.PasswordWO = types.StringNull()
	next.GeneratedPassword = types.StringUnknown()
	st3, uresp := doUpdate(t, r, planFrom(t, r, next), st2)
	mustOK(t, "rotate", uresp.Diagnostics.HasError(), uresp.Diagnostics)
	rot := api.last(http.MethodPost, provBase+"/iacct_1/password")
	if rot == nil || len(rot.Body) != 0 {
		t.Fatalf("a generated rotation sends an empty body, got %+v", rot)
	}
	var after atlasIDProvisionedAccountModel
	st3.Get(ctx, &after)
	if after.GeneratedPassword.ValueString() != "Generated-Again-77" {
		t.Errorf("rotated password not stored: %+v", after.GeneratedPassword)
	}

	// New version with a supplied password: it is sent, and nothing is kept.
	third := after
	third.PasswordWOVersion = types.Int64Value(3)
	third.PasswordWO = types.StringValue("Supplied-Rotation-Pw-1")
	third.GeneratedPassword = types.StringUnknown()
	st4, uresp := doUpdate(t, r, planFrom(t, r, third), st3)
	mustOK(t, "rotate 2", uresp.Diagnostics.HasError(), uresp.Diagnostics)
	if got := api.last(http.MethodPost, provBase+"/iacct_1/password").Body["password"]; got != "Supplied-Rotation-Pw-1" {
		t.Errorf("supplied password not sent: %v", got)
	}
	var last atlasIDProvisionedAccountModel
	st4.Get(ctx, &last)
	if !last.GeneratedPassword.IsNull() {
		t.Error("a supplied rotation must clear generated_password")
	}
}

func TestProvisionedAccount_readAndDelete(t *testing.T) {
	ctx := context.Background()
	gone := false
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		switch call.Method {
		case http.MethodGet:
			if gone {
				return 404, map[string]interface{}{"errors": []map[string]string{{"code": "NOT_FOUND", "message": "x"}}}
			}
			return 200, provisionedResponse(map[string]interface{}{"expires_at": 1893456000000, "expired": false})
		case http.MethodDelete:
			return 200, map[string]interface{}{"deleted": true}
		}
		return 404, nil
	})
	r := NewAtlasIDProvisionedAccountResource()
	configureResource(t, r, c)

	m := baseProvisionedModel()
	m.ID = types.StringValue("iacct_1")
	m.ExpiresAt = types.StringValue("2030-01-01T00:00:00Z")
	m.Sub, m.Kind, m.Expired, m.CreatedAt = types.StringValue("usr_1"), types.StringValue("provisioned"), types.BoolValue(false), types.Int64Value(5)
	m.EmailVerified, m.GeneratedPassword = types.BoolValue(true), types.StringNull()
	state := stateFrom(t, r, m)

	refreshed, resp := doRead(t, r, state)
	mustOK(t, "read", resp.Diagnostics.HasError(), resp.Diagnostics)
	var got atlasIDProvisionedAccountModel
	refreshed.Get(ctx, &got)
	if got.ExpiresAt.ValueString() != "2030-01-01T00:00:00Z" {
		t.Errorf("an equal expiry must keep the configured spelling, got %s", got.ExpiresAt.ValueString())
	}

	if d := doDelete(t, r, refreshed); d.Diagnostics.HasError() {
		t.Fatalf("delete: %v", d.Diagnostics)
	}
	if api.last(http.MethodDelete, provBase+"/iacct_1") == nil {
		t.Error("delete did not call DELETE")
	}

	gone = true
	removed, rresp := doRead(t, r, state)
	mustOK(t, "read 404", rresp.Diagnostics.HasError(), rresp.Diagnostics)
	if !removed.Raw.IsNull() {
		t.Error("an account removed on the server must leave state")
	}
}

func TestProvisionedAccount_import(t *testing.T) {
	r := NewAtlasIDProvisionedAccountResource().(resource.ResourceWithImportState)
	imp := func(id string) resource.ImportStateResponse {
		resp := resource.ImportStateResponse{State: nullState(t, NewAtlasIDProvisionedAccountResource())}
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &resp)
		return resp
	}
	if resp := imp("oac_1"); !resp.Diagnostics.HasError() {
		t.Error("an id without the app must be refused")
	}
	resp := imp("oac_1/iacct_1")
	if resp.Diagnostics.HasError() {
		t.Fatalf("import: %v", resp.Diagnostics)
	}
	var app, id string
	resp.State.GetAttribute(context.Background(), path.Root("app_id"), &app)
	resp.State.GetAttribute(context.Background(), path.Root("id"), &id)
	if app != "oac_1" || id != "iacct_1" || strings.Contains(id, "/") {
		t.Errorf("imported %q %q", app, id)
	}
}

func TestAtlasIDApp_testModeSettings(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		switch call.Method {
		case http.MethodPost:
			return 201, appResponse(map[string]interface{}{"client_secret": "oacs_x", "test_mode": true, "provisioned_verified_domains_only": true})
		case http.MethodPatch:
			return 200, appResponse(map[string]interface{}{"test_mode": false, "provisioned_verified_domains_only": true})
		}
		return 200, appResponse(map[string]interface{}{"test_mode": false})
	})
	r := NewAtlasIDAppResource()
	configureResource(t, r, c)

	plan := baseAppModel(ctx)
	plan.TestMode = types.BoolValue(true)
	plan.ProvisionedVerifiedDomainsOnly = types.BoolValue(true)
	state, resp := doCreate(t, r, planFrom(t, r, plan))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	body := api.last(http.MethodPost, "/v1/platform/atlas_id_apps").Body
	if body["test_mode"] != true || body["provisioned_verified_domains_only"] != true {
		t.Errorf("test-mode settings not sent on create: %v", body)
	}
	var m atlasIDAppModel
	state.Get(ctx, &m)
	if !m.TestMode.ValueBool() || !m.ProvisionedVerifiedDomainsOnly.ValueBool() {
		t.Errorf("settings not stored: %+v", m)
	}

	next := m
	next.TestMode = types.BoolValue(false)
	updated, uresp := doUpdate(t, r, planFrom(t, r, next), state)
	mustOK(t, "update", uresp.Diagnostics.HasError(), uresp.Diagnostics)
	if v, ok := api.last(http.MethodPatch, "/v1/platform/atlas_id_apps/oac_1").Body["test_mode"]; !ok || v != false {
		t.Errorf("turning test mode off must be sent as false, got %v", v)
	}
	var um atlasIDAppModel
	updated.Get(ctx, &um)
	if um.TestMode.ValueBool() {
		t.Error("test_mode must read back false")
	}
}
