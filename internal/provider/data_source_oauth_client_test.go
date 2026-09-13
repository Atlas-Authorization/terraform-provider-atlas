package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccOAuthClientDataSource_basic creates an OAuth client, then reads it back
// through the data source by id and asserts the projected fields match the
// managed resource.
func TestAccOAuthClientDataSource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-oauth-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOAuthClientDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The data source's projection matches the resource it reads.
					resource.TestCheckResourceAttrPair(
						"data.atlas_oauth_client.test", "id",
						"atlas_oauth_client.test", "id"),
					resource.TestCheckResourceAttrPair(
						"data.atlas_oauth_client.test", "name",
						"atlas_oauth_client.test", "name"),
					resource.TestCheckResourceAttrPair(
						"data.atlas_oauth_client.test", "client_id",
						"atlas_oauth_client.test", "client_id"),
					resource.TestCheckResourceAttr("data.atlas_oauth_client.test", "name", name),
				),
			},
		},
	})
}

func testAccOAuthClientDataSourceConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_oauth_client" "test" {
  name          = %[1]q
  redirect_uris = ["https://app.example.com/callback"]
}

data "atlas_oauth_client" "test" {
  id = atlas_oauth_client.test.id
}
`, name)
}
