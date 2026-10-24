package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// Live acceptance tests (TF_ACC=1, sandbox instance) for atlas_audit_export,
// atlas_status_subscription, atlas_status_dependency, the 0.11.0 fields of
// atlas_webhook_endpoint and their data sources. All receivers are
// https://example.com/... and alert contacts are example.com addresses.

func TestAccAuditExport_webhookLifecycle(t *testing.T) {
	cfg := func(name, extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_audit_export" "test" {
  name        = %q
  type        = "webhook"
  destination = { url = "https://example.com/tfacc-audit" }
  %s
}
data "atlas_audit_export" "test" { id = atlas_audit_export.test.id }
`, name, extra)
	}
	var secret1, secret2 string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("tfacc export", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlas_audit_export.test", "id"),
					resource.TestCheckResourceAttrSet("atlas_audit_export.test", "secret"),
					resource.TestCheckResourceAttr("atlas_audit_export.test", "enabled", "true"),
					resource.TestCheckResourceAttr("atlas_audit_export.test", "destination.url", "https://example.com/tfacc-audit"),
					resource.TestCheckResourceAttrSet("atlas_audit_export.test", "status"),
					resource.TestCheckResourceAttrPair("data.atlas_audit_export.test", "name", "atlas_audit_export.test", "name"),
					resource.TestCheckNoResourceAttr("data.atlas_audit_export.test", "secret"),
					testAccCaptureAttr("atlas_audit_export.test", "secret", &secret1),
				),
			},
			{
				Config: cfg("tfacc export renamed", `
  enabled       = false
  action_filter = ["user.", "organization.*"]
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_audit_export.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_audit_export.test", "name", "tfacc export renamed"),
					resource.TestCheckResourceAttr("atlas_audit_export.test", "enabled", "false"),
					resource.TestCheckResourceAttr("atlas_audit_export.test", "action_filter.#", "2"),
					testAccCheckAttrUnchanged("atlas_audit_export.test", "secret", &secret1),
				),
			},
			{
				Config: cfg("tfacc export renamed", `
  enabled               = false
  action_filter         = ["user.", "organization.*"]
  rotate_secret_trigger = "r1"
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureAttr("atlas_audit_export.test", "secret", &secret2),
					testAccCheckAttrChanged("atlas_audit_export.test", "secret", &secret1),
				),
			},
			{
				ResourceName:            "atlas_audit_export.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "rotate_secret_trigger", "backfill_hours", "last_delivered_at", "exported_count", "status", "last_error"},
			},
		},
	})
}

func TestAccStatusSubscription_lifecycle(t *testing.T) {
	cfg := func(extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_status_subscription" "test" {
  %s
}
data "atlas_status_subscription" "test" { depends_on = [atlas_status_subscription.test] }
`, extra)
	}
	var rss string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_status_subscription.test", "id", "status_subscription"),
					resource.TestCheckResourceAttr("atlas_status_subscription.test", "min_impact", "minor"),
					resource.TestCheckResourceAttrSet("atlas_status_subscription.test", "rss_url"),
					resource.TestCheckResourceAttrSet("atlas_status_subscription.test", "atom_url"),
					resource.TestCheckResourceAttrPair("data.atlas_status_subscription.test", "min_impact", "atlas_status_subscription.test", "min_impact"),
					testAccCaptureAttr("atlas_status_subscription.test", "rss_url", &rss),
				),
			},
			{
				Config: cfg(`
  min_impact          = "major"
  include_maintenance = false
  email_owners        = false
  webhook             = false
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_status_subscription.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_status_subscription.test", "min_impact", "major"),
					resource.TestCheckResourceAttr("atlas_status_subscription.test", "include_maintenance", "false"),
					resource.TestCheckResourceAttr("atlas_status_subscription.test", "email_owners", "false"),
					testAccCheckAttrUnchanged("atlas_status_subscription.test", "rss_url", &rss),
				),
			},
			{
				Config: cfg(`
  min_impact          = "major"
  include_maintenance = false
  email_owners        = false
  webhook             = false
  rotate_feed_trigger = "r1"
`),
				Check: testAccCheckAttrChanged("atlas_status_subscription.test", "rss_url", &rss),
			},
			{
				ResourceName:            "atlas_status_subscription.test",
				ImportState:             true,
				ImportStateId:           "status_subscription",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rotate_feed_trigger"},
			},
		},
	})
}

func TestAccStatusDependency_lifecycle(t *testing.T) {
	cfg := func(name, extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_status_dependency" "test" {
  name = %q
  url  = "https://www.githubstatus.com"
  %s
}
data "atlas_status_dependency" "test" { id = atlas_status_dependency.test.id }
`, name, extra)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("tfacc dep", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlas_status_dependency.test", "id"),
					resource.TestCheckResourceAttr("atlas_status_dependency.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("atlas_status_dependency.test", "state"),
					resource.TestCheckResourceAttrPair("data.atlas_status_dependency.test", "url", "atlas_status_dependency.test", "url"),
				),
			},
			{
				Config: cfg("tfacc dep renamed", `
  enabled          = false
  component_filter = ["Git Operations"]
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_status_dependency.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_status_dependency.test", "name", "tfacc dep renamed"),
					resource.TestCheckResourceAttr("atlas_status_dependency.test", "enabled", "false"),
					resource.TestCheckResourceAttr("atlas_status_dependency.test", "component_filter.#", "1"),
				),
			},
			{
				ResourceName:      "atlas_status_dependency.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Poll-derived fields move on their own.
				ImportStateVerifyIgnore: []string{"state", "indicator", "description", "last_checked_at", "last_error", "stale"},
			},
		},
	})
}

func TestAccWebhookEndpoint_newFields(t *testing.T) {
	cfg := func(extra string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_webhook_endpoint" "test" {
  url            = "https://example.com/tfacc-hook"
  enabled_events = ["user.created"]
  external_ref   = "tfacc-wh-ref"
  %s
}
data "atlas_webhook_endpoint" "test" { id = atlas_webhook_endpoint.test.id }
`, extra)
	}
	var id string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(`alert_contacts = ["tfacc@example.com"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlas_webhook_endpoint.test", "secret"),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "alert_contacts.#", "1"),
					resource.TestCheckResourceAttrSet("atlas_webhook_endpoint.test", "event_naming"),
					resource.TestCheckResourceAttrSet("atlas_webhook_endpoint.test", "signature_versions.#"),
					resource.TestCheckResourceAttrSet("atlas_webhook_endpoint.test", "health"),
					resource.TestCheckResourceAttrSet("atlas_webhook_endpoint.test", "consecutive_failures"),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "external_ref", "tfacc-wh-ref"),
					resource.TestCheckResourceAttrPair("data.atlas_webhook_endpoint.test", "event_naming", "atlas_webhook_endpoint.test", "event_naming"),
					resource.TestCheckNoResourceAttr("data.atlas_webhook_endpoint.test", "secret"),
					testAccCaptureAttr("atlas_webhook_endpoint.test", "id", &id),
				),
			},
			{
				Config: cfg(`
  alert_contacts     = ["tfacc@example.com", "tfacc2@example.com"]
  description        = "tf acceptance"
  event_naming       = "legacy"
  external           = true
  signature_versions = ["v1", "v2"]
  metadata           = { team = "acc" }
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("atlas_webhook_endpoint.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "alert_contacts.#", "2"),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "description", "tf acceptance"),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "event_naming", "legacy"),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "external", "true"),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "signature_versions.#", "2"),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.test", "metadata.team", "acc"),
					testAccCheckAttrUnchanged("atlas_webhook_endpoint.test", "id", &id),
				),
			},
			{
				ResourceName:            "atlas_webhook_endpoint.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "external_ref", "health", "consecutive_failures", "failing_since", "last_failing_since", "last_success_at", "disabled_at", "active"},
			},
		},
	})
}

func TestAccWebhookEndpoint_destination(t *testing.T) {
	cfg := providerConfig + `
resource "atlas_webhook_endpoint" "dest" {
  enabled_events = ["user.created"]
  alert_contacts = ["tfacc@example.com"]
  destination = {
    type       = "cloudflare_queue"
    account_id = "0123456789abcdef0123456789abcdef"
    queue_id   = "0123456789abcdef0123456789abcdef"
  }
  destination_credentials = { api_token = "tfacc-not-a-real-token" }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.dest", "destination.type", "cloudflare_queue"),
					resource.TestCheckResourceAttr("atlas_webhook_endpoint.dest", "destination_credentials_set", "true"),
				),
			},
		},
	})
}

func TestAccAuditExport_s3Lifecycle(t *testing.T) {
	cfg := func(prefix string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_audit_export" "s3" {
  name = "tfacc s3"
  type = "s3"
  destination = {
    endpoint      = "https://example.com"
    bucket        = "tfacc-bucket"
    region        = "us-east-1"
    prefix        = %q
    access_key_id = "AKIATFACCNOTREAL"
  }
  secret_access_key = "tfacc-not-a-real-secret"
}
`, prefix)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("audit/"),
				Check:  resource.TestCheckResourceAttr("atlas_audit_export.s3", "destination.bucket", "tfacc-bucket"),
			},
			{
				Config: cfg("audit2/"),
				Check:  resource.TestCheckResourceAttr("atlas_audit_export.s3", "destination.prefix", "audit2/"),
			},
		},
	})
}
