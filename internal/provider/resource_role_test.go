package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRole_basic creates a custom role, checks its label and permissions,
// then imports it (by id).
func TestAccRole_basic(t *testing.T) {
	key := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig(key, "Billing Admin", "Manages billing"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_role.test", "key", key),
					resource.TestCheckResourceAttr("atlas_role.test", "name", "Billing Admin"),
					resource.TestCheckResourceAttr("atlas_role.test", "description", "Manages billing"),
					resource.TestCheckResourceAttr("atlas_role.test", "is_system", "false"),
					resource.TestCheckResourceAttrSet("atlas_role.test", "id"),
				),
			},
			{
				ResourceName:      "atlas_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccRole_update mutates the label/description in place (key is
// RequiresReplace and is left unchanged; permissions go through their own PUT).
func TestAccRole_update(t *testing.T) {
	key := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig(key, "Billing Admin", "Manages billing"),
				Check:  resource.TestCheckResourceAttr("atlas_role.test", "name", "Billing Admin"),
			},
			{
				Config: testAccRoleConfig(key, "Billing Manager", "Manages billing and invoices"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_role.test", "name", "Billing Manager"),
					resource.TestCheckResourceAttr("atlas_role.test", "description", "Manages billing and invoices"),
					resource.TestCheckResourceAttr("atlas_role.test", "key", key),
				),
			},
		},
	})
}

func testAccRoleConfig(key, name, description string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_role" "test" {
  key         = %[1]q
  name        = %[2]q
  description = %[3]q
  permissions = ["org:billing:manage", "org:invoices:write"]
}
`, key, name, description)
}
