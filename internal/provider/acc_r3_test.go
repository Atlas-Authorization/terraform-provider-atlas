package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// Live acceptance tests (TF_ACC=1 against a sandbox instance) for the 0.13.0
// additions: atlas_partner_platform (write-only token) and the per-call
// audience on atlas_integration. Write-only (_wo) attributes need Terraform 1.11+.

func TestAccPartnerPlatform_lifecycle(t *testing.T) {
	cfg := func(extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_partner_platform" "test" {
  partner          = "tfacc-consent"
  name             = "TF acceptance consent partner"
  platform_id      = "atlas"
  api_base_url     = "https://partner.example.com"
  jwks_url         = "https://example.com/.well-known/tfacc-jwks.json"
  signature_header = "x-tfacc-signature"
  allowed_sources  = ["tfacc-site-1"]
  %s
}
`, extra)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_partner_platform.test", "token_set", "false"),
					resource.TestCheckResourceAttr("atlas_partner_platform.test", "registration_status", "unregistered"),
					resource.TestCheckResourceAttrSet("atlas_partner_platform.test", "receiver_url"),
				),
			},
			{
				// Storing the token is an in-place update; the token never reaches state.
				Config: cfg(`
  token_wo         = "tfacc-not-a-real-token-0123456789"
  token_wo_version = 1
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_partner_platform.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_partner_platform.test", "token_set", "true"),
					resource.TestCheckNoResourceAttr("atlas_partner_platform.test", "token_wo"),
				),
			},
			{
				// A new version rotates the stored token.
				Config: cfg(`
  token_wo         = "tfacc-another-fake-token-9876543210"
  token_wo_version = 2
`),
				Check: resource.TestCheckResourceAttr("atlas_partner_platform.test", "token_set", "true"),
			},
			{
				ResourceName:            "atlas_partner_platform.test",
				ImportState:             true,
				ImportStateId:           "tfacc-consent",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token_wo", "token_wo_version", "api_base_url"},
			},
		},
	})
}

func TestAccIntegration_perCallAudience(t *testing.T) {
	cfg := func(block string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_integration" "test" {
  integration = "tfacc-percall"
  name        = "TF acceptance per-call audience"
  audience    = "tfacc:main"
  %s
}
data "atlas_integration" "test" { integration = atlas_integration.test.integration }
`, block)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(`
  per_call_audience {
    param    = "tenant"
    template = "tfacc:{tenant}"
    allowed  = ["main", "eu"]
  }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_integration.test", "per_call_audience.param", "tenant"),
					resource.TestCheckResourceAttr("atlas_integration.test", "per_call_audience.allowed.#", "2"),
					resource.TestCheckResourceAttr("data.atlas_integration.test", "per_call_audience.template", "tfacc:{tenant}"),
				),
			},
			{
				// Removing the block turns it off for a custom integration.
				Config: cfg(""),
				Check:  resource.TestCheckNoResourceAttr("atlas_integration.test", "per_call_audience.param"),
			},
		},
	})
}
