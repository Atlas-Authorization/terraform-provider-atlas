package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccWebhookEndpoint_basic creates an endpoint, checks its fields, then
// imports it. The signing secret is revealed only on create and is never
// re-read, so it is excluded from the import round-trip.
func TestAccWebhookEndpoint_basic(t *testing.T) {
	url := fmt.Sprintf("https://hooks.%s.example.com/atlas", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookEndpointConfig(url, `["user.created", "user.updated"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "url", url),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "enabled_events.#", "2"),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "active", "true"),
					resource.TestCheckResourceAttrSet("atlas_webhook_endpoint.test", "id"),
					resource.TestCheckResourceAttrSet("atlas_webhook_endpoint.test", "secret"),
				),
			},
			{
				ResourceName:      "atlas_webhook_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The signing secret is one-time (create only); a list-based read
				// never re-reveals it, so it cannot match on import.
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

// TestAccWebhookEndpoint_replace changes url (RequiresReplace): the Backend API
// has no update route, so the endpoint is destroyed and recreated. The check
// asserts the id actually changed, proving a replacement rather than an
// (impossible) in-place update.
func TestAccWebhookEndpoint_replace(t *testing.T) {
	url1 := fmt.Sprintf("https://hooks.%s.example.com/atlas", acctest.RandString(8))
	url2 := fmt.Sprintf("https://hooks.%s.example.com/atlas", acctest.RandString(8))
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookEndpointConfig(url1, `["user.created"]`),
				Check:  testAccCaptureAttr("atlas_webhook_endpoint.test", "id", &firstID),
			},
			{
				Config: testAccWebhookEndpointConfig(url2, `["user.created"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "url", url2),
					testAccCheckAttrChanged("atlas_webhook_endpoint.test", "id", &firstID),
				),
			},
		},
	})
}

// TestAccWebhookEndpoint_disappears deletes the endpoint out of band and expects
// the next plan to be non-empty (the provider re-plans its recreation).
func TestAccWebhookEndpoint_disappears(t *testing.T) {
	url := fmt.Sprintf("https://hooks.%s.example.com/atlas", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookEndpointConfig(url, `["user.created"]`),
				Check: deleteOutOfBand("atlas_webhook_endpoint.test", func(ctx context.Context, c *client.Client, id string) error {
					return c.DeleteWebhookEndpoint(ctx, id)
				}),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccWebhookEndpointConfig(url, events string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_webhook_endpoint" "test" {
  url            = %[1]q
  enabled_events = %[2]s
}
`, url, events)
}
