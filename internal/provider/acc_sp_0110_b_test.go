package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Live acceptance tests (TF_ACC=1) for atlas_integration (generic and presets),
// atlas_privacy_orchestrator and atlas_referral_settings. Receivers are always
// example.com; the referral settings never point at a real provider endpoint.

func accCheckIntegrationGone(ids ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		for _, id := range ids {
			i, err := testAccClient().GetIntegration(context.Background(), id)
			if err == nil && i.Connected {
				return fmt.Errorf("integration %s still connected after destroy", id)
			}
		}
		return nil
	}
}

func accCheckOrchestratorGone(_ *terraform.State) error {
	o, err := testAccClient().GetPrivacyOrchestrator(context.Background())
	if err == nil && o.Configured {
		return fmt.Errorf("privacy orchestrator still configured after destroy")
	}
	return nil
}

func accCheckReferralsOff(_ *terraform.State) error {
	r, err := testAccClient().GetReferralSettings(context.Background())
	if err != nil {
		return err
	}
	if r.Enabled {
		return fmt.Errorf("referral capture still enabled after destroy")
	}
	return nil
}

func TestAccIntegration_custom_lifecycle(t *testing.T) {
	cfg := func(name, extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_integration" "test" {
  integration     = "tfacc-custom"
  name            = %q
  role            = "other"
  audience        = "tfacc:one"
  audience_format = "^tfacc:[a-z0-9]+$"
  events_url      = "https://example.com/atlas/events"
  external_ref    = "tfacc-custom-integration"
  %s
}
data "atlas_integration" "test" { integration = atlas_integration.test.id }
`, name, extra)
	}
	resource.Test(t, resource.TestCase{
		CheckDestroy:             accCheckIntegrationGone("tfacc-custom"),
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("TF acc custom", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_integration.test", "id", "tfacc-custom"),
					resource.TestCheckResourceAttr("atlas_integration.test", "enabled", "true"),
					resource.TestCheckResourceAttr("atlas_integration.test", "audience", "tfacc:one"),
					resource.TestCheckResourceAttrSet("atlas_integration.test", "issuer"),
					resource.TestCheckResourceAttrSet("atlas_integration.test", "jwks_url"),
					resource.TestCheckResourceAttrPair("data.atlas_integration.test", "audience", "atlas_integration.test", "audience"),
				),
			},
			{
				Config: cfg("TF acc custom renamed", `
  ttl_seconds = 60
  metadata    = { team = "acc" }
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_integration.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_integration.test", "name", "TF acc custom renamed"),
					resource.TestCheckResourceAttr("atlas_integration.test", "ttl_seconds", "60"),
					resource.TestCheckResourceAttr("atlas_integration.test", "metadata.team", "acc"),
				),
			},
			{
				ResourceName:            "atlas_integration.test",
				ImportState:             true,
				ImportStateId:           "tfacc-custom",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"audience_format"},
			},
		},
	})
}

func accPresetIntegrationTest(t *testing.T, preset, resType, aud string) {
	cfg := func(extra string) string {
		return providerConfig + fmt.Sprintf(`
resource %[1]q "test" {
  audience   = %[2]q
  events_url = "https://example.com/atlas/%[3]s/events"
  %[4]s
}
data "atlas_integration" "test" { integration = %[1]s.test.id }
`, resType, aud, preset, extra)
	}
	addr := resType + ".test"
	resource.Test(t, resource.TestCase{
		CheckDestroy:             accCheckIntegrationGone(preset),
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", preset),
					resource.TestCheckResourceAttr(addr, "integration", preset),
					resource.TestCheckResourceAttr(addr, "preset", preset),
					resource.TestCheckResourceAttr(addr, "enabled", "true"),
					resource.TestCheckResourceAttrSet(addr, "name"),
					resource.TestCheckResourceAttrSet(addr, "role"),
					resource.TestCheckResourceAttrSet(addr, "issuer"),
					resource.TestCheckResourceAttr("data.atlas_integration.test", "audience", aud),
				),
			},
			{
				Config: cfg(`
  ttl_seconds = 90
  metadata    = { team = "acc" }
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "ttl_seconds", "90"),
					resource.TestCheckResourceAttr(addr, "metadata.team", "acc"),
				),
			},
			{
				ResourceName:      addr,
				ImportState:       true,
				ImportStateId:     preset,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccChatlyIntegration_lifecycle(t *testing.T) {
	accPresetIntegrationTest(t, "chatly", "atlas_chatly_integration", "chatly:tfacc")
}
func TestAccForumulateIntegration_lifecycle(t *testing.T) {
	accPresetIntegrationTest(t, "forumulate", "atlas_forumulate_integration", "forumulate:org_tfacc")
}
func TestAccCookiemunchIntegration_lifecycle(t *testing.T) {
	accPresetIntegrationTest(t, "cookiemunch", "atlas_cookiemunch_integration", "cookiemunch:dv-tfacc")
}
func TestAccReferalliIntegration_lifecycle(t *testing.T) {
	accPresetIntegrationTest(t, "referalli", "atlas_referalli_integration", "referalli:org_tfacc")
}

// The generic resource also accepts a preset id.
func TestAccIntegration_generic_preset(t *testing.T) {
	cfg := providerConfig + `
resource "atlas_integration" "test" {
  integration = "chatly"
  audience    = "chatly:tfacc2"
  events_url  = "https://example.com/atlas/chatly/events"
}
`
	resource.Test(t, resource.TestCase{
		CheckDestroy:             accCheckIntegrationGone("chatly"),
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: cfg, Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("atlas_integration.test", "preset", "chatly"),
				resource.TestCheckResourceAttrSet("atlas_integration.test", "name"),
			)},
			{ResourceName: "atlas_integration.test", ImportState: true, ImportStateId: "chatly", ImportStateVerify: true, ImportStateVerifyIgnore: []string{"audience_format"}},
		},
	})
}

func TestAccPrivacyOrchestrator_lifecycle(t *testing.T) {
	cfg := func(extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_privacy_orchestrator" "test" {
  provider_id = "tfacc"
  name        = "TF acc orchestrator"
  case_url    = "https://example.com/dsar/cases"
  cancel_url  = "https://example.com/dsar/cases/{id}/cancel"
  %s
}
data "atlas_privacy_orchestrator" "test" { depends_on = [atlas_privacy_orchestrator.test] }
`, extra)
	}
	resource.Test(t, resource.TestCase{
		CheckDestroy:             accCheckOrchestratorGone,
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(`
  api_key_wo         = "tfacc-not-a-real-key-0123456789"
  api_key_wo_version = 1
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_privacy_orchestrator.test", "id", "privacy_orchestrator"),
					resource.TestCheckResourceAttr("atlas_privacy_orchestrator.test", "provider_id", "tfacc"),
					resource.TestCheckResourceAttr("atlas_privacy_orchestrator.test", "enabled", "true"),
					resource.TestCheckResourceAttr("atlas_privacy_orchestrator.test", "api_key_set", "true"),
					resource.TestCheckResourceAttr("atlas_privacy_orchestrator.test", "auth_scheme", "bearer"),
					resource.TestCheckResourceAttr("data.atlas_privacy_orchestrator.test", "case_url", "https://example.com/dsar/cases"),
					resource.TestCheckNoResourceAttr("atlas_privacy_orchestrator.test", "api_key_wo"),
				),
			},
			{
				Config: cfg(`
  api_key_wo         = "tfacc-not-a-real-key-0123456789"
  api_key_wo_version = 1
  sign_requests      = true
  completion_webhook_url = "https://example.com/dsar/completed"
  open_closure_cases = false
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_privacy_orchestrator.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_privacy_orchestrator.test", "sign_requests", "true"),
					resource.TestCheckResourceAttrSet("atlas_privacy_orchestrator.test", "signing_secret"),
					resource.TestCheckResourceAttrSet("atlas_privacy_orchestrator.test", "completion_webhook_secret"),
					resource.TestCheckResourceAttr("atlas_privacy_orchestrator.test", "open_closure_cases", "false"),
				),
			},
			{
				// Bump the key version: key re-sent, secrets retained.
				Config: cfg(`
  api_key_wo         = "tfacc-not-a-real-key-9876543210"
  api_key_wo_version = 2
  sign_requests      = true
  completion_webhook_url = "https://example.com/dsar/completed"
  open_closure_cases = false
`),
				Check: resource.TestCheckResourceAttrSet("atlas_privacy_orchestrator.test", "signing_secret"),
			},
			{
				ResourceName:            "atlas_privacy_orchestrator.test",
				ImportState:             true,
				ImportStateId:           "privacy_orchestrator",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key_wo_version", "signing_secret", "completion_webhook_secret"},
			},
		},
	})
}

func TestAccReferralSettings_lifecycle(t *testing.T) {
	cfg := func(extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_referral_settings" "test" {
  provider_id  = "webhook"
  endpoint_url = "https://example.com/referrals/hook"
  program_id   = "tfacc-program"
  %s
}
data "atlas_referral_settings" "test" { depends_on = [atlas_referral_settings.test] }
`, extra)
	}
	resource.Test(t, resource.TestCase{
		CheckDestroy:             accCheckReferralsOff,
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "id", "referral_settings"),
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "enabled", "false"),
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "endpoint_url", "https://example.com/referrals/hook"),
					resource.TestCheckResourceAttr("data.atlas_referral_settings.test", "program_id", "tfacc-program"),
				),
			},
			{
				Config: cfg(`
  signing_secret_wo         = "tfacc-not-a-real-secret-0123456789"
  signing_secret_wo_version = 1
  context_cookies           = ["tfacc_ref", "tfacc_ref2"]
  url_params                = ["tfref", "ref"]
  browser_sources           = ["referalli"]
  consent_provider          = "gpc"
  on_high_risk              = "strip"
  hold_until_email_verified = true
  max_referred_signups_per_ip_hour = 9
  mirror_to_public_metadata = true
  link_template             = "https://example.com/?ref={code}"
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_referral_settings.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "signing_secret_set", "true"),
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "context_cookies.#", "2"),
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "url_params.0", "tfref"),
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "on_high_risk", "strip"),
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "max_referred_signups_per_ip_hour", "9"),
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "hold_until_email_verified", "true"),
				),
			},
			{
				// Enable capture: needs endpoint + signing secret, both supplied.
				Config: cfg(`
  enabled                   = true
  signing_secret_wo         = "tfacc-not-a-real-secret-0123456789"
  signing_secret_wo_version = 1
  context_cookies           = ["tfacc_ref", "tfacc_ref2"]
  url_params                = ["tfref", "ref"]
  browser_sources           = ["referalli"]
  consent_provider          = "gpc"
  on_high_risk              = "strip"
  hold_until_email_verified = true
  max_referred_signups_per_ip_hour = 9
  mirror_to_public_metadata = true
  link_template             = "https://example.com/?ref={code}"
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("atlas_referral_settings.test", "missing.#", "0"),
				),
			},
			{
				ResourceName:            "atlas_referral_settings.test",
				ImportState:             true,
				ImportStateId:           "referral_settings",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"signing_secret_wo_version", "api_key_wo_version"},
			},
		},
	})
}

// A preset orchestrator fills its own defaults; the plan must stay clean.
func TestAccPrivacyOrchestrator_preset(t *testing.T) {
	cfg := providerConfig + `
resource "atlas_privacy_orchestrator" "test" {
  provider_id = "cookiemunch"
  enabled     = false
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             accCheckOrchestratorGone,
		Steps: []resource.TestStep{
			{Config: cfg, Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("atlas_privacy_orchestrator.test", "provider_id", "cookiemunch"),
				resource.TestCheckResourceAttrSet("atlas_privacy_orchestrator.test", "case_url"),
				resource.TestCheckResourceAttrSet("atlas_privacy_orchestrator.test", "name"),
			)},
		},
	})
}
