package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// Live acceptance test (TF_ACC=1 + a disposable instance) for
// atlas_better_auth_bridge. secret_wo is a write-only argument, which needs
// Terraform 1.11 or later. The offline harness tests cover the same resource
// without a server.
func TestAccBetterAuthBridge_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "atlas_better_auth_bridge" "test" {
  enabled       = true
  ends_at       = "2099-01-01T00:00:00Z"
  cookie_prefix = "better-auth"
  live_fallback = {
    enabled           = true
    base_url          = "https://example.com/api/auth"
    secret_wo         = "0123456789abcdef0123456789abcdef"
    secret_wo_version = 1
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_better_auth_bridge.test", "enabled", "true"),
					resource.TestCheckResourceAttr("atlas_better_auth_bridge.test", "live_fallback.has_secret", "true"),
					resource.TestCheckNoResourceAttr("atlas_better_auth_bridge.test", "live_fallback.secret_wo"),
				),
			},
			{
				Config: providerConfig + `
resource "atlas_better_auth_bridge" "test" {
  enabled = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_better_auth_bridge.test", "enabled", "false"),
					resource.TestCheckNoResourceAttr("atlas_better_auth_bridge.test", "live_fallback.enabled"),
				),
			},
			{
				ResourceName: "atlas_better_auth_bridge.test", ImportState: true, ImportStateId: "better_auth_bridge",
				ImportStateVerify: true,
			},
		},
	})
}
