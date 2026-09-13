package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDomain_basic creates a custom domain, checks its fields (including the
// computed cname_target), then imports it. The domain has no secrets, so the
// import round-trip verifies every attribute.
func TestAccDomain_basic(t *testing.T) {
	host := fmt.Sprintf("auth-%s.example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig("fapi", host),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_domain.test", "role", "fapi"),
					resource.TestCheckResourceAttr("atlas_domain.test", "host", host),
					resource.TestCheckResourceAttrSet("atlas_domain.test", "id"),
					resource.TestCheckResourceAttrSet("atlas_domain.test", "cname_target"),
					resource.TestCheckResourceAttrSet("atlas_domain.test", "status"),
				),
			},
			{
				ResourceName:      "atlas_domain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDomain_replace changes host (RequiresReplace): the Backend API has no
// update route, so the domain is destroyed and recreated. The check asserts the
// id changed, proving a replacement rather than an in-place update.
func TestAccDomain_replace(t *testing.T) {
	host1 := fmt.Sprintf("auth-%s.example.com", acctest.RandString(8))
	host2 := fmt.Sprintf("auth-%s.example.com", acctest.RandString(8))
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig("fapi", host1),
				Check:  testAccCaptureAttr("atlas_domain.test", "id", &firstID),
			},
			{
				Config: testAccDomainConfig("fapi", host2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_domain.test", "host", host2),
					testAccCheckAttrChanged("atlas_domain.test", "id", &firstID),
				),
			},
		},
	})
}

// TestAccDomain_disappears deletes the domain out of band and expects the next
// plan to be non-empty (the provider re-plans its recreation).
func TestAccDomain_disappears(t *testing.T) {
	host := fmt.Sprintf("auth-%s.example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig("fapi", host),
				Check: deleteOutOfBand("atlas_domain.test", func(ctx context.Context, c *client.Client, id string) error {
					return c.DeleteDomain(ctx, id)
				}),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomainConfig(role, host string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_domain" "test" {
  role = %[1]q
  host = %[2]q
}
`, role, host)
}
