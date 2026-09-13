package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccResourceServerDataSource_basic creates a resource server, then reads it
// back through the data source by id and asserts the projected fields match the
// managed resource.
func TestAccResourceServerDataSource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-rs-ds")
	identifier := fmt.Sprintf("https://%s.example.com", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServerDataSourceConfig(identifier, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.atlas_resource_server.test", "id",
						"atlas_resource_server.test", "id"),
					resource.TestCheckResourceAttrPair(
						"data.atlas_resource_server.test", "identifier",
						"atlas_resource_server.test", "identifier"),
					resource.TestCheckResourceAttrPair(
						"data.atlas_resource_server.test", "scopes.0.value",
						"atlas_resource_server.test", "scopes.0.value"),
					resource.TestCheckResourceAttr("data.atlas_resource_server.test", "identifier", identifier),
				),
			},
		},
	})
}

func testAccResourceServerDataSourceConfig(identifier, name string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_resource_server" "test" {
  identifier = %[1]q
  name       = %[2]q

  scopes = [
    {
      value       = "read:widgets"
      description = "Read widgets"
    },
  ]
}

data "atlas_resource_server" "test" {
  id = atlas_resource_server.test.id
}
`, identifier, name)
}
