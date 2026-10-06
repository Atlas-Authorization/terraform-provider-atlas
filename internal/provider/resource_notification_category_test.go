package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccNotificationCategory_basic creates a tenant notification category,
// checks its fields, then imports it by key. `optional` is always true.
func TestAccNotificationCategory_basic(t *testing.T) {
	key := "tf_" + strings.ToLower(acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationCategoryConfig(key, "Marketing updates"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_notification_category.test", "key", key),
					resource.TestCheckResourceAttr("atlas_notification_category.test", "label", "Marketing updates"),
					resource.TestCheckResourceAttr("atlas_notification_category.test", "optional", "true"),
					resource.TestCheckResourceAttrSet("atlas_notification_category.test", "id"),
				),
			},
			{
				ResourceName:      "atlas_notification_category.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Import id is the category key, not the object id.
				ImportStateId: key,
			},
		},
	})
}

// TestAccNotificationCategory_updateLabel changes the label in place: the key
// is immutable, so the id is unchanged (no replacement).
func TestAccNotificationCategory_updateLabel(t *testing.T) {
	key := "tf_" + strings.ToLower(acctest.RandString(8))
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationCategoryConfig(key, "First label"),
				Check:  testAccCaptureAttr("atlas_notification_category.test", "id", &firstID),
			},
			{
				Config: testAccNotificationCategoryConfig(key, "Second label"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_notification_category.test", "label", "Second label"),
					testAccCheckAttrUnchanged("atlas_notification_category.test", "id", &firstID),
				),
			},
		},
	})
}

// TestAccNotificationCategory_disappears deletes the category out of band and
// expects the next plan to be non-empty (the provider re-plans its recreation).
func TestAccNotificationCategory_disappears(t *testing.T) {
	key := "tf_" + strings.ToLower(acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationCategoryConfig(key, "Will vanish"),
				Check: deleteOutOfBand("atlas_notification_category.test", func(ctx context.Context, c *client.Client, _ string) error {
					return c.DeleteNotificationCategory(ctx, key)
				}),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccNotificationCategoryConfig(key, label string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_notification_category" "test" {
  key   = %[1]q
  label = %[2]q
}
`, key, label)
}
