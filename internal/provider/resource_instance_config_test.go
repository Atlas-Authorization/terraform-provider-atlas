package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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
