package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// Live acceptance tests (TF_ACC=1 + a disposable instance) for the policy
// acceptance resources. signing_secret_wo is a write-only argument, which needs
// Terraform 1.11 or later. The offline harness tests cover the same resources
// without a server.
func TestAccPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "atlas_policy" "tos" {
  key        = "tf-acc-tos"
  title      = "Terms of service"
  display    = "scroll"
  applies_to = "sign_in_and_sign_up"
  reaccept   = "grace"
  grace_days = 14
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_policy.tos", "key", "tf-acc-tos"),
					resource.TestCheckResourceAttr("atlas_policy.tos", "grace_days", "14"),
					resource.TestCheckResourceAttr("atlas_policy.tos", "status", "active"),
				),
			},
			{
				Config: providerConfig + `
resource "atlas_policy" "tos" {
  key      = "tf-acc-tos"
  title    = "Terms of service"
  required = false
}
`,
				Check: resource.TestCheckResourceAttr("atlas_policy.tos", "required", "false"),
			},
			{ResourceName: "atlas_policy.tos", ImportState: true, ImportStateId: "tf-acc-tos", ImportStateVerify: true},
		},
	})
}

func TestAccPolicySettings_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "atlas_policy_settings" "this" {
  session_claim = true
}
`,
				Check: resource.TestCheckResourceAttr("atlas_policy_settings.this", "session_claim", "true"),
			},
			{ResourceName: "atlas_policy_settings.this", ImportState: true, ImportStateId: "policy_settings", ImportStateVerify: true},
		},
	})
}

func TestAccPolicySource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "atlas_policy_source" "this" {
  enabled                   = false
  endpoint_url              = "https://example.com/atlas"
  signing_secret_wo         = "0123456789abcdef0123456789abcdef"
  signing_secret_wo_version = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_policy_source.this", "signing_secret_set", "true"),
					resource.TestCheckNoResourceAttr("atlas_policy_source.this", "signing_secret_wo"),
				),
			},
		},
	})
}
