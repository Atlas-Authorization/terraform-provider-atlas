package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccOAuthClient_basic creates a confidential client, asserts the key
// fields, then imports it. client_secret is only revealed on create and is
// never returned by a read, so it is excluded from the import round-trip check.
func TestAccOAuthClient_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-oauth")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOAuthClientConfig(name, "https://app.example.com/callback"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_oauth_client.test", "name", name),
					resource.TestCheckResourceAttr("atlas_oauth_client.test", "redirect_uris.#", "1"),
					resource.TestCheckResourceAttr("atlas_oauth_client.test", "redirect_uris.0", "https://app.example.com/callback"),
					resource.TestCheckResourceAttr("atlas_oauth_client.test", "token_endpoint_auth_method", "client_secret_basic"),
					// Server-assigned attributes are populated after apply.
					resource.TestCheckResourceAttrSet("atlas_oauth_client.test", "id"),
					resource.TestCheckResourceAttrSet("atlas_oauth_client.test", "client_id"),
					resource.TestCheckResourceAttrSet("atlas_oauth_client.test", "created_at"),
				),
			},
			{
				ResourceName:      "atlas_oauth_client.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The confidential secret is revealed only on create; a read (and
				// therefore an import) never returns it, so state cannot match.
				ImportStateVerifyIgnore: []string{"client_secret"},
			},
		},
	})
}

// TestAccOAuthClient_update mutates the mutable fields (name, redirect_uris) in
// place — the OAuth client has a real PATCH route, so this is an in-place update
// rather than a replacement.
func TestAccOAuthClient_update(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-oauth")
	updated := name + "-renamed"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOAuthClientConfig(name, "https://app.example.com/callback"),
				Check: resource.TestCheckResourceAttr(
					"atlas_oauth_client.test", "name", name),
			},
			{
				Config: testAccOAuthClientConfig(updated, "https://app.example.com/callback2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_oauth_client.test", "name", updated),
					resource.TestCheckResourceAttr("atlas_oauth_client.test", "redirect_uris.0", "https://app.example.com/callback2"),
				),
			},
		},
	})
}

func testAccOAuthClientConfig(name, redirectURI string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_oauth_client" "test" {
  name                       = %[1]q
  redirect_uris              = [%[2]q]
  allowed_scopes             = ["openid", "profile", "email"]
  grant_types                = ["authorization_code", "refresh_token"]
  token_endpoint_auth_method = "client_secret_basic"
}
`, name, redirectURI)
}
