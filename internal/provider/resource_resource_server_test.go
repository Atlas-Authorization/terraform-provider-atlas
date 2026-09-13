package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccResourceServer_basic creates an API (resource server), checks the
// audience/name/scopes, then imports it. The resource has no write-only
// secrets, so the import round-trip verifies every attribute.
func TestAccResourceServer_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-rs")
	identifier := fmt.Sprintf("https://%s.example.com", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServerConfig(identifier, name, 3600),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_resource_server.test", "identifier", identifier),
					resource.TestCheckResourceAttr("atlas_resource_server.test", "name", name),
					resource.TestCheckResourceAttr("atlas_resource_server.test", "token_ttl_seconds", "3600"),
					resource.TestCheckResourceAttr("atlas_resource_server.test", "scopes.#", "1"),
					resource.TestCheckResourceAttr("atlas_resource_server.test", "scopes.0.value", "read:widgets"),
					resource.TestCheckResourceAttrSet("atlas_resource_server.test", "id"),
					resource.TestCheckResourceAttrSet("atlas_resource_server.test", "signing_alg"),
				),
			},
			{
				ResourceName:      "atlas_resource_server.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceServer_update changes the name and token TTL in place
// (identifier is RequiresReplace and is left unchanged).
func TestAccResourceServer_update(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-rs")
	identifier := fmt.Sprintf("https://%s.example.com", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServerConfig(identifier, name, 3600),
				Check:  resource.TestCheckResourceAttr("atlas_resource_server.test", "token_ttl_seconds", "3600"),
			},
			{
				Config: testAccResourceServerConfig(identifier, name+"-v2", 7200),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_resource_server.test", "name", name+"-v2"),
					resource.TestCheckResourceAttr("atlas_resource_server.test", "token_ttl_seconds", "7200"),
					// identifier is immutable — it must survive an in-place update.
					resource.TestCheckResourceAttr("atlas_resource_server.test", "identifier", identifier),
				),
			},
		},
	})
}

func testAccResourceServerConfig(identifier, name string, ttl int) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_resource_server" "test" {
  identifier        = %[1]q
  name              = %[2]q
  token_ttl_seconds = %[3]d
  signing_alg       = "RS256"

  scopes = [
    {
      value       = "read:widgets"
      description = "Read widgets"
    },
  ]
}
`, identifier, name, ttl)
}
