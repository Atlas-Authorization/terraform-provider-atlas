package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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
