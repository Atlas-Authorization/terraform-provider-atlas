package provider

import (
	"context"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestMapOAuthProviderToState verifies the API projection lands on the model and
// that write-only credentials/settings (never returned by the API) are preserved
// across a read rather than being blanked.
func TestMapOAuthProviderToState(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Seed the model with the write-only fields a prior apply would have set.
	creds, d := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"client_id":     "cid-123",
		"client_secret": "shh",
	})
	diags.Append(d...)
	settings, d := types.MapValueFrom(ctx, types.StringType, map[string]string{"team": "acme"})
	diags.Append(d...)
	if diags.HasError() {
		t.Fatalf("setup diagnostics: %v", diags)
	}
	m := oauthProviderModel{Credentials: creds, Settings: settings}

	updatedAt := int64(1700000000000)
	clientID := "cid-123"
	api := &client.OAuthProvider{
		Provider:    "google",
		DisplayName: "Google",
		RedirectURI: "https://acme.atlas.dev/v1/oauth_callbacks/google",
		Configured:  true,
		ClientID:    &clientID,
		HasSecret:   true,
		Enabled:     true,
		AllowSignIn: true,
		AllowSignUp: false,
		Scopes:      []string{"openid", "email"},
		UpdatedAt:   &updatedAt,
	}

	mapOAuthProviderToState(ctx, api, &m, &diags)
	if diags.HasError() {
		t.Fatalf("mapping diagnostics: %v", diags)
	}

	if m.Provider.ValueString() != "google" || m.ID.ValueString() != "google" {
		t.Errorf("provider/id = %q/%q, want google/google", m.Provider.ValueString(), m.ID.ValueString())
	}
	if m.ClientID.ValueString() != "cid-123" || !m.HasSecret.ValueBool() || !m.Configured.ValueBool() {
		t.Errorf("readable fields mismatch: %+v", m)
	}
	if m.RedirectURI.ValueString() != api.RedirectURI || m.DisplayName.ValueString() != "Google" {
		t.Errorf("redirect_uri/display_name mismatch: %+v", m)
	}
	if !m.Enabled.ValueBool() || !m.AllowSignIn.ValueBool() || m.AllowSignUp.ValueBool() {
		t.Errorf("enabled/scope flags mismatch: %+v", m)
	}
	if m.UpdatedAt.ValueInt64() != updatedAt {
		t.Errorf("updated_at = %d, want %d", m.UpdatedAt.ValueInt64(), updatedAt)
	}
	// Write-only fields must survive the read untouched.
	var gotCreds map[string]string
	m.Credentials.ElementsAs(ctx, &gotCreds, false)
	if gotCreds["client_secret"] != "shh" || gotCreds["client_id"] != "cid-123" {
		t.Errorf("credentials not preserved across read: %v", gotCreds)
	}
	var gotSettings map[string]string
	m.Settings.ElementsAs(ctx, &gotSettings, false)
	if gotSettings["team"] != "acme" {
		t.Errorf("settings not preserved across read: %v", gotSettings)
	}
}

// TestBoolOrDefault covers the Optional+Computed resolution: an unset value falls
// back to the default, an explicit value wins.
func TestBoolOrDefault(t *testing.T) {
	cases := []struct {
		name string
		in   types.Bool
		def  bool
		want bool
	}{
		{"null falls back to default true", types.BoolNull(), true, true},
		{"null falls back to default false", types.BoolNull(), false, false},
		{"unknown falls back to default", types.BoolUnknown(), true, true},
		{"explicit true wins over default false", types.BoolValue(true), false, true},
		{"explicit false wins over default true", types.BoolValue(false), true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := boolOrDefault(tc.in, tc.def); got != tc.want {
				t.Errorf("boolOrDefault(%v, %v) = %v, want %v", tc.in, tc.def, got, tc.want)
			}
		})
	}
}

// TestAccOAuthProvider_basic configures the GitHub provider, checks its readable
// fields, then imports it. The write-only credentials are never returned by the
// API, so they are excluded from the import round-trip. Self-skips without TF_ACC.
func TestAccOAuthProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOAuthProviderConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_oauth_provider.github", "provider_key", "github"),
					resource.TestCheckResourceAttr("atlas_oauth_provider.github", "configured", "true"),
					resource.TestCheckResourceAttr("atlas_oauth_provider.github", "has_secret", "true"),
					resource.TestCheckResourceAttr("atlas_oauth_provider.github", "enabled", "true"),
					resource.TestCheckResourceAttrSet("atlas_oauth_provider.github", "redirect_uri"),
				),
			},
			{
				ResourceName:            "atlas_oauth_provider.github",
				ImportState:             true,
				ImportStateId:           "github",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials", "settings"},
			},
		},
	})
}

func testAccOAuthProviderConfig(enabled bool) string {
	return providerConfig + `
resource "atlas_oauth_provider" "github" {
  provider_key = "github"
  credentials = {
    client_id     = "gh-client-id"
    client_secret = "gh-client-secret"
  }
  scopes        = ["read:user", "user:email"]
  enabled       = ` + boolLit(enabled) + `
  allow_sign_in = true
  allow_sign_up = true
}
`
}

func boolLit(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
