package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccPolicyOrgID resolves the existing organization id whose policy the test
// manages. It has no config equivalent — the org must already exist in the
// target instance — so it is supplied via ATLAS_TEST_ORG_ID and the test skips
// when absent.
func testAccPolicyOrgID(t *testing.T) string {
	t.Helper()
	id := os.Getenv("ATLAS_TEST_ORG_ID")
	if id == "" {
		t.Skip("ATLAS_TEST_ORG_ID must be set (an existing organization id) to run the organization policy acceptance test")
	}
	return id
}

// TestAccOrganizationPolicy_basic sets an org policy, checks it, then imports it
// by organization id.
func TestAccOrganizationPolicy_basic(t *testing.T) {
	orgID := testAccPolicyOrgID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationPolicyConfig(orgID, "true", `["password", "oauth"]`, 3600),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_organization_policy.test", "organization_id", orgID),
					resource.TestCheckResourceAttr("atlas_organization_policy.test", "require_mfa", "true"),
					resource.TestCheckResourceAttr("atlas_organization_policy.test", "allowed_sign_in_methods.#", "2"),
					resource.TestCheckResourceAttr("atlas_organization_policy.test", "max_session_age_seconds", "3600"),
				),
			},
			{
				ResourceName:      "atlas_organization_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccOrganizationPolicy_update tightens then relaxes a field in place.
func TestAccOrganizationPolicy_update(t *testing.T) {
	orgID := testAccPolicyOrgID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationPolicyConfig(orgID, "true", `["password"]`, 3600),
				Check:  resource.TestCheckResourceAttr("atlas_organization_policy.test", "require_mfa", "true"),
			},
			{
				Config: testAccOrganizationPolicyConfig(orgID, "false", `["password", "oauth"]`, 7200),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_organization_policy.test", "require_mfa", "false"),
					resource.TestCheckResourceAttr("atlas_organization_policy.test", "allowed_sign_in_methods.#", "2"),
					resource.TestCheckResourceAttr("atlas_organization_policy.test", "max_session_age_seconds", "7200"),
				),
			},
		},
	})
}

func testAccOrganizationPolicyConfig(orgID, requireMfa, methods string, maxAge int) string {
	return providerConfig + fmt.Sprintf(`
resource "atlas_organization_policy" "test" {
  organization_id         = %[1]q
  require_mfa             = %[2]s
  allowed_sign_in_methods = %[3]s
  max_session_age_seconds = %[4]d
}
`, orgID, requireMfa, methods, maxAge)
}
