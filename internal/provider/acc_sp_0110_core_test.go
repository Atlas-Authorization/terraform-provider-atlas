package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// Live acceptance tests (TF_ACC=1 against a sandbox instance) for the 0.11.0
// resources and data sources. Each test creates, reads back, checks that the
// follow-up plan is empty (the framework does this after every apply step),
// updates in place where the resource supports it, imports where supported and
// destroys. Targets are harmless: example.com URLs and the sandbox's own ids.
//
// Write-only (_wo) attributes need Terraform 1.11+.

// ---- atlas_secret_key ------------------------------------------------------

func TestAccSecretKey_lifecycle(t *testing.T) {
	cfg := func(name, extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_secret_key" "test" {
  name         = %q
  scopes       = ["users:read"]
  external_ref = "tfacc-secret-key"
  %s
}
data "atlas_secret_key" "test" { id = atlas_secret_key.test.id }
`, name, extra)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("tfacc key", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlas_secret_key.test", "id"),
					resource.TestCheckResourceAttrSet("atlas_secret_key.test", "secret"),
					resource.TestCheckResourceAttrSet("atlas_secret_key.test", "prefix"),
					resource.TestCheckResourceAttrSet("atlas_secret_key.test", "expires_at"),
					resource.TestCheckResourceAttr("atlas_secret_key.test", "status", "active"),
					resource.TestCheckResourceAttr("atlas_secret_key.test", "environment", "test"),
					resource.TestCheckResourceAttrPair("data.atlas_secret_key.test", "prefix", "atlas_secret_key.test", "prefix"),
					resource.TestCheckResourceAttrPair("data.atlas_secret_key.test", "name", "atlas_secret_key.test", "name"),
					resource.TestCheckNoResourceAttr("data.atlas_secret_key.test", "secret"),
				),
			},
			{
				Config: cfg("tfacc key renamed", `
  allowed_ips = ["203.0.113.0/24"]
  metadata    = { team = "acc" }
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_secret_key.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_secret_key.test", "name", "tfacc key renamed"),
					resource.TestCheckResourceAttr("atlas_secret_key.test", "metadata.team", "acc"),
					resource.TestCheckResourceAttr("atlas_secret_key.test", "allowed_ips.#", "1"),
					resource.TestCheckResourceAttrSet("atlas_secret_key.test", "secret"),
				),
			},
			{
				ResourceName:            "atlas_secret_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

// ---- atlas_partner_link / atlas_outbound_partner_link ----------------------

func TestAccPartnerLink_lifecycle(t *testing.T) {
	cfg := func(extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_partner_link" "test" {
  partner = "tfacc-partner"
  name    = "TF acceptance partner"
  scopes  = ["users:read"]
  %s
}
data "atlas_partner_link" "test" { id = atlas_partner_link.test.id }
`, extra)
	}
	var secret1 string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlas_partner_link.test", "client_id"),
					resource.TestCheckResourceAttrSet("atlas_partner_link.test", "client_secret"),
					resource.TestCheckResourceAttrSet("atlas_partner_link.test", "token_endpoint"),
					resource.TestCheckResourceAttr("atlas_partner_link.test", "status", "active"),
					resource.TestCheckResourceAttrPair("data.atlas_partner_link.test", "client_id", "atlas_partner_link.test", "client_id"),
					resource.TestCheckNoResourceAttr("data.atlas_partner_link.test", "client_secret"),
					testAccCaptureAttr("atlas_partner_link.test", "client_secret", &secret1),
				),
			},
			{
				// mTLS lock updates in place and keeps the secret.
				Config: cfg(`cert_fingerprints = ["Zm9vYmFyZm9vYmFyZm9vYmFyZm9vYmFyZm9vYmFyMTI"]`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_partner_link.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_partner_link.test", "cert_fingerprints.#", "1"),
				),
			},
			{
				// Rotation lands a new secret in place.
				Config: cfg(`
  cert_fingerprints     = ["Zm9vYmFyZm9vYmFyZm9vYmFyZm9vYmFyZm9vYmFyMTI"]
  rotate_secret_trigger = "r1"
`),
				Check: testAccCheckAttrChanged("atlas_partner_link.test", "client_secret", &secret1),
			},
			{
				ResourceName:            "atlas_partner_link.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret", "rotate_secret_trigger"},
			},
		},
	})
}

func TestAccOutboundPartnerLink_lifecycle(t *testing.T) {
	cfg := func(scope string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_outbound_partner_link" "test" {
  partner       = "tfaccout"
  name          = "TF acceptance outbound"
  client_id     = "tfacc-client"
  client_secret = "tfacc-not-a-real-secret-0000"
  scopes        = [%q]
  base_url      = "https://example.com"
}
data "atlas_outbound_partner_link" "test" { partner = atlas_outbound_partner_link.test.partner }
`, scope)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("dsar:write"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_outbound_partner_link.test", "base_url", "https://example.com"),
					resource.TestCheckResourceAttrSet("atlas_outbound_partner_link.test", "token_url"),
					resource.TestCheckResourceAttr("atlas_outbound_partner_link.test", "auth_method", "client_credentials"),
					resource.TestCheckResourceAttr("atlas_outbound_partner_link.test", "status", "active"),
					resource.TestCheckResourceAttr("data.atlas_outbound_partner_link.test", "client_id", "tfacc-client"),
					resource.TestCheckNoResourceAttr("data.atlas_outbound_partner_link.test", "client_secret"),
				),
			},
			{
				Config: cfg("dsar:read"),
				Check:  resource.TestCheckResourceAttr("atlas_outbound_partner_link.test", "scopes.#", "1"),
			},
			{
				ResourceName:            "atlas_outbound_partner_link.test",
				ImportState:             true,
				ImportStateId:           "tfaccout",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret"},
			},
		},
	})
}

// ---- atlas_whoami ----------------------------------------------------------

func TestAccWhoamiDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "atlas_whoami" "me" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlas_whoami.me", "id"),
					resource.TestCheckResourceAttr("data.atlas_whoami.me", "key_type", "secret"),
					resource.TestCheckResourceAttr("data.atlas_whoami.me", "environment", "test"),
					resource.TestCheckResourceAttrSet("data.atlas_whoami.me", "instance_id"),
					resource.TestCheckResourceAttrSet("data.atlas_whoami.me", "account_id"),
					resource.TestCheckResourceAttrSet("data.atlas_whoami.me", "effective_scopes.#"),
				),
			},
		},
	})
}
