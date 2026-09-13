package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccJwtTemplate_basic creates a template, checks its claims, then imports
// it. The template is imported by name (which is also its id).
func TestAccJwtTemplate_basic(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJwtTemplateConfig(name, "authenticated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_jwt_template.test", "name", name),
					resource.TestCheckResourceAttr("atlas_jwt_template.test", "id", name),
					resource.TestCheckResourceAttr("atlas_jwt_template.test", "claims.role", "authenticated"),
				),
			},
			{
				ResourceName:      "atlas_jwt_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccJwtTemplate_update mutates the claims map in place (name is
// RequiresReplace and is left unchanged).
func TestAccJwtTemplate_update(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJwtTemplateConfig(name, "authenticated"),
				Check:  resource.TestCheckResourceAttr("atlas_jwt_template.test", "claims.role", "authenticated"),
			},
			{
				Config: testAccJwtTemplateConfig(name, "service_role"),
				Check:  resource.TestCheckResourceAttr("atlas_jwt_template.test", "claims.role", "service_role"),
			},
		},
	})
}

func testAccJwtTemplateConfig(name, role string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_jwt_template" "test" {
  name = %[1]q
  claims = {
    role      = %[2]q
    user_role = "{{user.public_metadata.role}}"
  }
}
`, name, role)
}
