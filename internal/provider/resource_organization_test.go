package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccOrgUserID resolves the existing user id required as an organization's
// created_by (its first admin). It has no config equivalent — the user must
// already exist in the target instance — so it is supplied out of band via
// ATLAS_TEST_USER_ID, and the test skips when that is absent.
func testAccOrgUserID(t *testing.T) string {
	t.Helper()
	id := os.Getenv("ATLAS_TEST_USER_ID")
	if id == "" {
		t.Skip("ATLAS_TEST_USER_ID must be set (an existing user id) to run the organization acceptance test")
	}
	return id
}

// TestAccOrganization_basic creates an org, checks its profile, then imports it.
func TestAccOrganization_basic(t *testing.T) {
	slug := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	name := "Acme " + slug

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig(name, slug, testAccOrgUserID(t), 50),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_organization.test", "name", name),
					resource.TestCheckResourceAttr("atlas_organization.test", "slug", slug),
					resource.TestCheckResourceAttr("atlas_organization.test", "max_allowed_memberships", "50"),
					resource.TestCheckResourceAttrSet("atlas_organization.test", "id"),
				),
			},
			{
				ResourceName:      "atlas_organization.test",
				ImportState:       true,
				ImportStateVerify: true,
				// created_by is not returned by the profile projection (the
				// provider carries it forward from prior state), so an import
				// cannot repopulate it and it is excluded from the round-trip.
				ImportStateVerifyIgnore: []string{"created_by"},
			},
		},
	})
}

// TestAccOrganization_update changes the name and seat cap in place (created_by
// is RequiresReplace and is left unchanged).
func TestAccOrganization_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	userID := testAccOrgUserID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig("Acme "+slug, slug, userID, 50),
				Check:  resource.TestCheckResourceAttr("atlas_organization.test", "max_allowed_memberships", "50"),
			},
			{
				Config: testAccOrganizationConfig("Acme Renamed "+slug, slug, userID, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_organization.test", "name", "Acme Renamed "+slug),
					resource.TestCheckResourceAttr("atlas_organization.test", "max_allowed_memberships", "100"),
				),
			},
		},
	})
}

func testAccOrganizationConfig(name, slug, createdBy string, maxMemberships int) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_organization" "test" {
  name                    = %[1]q
  slug                    = %[2]q
  created_by              = %[3]q
  max_allowed_memberships = %[4]d
}
`, name, slug, createdBy, maxMemberships)
}
