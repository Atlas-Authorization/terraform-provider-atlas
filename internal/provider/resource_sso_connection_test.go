package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSsoConnection_basic creates an OIDC connection, checks its fields, then
// imports it. The write-only secret (oidc_client_secret) is never returned by
// the API, so it is excluded from the import round-trip.
func TestAccSsoConnection_basic(t *testing.T) {
	issuer := fmt.Sprintf("https://%s.okta.example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSsoConnectionConfig(issuer, "draft"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_sso_connection.test", "type", "oidc"),
					resource.TestCheckResourceAttr("atlas_sso_connection.test", "status", "draft"),
					resource.TestCheckResourceAttr("atlas_sso_connection.test", "oidc_issuer", issuer),
					resource.TestCheckResourceAttr("atlas_sso_connection.test", "oidc_client_id", "client-abc"),
					// The API confirms a secret is stored via the has_* flag only.
					resource.TestCheckResourceAttr("atlas_sso_connection.test", "has_secret", "true"),
					resource.TestCheckResourceAttrSet("atlas_sso_connection.test", "id"),
				),
			},
			{
				ResourceName:      "atlas_sso_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Every write-only secret is absent from a read/import: the API
				// exposes only has_secret / has_saml_certificate / has_discourse_secret.
				ImportStateVerifyIgnore: []string{
					"oidc_client_secret",
					"saml_idp_certificate",
					"discourse_secret",
				},
			},
		},
	})
}

// TestAccSsoConnection_update flips the connection from draft to active in place
// (type is RequiresReplace; the other fields PATCH).
func TestAccSsoConnection_update(t *testing.T) {
	issuer := fmt.Sprintf("https://%s.okta.example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSsoConnectionConfig(issuer, "draft"),
				Check:  resource.TestCheckResourceAttr("atlas_sso_connection.test", "status", "draft"),
			},
			{
				Config: testAccSsoConnectionConfig(issuer, "active"),
				Check:  resource.TestCheckResourceAttr("atlas_sso_connection.test", "status", "active"),
			},
		},
	})
}

func testAccSsoConnectionConfig(issuer, status string) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_sso_connection" "test" {
  type               = "oidc"
  status             = %[2]q
  oidc_issuer        = %[1]q
  oidc_client_id     = "client-abc"
  oidc_client_secret = "super-secret-value"
  allowed_domains    = ["example.com"]

  claim_role_mappings = [
    {
      claim    = "groups"
      value    = "Engineering"
      role_key = "org:member"
    },
  ]
}
`, issuer, status)
}
