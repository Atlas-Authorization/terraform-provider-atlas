package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRole_basic creates a custom role, checks its label and permissions,
// then imports it (by id).
func TestAccRole_basic(t *testing.T) {
	key := "org:" + acctest.RandStringFromCharSet(10, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccEnsureRolePermissions(t) },
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
	key := "org:" + acctest.RandStringFromCharSet(10, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccEnsureRolePermissions(t) },
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

// testAccEnsureRolePermissions defines the permissions the role configs grant.
// The API drops undefined permission keys, so they must exist first; an
// already-defined key (409) is fine.
func testAccEnsureRolePermissions(t *testing.T) {
	t.Helper()
	for _, key := range []string{"org:billing:manage", "org:invoices:write"} {
		err := testAccClient().CreatePermission(context.Background(), key, key)
		var apiErr *client.APIError
		if err != nil && !(errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict) {
			t.Fatalf("defining permission %s: %v", key, err)
		}
	}
}
