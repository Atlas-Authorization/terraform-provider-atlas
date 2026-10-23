package provider

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// fakeBridge is an in-memory Better Auth bridge with the API's merge semantics:
// a PUT merges onto the stored settings and the secret is write-only.
type fakeBridge struct {
	mu         sync.Mutex
	configured bool
	enabled    bool
	endsAt     interface{}
	prefix     string
	liveOn     bool
	baseURL    interface{}
	secret     string
}

func (b *fakeBridge) view() map[string]interface{} {
	var created interface{}
	if b.configured {
		created = 1
	}
	prefix := b.prefix
	if prefix == "" {
		prefix = "better-auth"
	}
	return map[string]interface{}{
		"object": "better_auth_bridge", "enabled": b.enabled, "ends_at": b.endsAt, "cookie_prefix": prefix,
		"live_fallback": map[string]interface{}{"enabled": b.liveOn, "base_url": b.baseURL, "has_secret": b.secret != ""},
		"exchange_path": "/v1/client/migrations/better_auth/exchange",
		"stats":         map[string]interface{}{"exchanged": 0}, "created_at": created, "updated_at": created,
	}
}

func (b *fakeBridge) handle(call apiCall) (int, interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch call.Method {
	case http.MethodDelete:
		b.configured, b.enabled, b.endsAt, b.prefix, b.liveOn, b.baseURL, b.secret = false, false, nil, "", false, nil, ""
		return 200, map[string]interface{}{"object": "better_auth_bridge", "deleted": true}
	case http.MethodPut:
		b.configured = true
		if v, ok := call.Body["enabled"].(bool); ok {
			b.enabled = v
		}
		if v, ok := call.Body["ends_at"]; ok {
			if s, isStr := v.(string); isStr {
				t, _ := time.Parse(time.RFC3339, s)
				b.endsAt = t.UnixMilli()
			} else {
				b.endsAt = nil
			}
		}
		if v, ok := call.Body["cookie_prefix"].(string); ok {
			b.prefix = v
		}
		if live, ok := call.Body["live_fallback"].(map[string]interface{}); ok {
			if v, ok := live["enabled"].(bool); ok {
				b.liveOn = v
			}
			if v, ok := live["base_url"]; ok {
				if s, isStr := v.(string); isStr {
					b.baseURL = strings.TrimRight(s, "/")
				} else {
					b.baseURL = nil
				}
			}
			if v, ok := live["secret"]; ok {
				s, _ := v.(string)
				b.secret = s
			}
		}
	}
	return 200, b.view()
}

func liveFallbackObject(t *testing.T, enabled bool, baseURL, secret types.String, version types.Int64) types.Object {
	t.Helper()
	obj, d := types.ObjectValue(betterAuthLiveFallbackAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(enabled), "base_url": baseURL, "secret_wo": secret,
		"secret_wo_version": version, "has_secret": types.BoolUnknown(),
	})
	if d.HasError() {
		t.Fatal(d)
	}
	return obj
}

func bridgeModel(enabled bool, endsAt types.String, prefix string, live types.Object) betterAuthBridgeModel {
	return betterAuthBridgeModel{
		ID: types.StringUnknown(), Enabled: types.BoolValue(enabled), EndsAt: endsAt,
		CookiePrefix: types.StringValue(prefix), LiveFallback: live, ExchangePath: types.StringUnknown(),
	}
}

func readLive(t *testing.T, st tfsdk.State) betterAuthLiveFallbackModel {
	t.Helper()
	var m betterAuthBridgeModel
	st.Get(context.Background(), &m)
	var diags diag.Diagnostics
	l := liveFallbackOf(context.Background(), m.LiveFallback, &diags)
	if l == nil {
		t.Fatal("live_fallback is null")
	}
	return *l
}

func TestBetterAuthBridge_lifecycleAndWriteOnlySecret(t *testing.T) {
	ctx := context.Background()
	fake := &fakeBridge{}
	c, api := newFakeAPI(t, fake.handle)
	r := NewBetterAuthBridgeResource()
	configureResource(t, r, c)
	if typeNameOf(t, r) != "atlas_better_auth_bridge" {
		t.Fatalf("type name %s", typeNameOf(t, r))
	}

	// Create with a window (offset spelling), a custom prefix and the live fallback with a secret.
	plan := planFrom(t, r, bridgeModel(true, types.StringValue("2026-12-31T02:00:00+02:00"), "myapp",
		liveFallbackObject(t, true, types.StringValue("https://app.example.com/api/auth/"), types.StringValue("0123456789abcdef0123"), types.Int64Value(1))))
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	put := api.last(http.MethodPut, "/v1/migrations/better_auth/bridge")
	if put == nil || put.Body["enabled"] != true || put.Body["cookie_prefix"] != "myapp" || put.Body["ends_at"] != "2026-12-31T02:00:00+02:00" {
		t.Fatalf("PUT body = %+v", put)
	}
	live, _ := put.Body["live_fallback"].(map[string]interface{})
	if live["secret"] != "0123456789abcdef0123" || live["enabled"] != true || live["base_url"] != "https://app.example.com/api/auth/" {
		t.Fatalf("live_fallback = %v", live)
	}
	var m betterAuthBridgeModel
	state.Get(ctx, &m)
	if m.ID.ValueString() != "better_auth_bridge" || m.EndsAt.ValueString() != "2026-12-31T02:00:00+02:00" ||
		m.ExchangePath.ValueString() != "/v1/client/migrations/better_auth/exchange" {
		t.Errorf("state = %+v", m)
	}
	l := readLive(t, state)
	if !l.SecretWO.IsNull() {
		t.Error("the write-only secret must never be stored in state")
	}
	if !l.HasSecret.ValueBool() || l.SecretWOVersion.ValueInt64() != 1 || l.BaseURL.ValueString() != "https://app.example.com/api/auth/" {
		t.Errorf("live state = %+v", l)
	}

	// Refresh is stable (same instant and the trailing-slash URL keep their spelling).
	refreshed, rr := doRead(t, r, state)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	if !refreshed.Raw.Equal(state.Raw) {
		t.Errorf("refresh drifted:\n%v\n%v", refreshed.Raw, state.Raw)
	}

	// An update with the same version does not resend the secret (it is not in state to compare).
	plan2 := planFrom(t, r, bridgeModel(true, types.StringNull(), "myapp",
		liveFallbackObject(t, true, types.StringValue("https://app.example.com/api/auth"), types.StringValue("0123456789abcdef0123"), types.Int64Value(1))))
	state2, ur := doUpdate(t, r, plan2, state)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	put = api.last(http.MethodPut, "/v1/migrations/better_auth/bridge")
	live, _ = put.Body["live_fallback"].(map[string]interface{})
	if _, sent := live["secret"]; sent {
		t.Error("an unchanged secret_wo_version must not resend the secret")
	}
	if v, sent := put.Body["ends_at"]; !sent || v != nil {
		t.Errorf("an unset ends_at must be sent as null, got %v", v)
	}
	if !readLive(t, state2).HasSecret.ValueBool() {
		t.Error("the stored secret must be kept")
	}

	// Bumping the version with no secret clears it.
	plan3 := planFrom(t, r, bridgeModel(true, types.StringNull(), "myapp",
		liveFallbackObject(t, false, types.StringValue("https://app.example.com/api/auth"), types.StringNull(), types.Int64Value(2))))
	state3, ur := doUpdate(t, r, plan3, state2)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	put = api.last(http.MethodPut, "/v1/migrations/better_auth/bridge")
	live, _ = put.Body["live_fallback"].(map[string]interface{})
	if v, sent := live["secret"]; !sent || v != nil {
		t.Errorf("a bumped version without secret_wo must send secret: null, got %v", live)
	}
	if readLive(t, state3).HasSecret.ValueBool() {
		t.Error("has_secret must be false after clearing")
	}

	// Rotating: a new version with a secret sends it.
	plan4 := planFrom(t, r, bridgeModel(true, types.StringNull(), "myapp",
		liveFallbackObject(t, true, types.StringValue("https://app.example.com/api/auth"), types.StringValue("fedcba9876543210fedc"), types.Int64Value(3))))
	state4, ur := doUpdate(t, r, plan4, state3)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	live, _ = api.last(http.MethodPut, "/v1/migrations/better_auth/bridge").Body["live_fallback"].(map[string]interface{})
	if live["secret"] != "fedcba9876543210fedc" || fake.secret != "fedcba9876543210fedc" {
		t.Errorf("rotation must send the new secret: %v", live)
	}

	// Removing the block turns the live fallback off and clears the secret.
	plan5 := planFrom(t, r, bridgeModel(false, types.StringNull(), "myapp", types.ObjectNull(betterAuthLiveFallbackAttrTypes)))
	state5, ur := doUpdate(t, r, plan5, state4)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	live, _ = api.last(http.MethodPut, "/v1/migrations/better_auth/bridge").Body["live_fallback"].(map[string]interface{})
	if live["enabled"] != false || live["base_url"] != nil || live["secret"] != nil {
		t.Errorf("an unset block must turn the fallback off and clear the secret: %v", live)
	}
	if _, sent := live["secret"]; !sent {
		t.Error("secret: null must be sent when the block is removed")
	}
	state5.Get(ctx, &m)
	if !m.LiveFallback.IsNull() || m.Enabled.ValueBool() {
		t.Errorf("state after removing the block = %+v", m)
	}

	// Destroy deletes; a refresh after an out-of-band delete drops the resource.
	if d := doDelete(t, r, state5); d.Diagnostics.HasError() || api.count(http.MethodDelete, "/v1/migrations/better_auth/bridge") != 1 {
		t.Errorf("delete must DELETE the bridge: %v", d.Diagnostics)
	}
	_, rr = doRead(t, r, state5)
	if !rr.State.Raw.IsNull() {
		t.Error("an unconfigured bridge must drop from state")
	}
	for _, call := range api.all() {
		if call.Auth != "Bearer sk_test_key" {
			t.Errorf("every call uses the secret key, got %q", call.Auth)
		}
	}
}

func TestBetterAuthBridge_importReadsLiveFallback(t *testing.T) {
	ctx := context.Background()
	fake := &fakeBridge{configured: true, enabled: true, endsAt: int64(1798675200000), prefix: "ba",
		liveOn: true, baseURL: "https://app.example.com/api/auth", secret: "s3cr3t-s3cr3t-s3cr3t"}
	c, _ := newFakeAPI(t, fake.handle)
	r := NewBetterAuthBridgeResource()
	configureResource(t, r, c)

	var ir resource.ImportStateResponse
	ir.State = nullState(t, r)
	r.(resource.ResourceWithImportState).ImportState(ctx, resource.ImportStateRequest{ID: "better_auth_bridge"}, &ir)
	mustOK(t, "import", ir.Diagnostics.HasError(), ir.Diagnostics)
	st, rr := doRead(t, r, ir.State)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	var m betterAuthBridgeModel
	st.Get(ctx, &m)
	if m.EndsAt.ValueString() != "2026-12-31T00:00:00Z" || m.CookiePrefix.ValueString() != "ba" || !m.Enabled.ValueBool() {
		t.Errorf("imported = %+v", m)
	}
	l := readLive(t, st)
	if !l.Enabled.ValueBool() || !l.HasSecret.ValueBool() || !l.SecretWO.IsNull() || !l.SecretWOVersion.IsNull() {
		t.Errorf("imported live_fallback = %+v", l)
	}
	var id types.String
	st.GetAttribute(ctx, path.Root("id"), &id)
	if id.ValueString() != "better_auth_bridge" {
		t.Errorf("id = %v", id)
	}
}

func TestBetterAuthBridge_defaultsAndAPIError(t *testing.T) {
	ctx := context.Background()
	fake := &fakeBridge{}
	c, api := newFakeAPI(t, fake.handle)
	r := NewBetterAuthBridgeResource()
	configureResource(t, r, c)
	state, resp := doCreate(t, r, planFrom(t, r, bridgeModel(false, types.StringNull(), "better-auth", types.ObjectNull(betterAuthLiveFallbackAttrTypes))))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	put := api.last(http.MethodPut, "/v1/migrations/better_auth/bridge")
	if put.Body["enabled"] != false || put.Body["cookie_prefix"] != "better-auth" {
		t.Errorf("PUT = %v", put.Body)
	}
	var m betterAuthBridgeModel
	state.Get(ctx, &m)
	if !m.LiveFallback.IsNull() {
		t.Error("an unset live_fallback must stay null when the API reports the defaults")
	}

	failing, _ := newFakeAPI(t, func(apiCall) (int, interface{}) {
		return 422, map[string]interface{}{"errors": []map[string]interface{}{{"code": "VALIDATION", "message": "live_fallback.base_url is not an acceptable endpoint", "param": "live_fallback.base_url"}}}
	})
	r2 := NewBetterAuthBridgeResource()
	configureResource(t, r2, failing)
	_, resp = doCreate(t, r2, planFrom(t, r2, bridgeModel(true, types.StringNull(), "better-auth",
		liveFallbackObject(t, true, types.StringValue("https://10.0.0.1/api/auth"), types.StringNull(), types.Int64Null()))))
	if !resp.Diagnostics.HasError() || !strings.Contains(resp.Diagnostics[0].Detail(), "acceptable endpoint") {
		t.Errorf("the API's refusal must surface: %v", resp.Diagnostics)
	}
}

func TestBetterAuthBridge_validation(t *testing.T) {
	ctx := context.Background()
	r := NewBetterAuthBridgeResource()
	validate := func(m betterAuthBridgeModel) []string {
		p := planFrom(t, r, m)
		var resp resource.ValidateConfigResponse
		r.(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: p.Schema, Raw: p.Raw}}, &resp)
		var out []string
		for _, d := range resp.Diagnostics {
			out = append(out, d.Summary())
		}
		return out
	}
	none := types.ObjectNull(betterAuthLiveFallbackAttrTypes)
	if got := validate(bridgeModel(true, types.StringValue("2026-12-31T00:00:00Z"), "better-auth", none)); len(got) != 0 {
		t.Errorf("a valid config must pass: %v", got)
	}
	cases := map[string]betterAuthBridgeModel{
		"Invalid cookie_prefix":                bridgeModel(true, types.StringNull(), "bad prefix!", none),
		"Invalid ends_at":                      bridgeModel(true, types.StringValue("next tuesday"), "ba", none),
		"live_fallback.base_url is required":   bridgeModel(true, types.StringNull(), "ba", liveFallbackObject(t, true, types.StringNull(), types.StringNull(), types.Int64Null())),
		"Invalid live_fallback.base_url":       bridgeModel(true, types.StringNull(), "ba", liveFallbackObject(t, false, types.StringValue("not a url"), types.StringNull(), types.Int64Null())),
		"live_fallback.secret_wo is too short": bridgeModel(true, types.StringNull(), "ba", liveFallbackObject(t, false, types.StringNull(), types.StringValue("short"), types.Int64Value(1))),
	}
	for want, m := range cases {
		got := validate(m)
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s: got %v", want, got)
		}
	}
}

func TestBetterAuthBridge_schema(t *testing.T) {
	s := schemaOf(t, NewBetterAuthBridgeResource())
	live, ok := s.Attributes["live_fallback"]
	if !ok {
		t.Fatal("live_fallback missing")
	}
	nested := live.(rschema.SingleNestedAttribute).Attributes
	secret := nested["secret_wo"]
	if !secret.IsWriteOnly() || !secret.IsSensitive() || secret.IsComputed() {
		t.Error("secret_wo must be write-only, sensitive and not computed")
	}
	if !nested["has_secret"].IsComputed() {
		t.Error("has_secret must be computed")
	}
}

// The provider server validates every schema (write-only rules included) when
// Terraform asks for it; the bridge's must be accepted and advertised.
func TestBetterAuthBridge_providerSchemaIsValid(t *testing.T) {
	ctx := context.Background()
	server, err := testAccProtoV6ProviderFactories["atlas"]()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range resp.Diagnostics {
		t.Errorf("schema diagnostic: %s: %s", d.Summary, d.Detail)
	}
	s, ok := resp.ResourceSchemas["atlas_better_auth_bridge"]
	if !ok {
		t.Fatal("atlas_better_auth_bridge is not registered")
	}
	var live *tfprotov6.SchemaAttribute
	for _, a := range s.Block.Attributes {
		if a.Name == "live_fallback" {
			live = a
		}
	}
	if live == nil || live.NestedType == nil {
		t.Fatal("live_fallback missing")
	}
	for _, a := range live.NestedType.Attributes {
		if a.Name == "secret_wo" && (!a.WriteOnly || !a.Sensitive) {
			t.Error("secret_wo must be advertised as write-only and sensitive")
		}
	}
}
