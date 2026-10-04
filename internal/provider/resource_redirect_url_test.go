package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRedirectURL_basic creates an allowlisted redirect URL, checks it, then
// imports it by id.
func TestAccRedirectURL_basic(t *testing.T) {
	url := fmt.Sprintf("https://app.%s.example.com/callback", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedirectURLConfig(url),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_redirect_url.test", "url", url),
					resource.TestCheckResourceAttrSet("atlas_redirect_url.test", "id"),
				),
			},
			{
				ResourceName:      "atlas_redirect_url.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccRedirectURL_disappears deletes the URL out of band and expects a
// non-empty follow-up plan (the provider re-plans its recreation).
func TestAccRedirectURL_disappears(t *testing.T) {
	url := fmt.Sprintf("https://app.%s.example.com/callback", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedirectURLConfig(url),
				Check: deleteOutOfBand("atlas_redirect_url.test", func(ctx context.Context, c *client.Client, id string) error {
					return c.DeleteRedirectURL(ctx, id)
				}),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRedirectURLConfig(url string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_redirect_url" "test" {
  url = %[1]q
}
`, url)
}
