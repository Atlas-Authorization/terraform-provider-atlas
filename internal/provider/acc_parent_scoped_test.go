package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Live acceptance tests (TF_ACC=1 + a disposable instance) for the generated
// resources that live under a parent, which TestAccGeneric_examples skips. Each
// creates the parent in the same configuration where the provider can, then the
// resource, then imports it. Where the parent is not Terraform-managed (a
// connection, a published policy version, a verified DNS record, a user) the
// test reads its id from the environment and skips when it is unset.

// accEnv returns the environment value or skips the test.
func accEnv(t *testing.T, name, why string) string {
	t.Helper()
	v := os.Getenv(name)
	if v == "" {
		t.Skipf("set %s (%s)", name, why)
	}
	return v
}

// accImportID joins the named attributes of a resource with "/" for import.
func accImportID(resName string, attrs ...string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resName]
		if !ok {
			return "", fmt.Errorf("%s not in state", resName)
		}
		id := ""
		for i, a := range attrs {
			if i > 0 {
				id += "/"
			}
			id += rs.Primary.Attributes[a]
		}
		return id, nil
	}
}

func TestAccPolicyAudience_parentScoped(t *testing.T) {
	cfg := func(audience string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_policy" "tos" {
  key        = "tf-acc-audience"
  title      = "Terms of service"
  display    = "scroll"
  applies_to = "sign_in_and_sign_up"
}

resource "atlas_policy_audience" "tos" {
  policy_id = atlas_policy.tos.id
  audience  = %q
}
`, audience)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: cfg("everyone"), Check: resource.TestCheckResourceAttr("atlas_policy_audience.tos", "audience", "everyone")},
			{
				ResourceName:      "atlas_policy_audience.tos",
				ImportState:       true,
				ImportStateIdFunc: accImportID("atlas_policy_audience.tos", "policy_id"),
				ImportStateVerify: true,
			},
		},
	})
}

// Policy versions are content, not Terraform resources, so the policy and one
// of its published versions come from the environment.
func TestAccPolicyTranslation_parentScoped(t *testing.T) {
	policyID := accEnv(t, "ATLAS_ACC_POLICY_ID", "a policy with a published version")
	versionID := accEnv(t, "ATLAS_ACC_POLICY_VERSION_ID", "a version of that policy")
	cfg := func(title string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_policy_translation" "fr" {
  policy_id        = %q
  version_id       = %q
  locale           = "fr"
  title            = %q
  content_markdown = "Conditions d'utilisation."
}
`, policyID, versionID, title)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: cfg("Conditions"), Check: resource.TestCheckResourceAttr("atlas_policy_translation.fr", "title", "Conditions")},
			{Config: cfg("Conditions generales"), Check: resource.TestCheckResourceAttr("atlas_policy_translation.fr", "title", "Conditions generales")},
			{
				ResourceName:            "atlas_policy_translation.fr",
				ImportState:             true,
				ImportStateIdFunc:       accImportID("atlas_policy_translation.fr", "policy_id", "version_id", "locale"),
				ImportStateVerifyIgnore: []string{"content_markdown", "content_html"},
			},
		},
	})
}

// A connection is created by the owner's API call, not by Terraform.
func TestAccConnectionAccessRules_parentScoped(t *testing.T) {
	connID := accEnv(t, "ATLAS_ACC_CONNECTION_ID", "an existing connection")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "atlas_connection_access_rules" "this" {
  connection_id = %q
  rules         = jsonencode([])
}
`, connID),
				Check: resource.TestCheckResourceAttr("atlas_connection_access_rules.this", "connection_id", connID),
			},
			{
				ResourceName:      "atlas_connection_access_rules.this",
				ImportState:       true,
				ImportStateIdFunc: accImportID("atlas_connection_access_rules.this", "connection_id"),
				ImportStateVerify: true,
			},
		},
	})
}

// Attaching an app needs a publisher whose DNS record is already published
// (verify is an action, not Terraform state) and an app whose homepage_url is on
// that domain.
func TestAccOAuthPublisherClient_parentScoped(t *testing.T) {
	domain := accEnv(t, "ATLAS_ACC_PUBLISHER_DOMAIN", "a domain whose Atlas TXT record is published")
	clientID := accEnv(t, "ATLAS_ACC_PUBLISHER_CLIENT_ID", "an app with an https homepage_url on that domain")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "atlas_oauth_publisher" "this" {
  name   = "Terraform acceptance"
  domain = %q
}

resource "atlas_oauth_publisher_client" "this" {
  publisher_id = atlas_oauth_publisher.this.id
  client_id    = %q
}
`, domain, clientID),
				Check: resource.TestCheckResourceAttrPair("atlas_oauth_publisher_client.this", "publisher_id", "atlas_oauth_publisher.this", "id"),
			},
			{
				ResourceName:            "atlas_oauth_publisher_client.this",
				ImportState:             true,
				ImportStateIdFunc:       accImportID("atlas_oauth_publisher_client.this", "publisher_id", "client_id"),
				ImportStateVerifyIgnore: []string{"allow_subject_change"},
			},
		},
	})
}

// The delegator is a user, which Terraform does not create, so its id is read
// from the environment; the agent (machine) is created in the same config.
func TestAccAgentDelegation_parentScoped(t *testing.T) {
	userID := accEnv(t, "ATLAS_ACC_USER_ID", "a user who can delegate to an agent")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "atlas_machine" "agent" {
  name          = "tf-acc-agent"
  owner_user_id = %q
}

resource "atlas_agent_delegation" "this" {
  child_machine_id  = atlas_machine.agent.id
  delegator_user_id = %q
  scopes            = ["read"]
}
`, userID, userID),
				Check: resource.TestCheckResourceAttrPair("atlas_agent_delegation.this", "child_machine_id", "atlas_machine.agent", "id"),
			},
			{
				ResourceName:            "atlas_agent_delegation.this",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"expires_in_seconds"},
			},
		},
	})
}
