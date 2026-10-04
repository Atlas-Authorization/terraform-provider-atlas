package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccBillingPlan_basic creates a free-tier plan (no stripe_price_id so the
// test needs no Stripe price), checks its fields, then imports it by id.
func TestAccBillingPlan_basic(t *testing.T) {
	slug := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingPlanConfig(slug, "Free", `["dashboard"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_billing_plan.test", "slug", slug),
					resource.TestCheckResourceAttr("atlas_billing_plan.test", "name", "Free"),
					resource.TestCheckResourceAttr("atlas_billing_plan.test", "free", "true"),
					resource.TestCheckResourceAttr("atlas_billing_plan.test", "audience", "user"),
					resource.TestCheckResourceAttr("atlas_billing_plan.test", "features.#", "1"),
					resource.TestCheckResourceAttrSet("atlas_billing_plan.test", "id"),
				),
			},
			{
				ResourceName:      "atlas_billing_plan.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccBillingPlan_update changes the name and feature set in place.
func TestAccBillingPlan_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingPlanConfig(slug, "Starter", `["dashboard"]`),
				Check:  resource.TestCheckResourceAttr("atlas_billing_plan.test", "name", "Starter"),
			},
			{
				Config: testAccBillingPlanConfig(slug, "Starter Plus", `["dashboard", "sso"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_billing_plan.test", "name", "Starter Plus"),
					resource.TestCheckResourceAttr("atlas_billing_plan.test", "features.#", "2"),
					resource.TestCheckResourceAttr("atlas_billing_plan.test", "slug", slug),
				),
			},
		},
	})
}

func testAccBillingPlanConfig(slug, name, features string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_billing_plan" "test" {
  name     = %[1]q
  slug     = %[2]q
  audience = "user"
  features = %[3]s
}
`, name, slug, features)
}
