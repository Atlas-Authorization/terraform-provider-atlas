package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// androidAppAttrTypes / nativeAppsAttrTypes mirror the schema's nested object
// types so a test can build a native_apps types.Object with ObjectValueFrom.
func androidAppAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"package_name":             types.StringType,
		"sha256_cert_fingerprints": types.ListType{ElemType: types.StringType},
	}
}

func nativeAppsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"apple_app_ids": types.ListType{ElemType: types.StringType},
		"android_apps":  types.ListType{ElemType: types.ObjectType{AttrTypes: androidAppAttrTypes()}},
	}
}

// TestMergeNativeAppsPatch covers the pure fold: the camelCase nativeApps shape
// is injected, OTHER auth_config keys are preserved, a raw-JSON nativeApps is
// overridden, and an empty/absent raw patch starts from {}.
func TestMergeNativeAppsPatch(t *testing.T) {
	native := map[string]any{
		"appleAppIds": []string{"LB4397Q8XJ.com.acme.app"},
		"androidApps": []map[string]any{
			{"packageName": "com.acme.app", "sha256CertFingerprints": []string{"AB:CD:EF"}},
		},
	}

	t.Run("preserves other keys and injects nativeApps", func(t *testing.T) {
		raw := json.RawMessage(`{"session":{"inactivityTimeout":1800000}}`)
		out, err := mergeNativeAppsPatch(raw, native)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var got map[string]any
		if err := json.Unmarshal(out, &got); err != nil {
			t.Fatalf("output not valid JSON: %v", err)
		}
		if _, ok := got["session"]; !ok {
			t.Errorf("other auth_config key `session` was dropped: %s", out)
		}
		na, ok := got["nativeApps"].(map[string]any)
		if !ok {
			t.Fatalf("nativeApps missing or not an object: %s", out)
		}
		ids, _ := na["appleAppIds"].([]any)
		if len(ids) != 1 || ids[0] != "LB4397Q8XJ.com.acme.app" {
			t.Errorf("appleAppIds wrong: %v", na["appleAppIds"])
		}
		apps, _ := na["androidApps"].([]any)
		if len(apps) != 1 {
			t.Fatalf("androidApps wrong: %v", na["androidApps"])
		}
		app := apps[0].(map[string]any)
		if app["packageName"] != "com.acme.app" {
			t.Errorf("packageName wrong: %v", app["packageName"])
		}
		if fps, _ := app["sha256CertFingerprints"].([]any); len(fps) != 1 || fps[0] != "AB:CD:EF" {
			t.Errorf("sha256CertFingerprints wrong: %v", app["sha256CertFingerprints"])
		}
	})

	t.Run("typed value overrides a raw-JSON nativeApps", func(t *testing.T) {
		raw := json.RawMessage(`{"nativeApps":{"appleAppIds":["OLD.com.old"]},"mfa":true}`)
		out, err := mergeNativeAppsPatch(raw, native)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var got map[string]any
		if err := json.Unmarshal(out, &got); err != nil {
			t.Fatalf("output not valid JSON: %v", err)
		}
		if got["mfa"] != true {
			t.Errorf("unrelated key `mfa` lost: %s", out)
		}
		na := got["nativeApps"].(map[string]any)
		ids := na["appleAppIds"].([]any)
		if len(ids) != 1 || ids[0] != "LB4397Q8XJ.com.acme.app" {
			t.Errorf("typed native_apps did not override raw nativeApps: %v", na["appleAppIds"])
		}
	})

	t.Run("nil/empty raw starts from {}", func(t *testing.T) {
		for _, raw := range []json.RawMessage{nil, json.RawMessage("null"), json.RawMessage("")} {
			out, err := mergeNativeAppsPatch(raw, native)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(out, &got); err != nil {
				t.Fatalf("output not valid JSON: %v", err)
			}
			if _, ok := got["nativeApps"]; !ok {
				t.Errorf("nativeApps missing for raw=%q: %s", raw, out)
			}
		}
	})
}

// TestNativeAppsToJSON proves the typed object decodes into the camelCase shape.
func TestNativeAppsToJSON(t *testing.T) {
	ctx := context.Background()
	appleIDs, _ := types.ListValueFrom(ctx, types.StringType, []string{"LB4397Q8XJ.com.acme.app"})
	fps, _ := types.ListValueFrom(ctx, types.StringType, []string{"AB:CD:EF"})
	androidObj, _ := types.ObjectValue(androidAppAttrTypes(), map[string]attr.Value{
		"package_name":             types.StringValue("com.acme.app"),
		"sha256_cert_fingerprints": fps,
	})
	androidList, _ := types.ListValue(types.ObjectType{AttrTypes: androidAppAttrTypes()}, []attr.Value{androidObj})
	obj, _ := types.ObjectValue(nativeAppsAttrTypes(), map[string]attr.Value{
		"apple_app_ids": appleIDs,
		"android_apps":  androidList,
	})

	var diags diag.Diagnostics
	got := nativeAppsToJSON(ctx, obj, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	want := map[string]any{
		"appleAppIds": []string{"LB4397Q8XJ.com.acme.app"},
		"androidApps": []map[string]any{
			{"packageName": "com.acme.app", "sha256CertFingerprints": []string{"AB:CD:EF"}},
		},
	}
	// Compare via canonical JSON so slice/map element types don't trip reflect.
	gj, _ := json.Marshal(got)
	wj, _ := json.Marshal(want)
	if !reflect.DeepEqual(gj, wj) {
		t.Errorf("nativeAppsToJSON = %s, want %s", gj, wj)
	}

	// A null object yields nil (nothing folded).
	if nativeAppsToJSON(ctx, types.ObjectNull(nativeAppsAttrTypes()), &diags) != nil {
		t.Errorf("null native_apps should decode to nil")
	}
}

// TestInstanceConfigMapToStateKeepsConfiguredAuthConfig proves the diff/apply
// fix: mapToState must preserve the user's partial `auth_config` verbatim (so a
// re-plan after apply is empty — no phantom diff and no "inconsistent result"),
// while the server-merged value is surfaced only through the computed
// `auth_config_resolved`.
func TestInstanceConfigMapToStateKeepsConfiguredAuthConfig(t *testing.T) {
	r := &instanceConfigResource{}

	// What the user wrote: a partial patch (a subset of the merged config).
	userWrote := `{"session":{"inactivityTimeout":1800000}}`
	m := &instanceConfigModel{AuthConfig: types.StringValue(userWrote)}

	// What the API returns: the full server-merged object (auth_config) plus a
	// fully-expanded effective view (auth_config_resolved).
	in := &client.Instance{
		ID:                 "ins_123",
		Environment:        "production",
		PublishableKey:     "pk_live_abc",
		FrontendAPIHost:    "fapi.example.com",
		AllowedOrigins:     []string{"https://app.example.com"},
		AuthConfig:         json.RawMessage(`{"session":{"inactivityTimeout":1800000},"mfa":false}`),
		AuthConfigResolved: json.RawMessage(`{"session":{"inactivityTimeout":1800000},"mfa":false,"passwordless":true}`),
		CreatedAt:          1700000000000,
	}

	var diags diag.Diagnostics
	r.mapToState(context.Background(), in, m, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// auth_config must be the user's input verbatim — NOT the merged object.
	if got := m.AuthConfig.ValueString(); got != userWrote {
		t.Errorf("auth_config was overwritten with the merged value: got %q, want %q", got, userWrote)
	}
	// auth_config_resolved must reflect the merged/effective config (canonicalised).
	wantResolved := `{"mfa":false,"passwordless":true,"session":{"inactivityTimeout":1800000}}`
	if got := m.AuthConfigResolved.ValueString(); got != wantResolved {
		t.Errorf("auth_config_resolved = %q, want %q", got, wantResolved)
	}
}

// TestAccInstanceConfig_basic sets the instance allowed_origins, verifies them,
// then imports the singleton. auth_config is left unmanaged here so the test
// does not mutate the instance's live auth behaviour.
func TestAccInstanceConfig_basic(t *testing.T) {
	origin := fmt.Sprintf("https://%s.example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigConfig(origin),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_instance_config.test", "allowed_origins.#", "1"),
					resource.TestCheckResourceAttr("atlas_instance_config.test", "allowed_origins.0", origin),
					resource.TestCheckResourceAttrSet("atlas_instance_config.test", "id"),
					resource.TestCheckResourceAttrSet("atlas_instance_config.test", "publishable_key"),
				),
			},
			{
				ResourceName:      "atlas_instance_config.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccInstanceConfigConfig(origin string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_instance_config" "test" {
  allowed_origins = [%[1]q]
}
`, origin)
}
