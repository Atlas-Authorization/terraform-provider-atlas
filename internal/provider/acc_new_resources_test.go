package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Live acceptance tests (TF_ACC=1 + a disposable instance) for the 0.7.0
// resources that need only a secret key. The offline harness tests cover the
// same resources without a server.

func TestAccCookieConsent_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "atlas_cookie_consent" "test" {
  script_url = "https://consent.cookiebot.com/uc.js"
  script_id  = "Cookiebot"
  site_id    = "00000000-0000-0000-0000-000000000000"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_cookie_consent.test", "script_id", "Cookiebot"),
				),
			},
			{ResourceName: "atlas_cookie_consent.test", ImportState: true, ImportStateId: "cookie_consent", ImportStateVerify: true},
		},
	})
}

func TestAccInstanceDesign_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "atlas_instance_design" "test" {
  design = jsonencode({ preset = "midnight", palette = { primary = "#4f46e5" } })
}
`,
				Check: resource.TestCheckResourceAttrSet("atlas_instance_design.test", "resolved"),
			},
		},
	})
}

func TestAccInstancePreflightDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "atlas_instance_preflight" "this" {}`,
				Check:  resource.TestCheckResourceAttrSet("data.atlas_instance_preflight.this", "checked_at"),
			},
		},
	})
}
