package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Live acceptance tests for the resources the first acceptance suites did not
// cover: Sign in with Atlas apps (platform key), organization design,
// entitlement overrides, email/SMS providers, email and notification
// templates, the org settings schema and the oauth_provider data source.

// testAccPlatformPreCheck additionally requires a platform key (psk_...).
func testAccPlatformPreCheck(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	if os.Getenv("ATLAS_PLATFORM_KEY") == "" {
		t.Fatal("ATLAS_PLATFORM_KEY must be set for the atlas_id_app acceptance tests (a psk_ platform key)")
	}
}

func testAccPlatformClient() *client.Client {
	c := testAccClient()
	c.PlatformKey = os.Getenv("ATLAS_PLATFORM_KEY")
	return c
}

// testAccCheckIDAppDestroyed asserts every atlas_id_app in the final state is
// gone on the server after destroy.
func testAccCheckIDAppDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "atlas_id_app" {
			continue
		}
		_, err := testAccPlatformClient().GetAtlasIDApp(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("atlas_id_app %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("checking atlas_id_app %s: %v", rs.Primary.ID, err)
		}
	}
	return nil
}

const idAppStep1 = `
resource "atlas_id_app" "test" {
  name                      = %[1]q
  redirect_uris             = ["https://app.example.com/callback"]
  post_logout_redirect_uris = ["https://app.example.com/signed-out"]
  homepage_url              = "https://app.example.com"

  branding = jsonencode({
    theme        = "dark"
    accent_color = "#4F46E5"
    copy = {
      sign_in_title = "Welcome back"
    }
    design = {
      palette = { primary = "#4f46e5" }
    }
  })
}
`

// Step 2 changes the logout URIs, adds back/front-channel logout and rewrites
// branding (a key removed, keys added) in place.
const idAppStep2 = `
resource "atlas_id_app" "test" {
  name                                 = "%[1]s-v2"
  redirect_uris                        = ["https://app.example.com/callback", "http://localhost:3000/callback"]
  post_logout_redirect_uris            = ["https://app.example.com/bye", "http://localhost:3000/bye"]
  backchannel_logout_uri               = "https://app.example.com/oidc/backchannel"
  backchannel_logout_session_required  = true
  frontchannel_logout_uri              = "https://app.example.com/oidc/frontchannel"
  frontchannel_logout_session_required = true
  allowed_scopes                       = ["openid", "profile", "email", "offline_access"]
  homepage_url                         = "https://app.example.com"

  branding = jsonencode({
    accent_color       = "#0EA5E9"
    background_color   = "#0B1020"
    privacy_policy_url = "https://app.example.com/privacy"
    copy = {
      sign_in_title    = "Sign in to Example"
      sign_in_subtitle = "Use your Atlas account"
    }
    design = {
      palette = { primary = "#0ea5e9" }
      shape   = { radius = 12 }
    }
  })
}
`

// Step 3 is step 2 with the branding JSON written by hand: different key
// order and whitespace. It must plan clean (semantic comparison).
const idAppStep3 = `
resource "atlas_id_app" "test" {
  name                                 = "%[1]s-v2"
  redirect_uris                        = ["https://app.example.com/callback", "http://localhost:3000/callback"]
  post_logout_redirect_uris            = ["https://app.example.com/bye", "http://localhost:3000/bye"]
  backchannel_logout_uri               = "https://app.example.com/oidc/backchannel"
  backchannel_logout_session_required  = true
  frontchannel_logout_uri              = "https://app.example.com/oidc/frontchannel"
  frontchannel_logout_session_required = true
  allowed_scopes                       = ["openid", "profile", "email", "offline_access"]
  homepage_url                         = "https://app.example.com"

  branding = <<-JSON
    {
      "design": { "shape": { "radius": 12 }, "palette": { "primary": "#0ea5e9" } },
      "copy":   { "sign_in_subtitle": "Use your Atlas account", "sign_in_title": "Sign in to Example" },
      "privacy_policy_url": "https://app.example.com/privacy",
      "background_color":   "#0B1020",
      "accent_color":       "#0EA5E9"
    }
  JSON
}
`

func TestAccIDApp_full(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-idapp")
	var secret1, id1 string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPlatformPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIDAppDestroyed,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(idAppStep1, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_id_app.test", "name", name),
					resource.TestCheckResourceAttr("atlas_id_app.test", "type", "confidential"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "post_logout_redirect_uris.#", "1"),
					resource.TestCheckResourceAttrSet("atlas_id_app.test", "client_id"),
					resource.TestCheckResourceAttrSet("atlas_id_app.test", "client_secret"),
					resource.TestCheckResourceAttrSet("atlas_id_app.test", "secret_prefix"),
					testAccCaptureAttr("atlas_id_app.test", "client_secret", &secret1),
					testAccCaptureAttr("atlas_id_app.test", "id", &id1),
				),
			},
			{
				Config: providerConfig + fmt.Sprintf(idAppStep2, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttrUnchanged("atlas_id_app.test", "id", &id1),
					testAccCheckAttrUnchanged("atlas_id_app.test", "client_secret", &secret1),
					resource.TestCheckResourceAttr("atlas_id_app.test", "name", name+"-v2"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "redirect_uris.#", "2"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "post_logout_redirect_uris.#", "2"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "post_logout_redirect_uris.0", "https://app.example.com/bye"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "backchannel_logout_uri", "https://app.example.com/oidc/backchannel"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "backchannel_logout_session_required", "true"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "frontchannel_logout_uri", "https://app.example.com/oidc/frontchannel"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "frontchannel_logout_session_required", "true"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "allowed_scopes.#", "4"),
				),
			},
			{
				// Same document written by hand (other key order and whitespace).
				// Terraform compares configuration text, so this applies once as
				// a no-op update; the harness then asserts the refresh plan is
				// empty (the server's canonical JSON is semantically equal).
				Config: providerConfig + fmt.Sprintf(idAppStep3, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttrUnchanged("atlas_id_app.test", "id", &id1),
					resource.TestCheckResourceAttr("atlas_id_app.test", "post_logout_redirect_uris.#", "2"),
				),
			},
			{
				// Re-planning the unchanged configuration shows no diff.
				Config:   providerConfig + fmt.Sprintf(idAppStep3, name),
				PlanOnly: true,
			},
			{
				// Rotating the secret keeps the app and replaces client_secret.
				Config: providerConfig + fmt.Sprintf(idAppStep2[:len(idAppStep2)-3]+"\n  rotate_secret_trigger = \"1\"\n}\n", name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttrUnchanged("atlas_id_app.test", "id", &id1),
					testAccCheckAttrChanged("atlas_id_app.test", "client_secret", &secret1),
				),
			},
			{
				ResourceName:      "atlas_id_app.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The secret is revealed on create/rotation only; the trigger is
				// configuration-only.
				ImportStateVerifyIgnore: []string{"client_secret", "rotate_secret_trigger"},
			},
		},
	})
}

func TestAccIDApp_public(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-idapp-pub")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPlatformPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIDAppDestroyed,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "atlas_id_app" "test" {
  name          = %q
  type          = "public"
  redirect_uris = ["http://localhost:5173/callback"]
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_id_app.test", "type", "public"),
					resource.TestCheckNoResourceAttr("atlas_id_app.test", "client_secret"),
					resource.TestCheckResourceAttr("atlas_id_app.test", "post_logout_redirect_uris.#", "0"),
				),
			},
			{ResourceName: "atlas_id_app.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccOrganizationDesign_basic(t *testing.T) {
	orgID := os.Getenv("ATLAS_TEST_ORG_ID")
	if orgID == "" {
		t.Skip("ATLAS_TEST_ORG_ID must be set (an existing organization id) to run the organization design acceptance test")
	}
	cfg := func(primary string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_organization_design" "test" {
  organization_id = %q
  design          = jsonencode({ palette = { primary = %q } })
}
`, orgID, primary)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("#4f46e5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_organization_design.test", "organization_id", orgID),
					resource.TestCheckResourceAttrSet("atlas_organization_design.test", "resolved"),
				),
			},
			{Config: cfg("#0ea5e9")},
			{
				ResourceName:                         "atlas_organization_design.test",
				ImportState:                          true,
				ImportStateId:                        orgID,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "organization_id",
			},
		},
	})
}

func TestAccEntitlementOverride_basic(t *testing.T) {
	userID := os.Getenv("ATLAS_TEST_USER_ID")
	if userID == "" {
		t.Skip("ATLAS_TEST_USER_ID must be set (an existing user id) to run the entitlement override acceptance test")
	}
	cfg := func(limit int) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_entitlement_override" "test" {
  subject_type = "user"
  subject_id   = %q
  features     = ["sso", "audit_log"]
  limits       = { seats = %d, projects = -1 }
}
`, userID, limit)
	}
	var id string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(25),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_entitlement_override.test", "features.#", "2"),
					resource.TestCheckResourceAttr("atlas_entitlement_override.test", "limits.seats", "25"),
					resource.TestCheckResourceAttr("atlas_entitlement_override.test", "limits.projects", "-1"),
					testAccCaptureAttr("atlas_entitlement_override.test", "id", &id),
				),
			},
			{
				// Overrides are immutable: a changed limit replaces the grant.
				Config: cfg(50),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_entitlement_override.test", "limits.seats", "50"),
					testAccCheckAttrChanged("atlas_entitlement_override.test", "id", &id),
				),
			},
			{
				ResourceName: "atlas_entitlement_override.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["atlas_entitlement_override.test"]
					return "user/" + userID + "/" + rs.Primary.ID, nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEmailProvider_smtp(t *testing.T) {
	cfg := func(port int) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_email_provider" "test" {
  from_address  = "Acme <no-reply@example.com>"
  force_destroy = true
  smtp = {
    host     = "smtp.example.com"
    port     = %d
    username = "mailer"
    password = "not-a-real-password"
  }
}
`, port)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(587),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_email_provider.test", "id", "email"),
					resource.TestCheckResourceAttr("atlas_email_provider.test", "transport", "smtp"),
					resource.TestCheckResourceAttr("atlas_email_provider.test", "smtp.port", "587"),
					resource.TestCheckTypeSetElemAttr("atlas_email_provider.test", "secrets_set.*", "password"),
				),
			},
			{
				Config: cfg(2525),
				Check:  resource.TestCheckResourceAttr("atlas_email_provider.test", "smtp.port", "2525"),
			},
			{
				ResourceName:            "atlas_email_provider.test",
				ImportState:             true,
				ImportStateId:           "email",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"smtp.password", "force_destroy"},
			},
		},
	})
}

func TestAccSMSProvider_twilio(t *testing.T) {
	cfg := func(from string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_sms_provider" "test" {
  from_address = %q
  twilio = {
    account_sid = "AC00000000000000000000000000000000"
    auth_token  = "not-a-real-token"
  }
}
`, from)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("+15005550006"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_sms_provider.test", "id", "sms"),
					resource.TestCheckResourceAttr("atlas_sms_provider.test", "transport", "twilio"),
					resource.TestCheckTypeSetElemAttr("atlas_sms_provider.test", "secrets_set.*", "auth_token"),
				),
			},
			{
				Config: cfg("+15005550007"),
				Check:  resource.TestCheckResourceAttr("atlas_sms_provider.test", "from_address", "+15005550007"),
			},
			{
				ResourceName:            "atlas_sms_provider.test",
				ImportState:             true,
				ImportStateId:           "sms",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"twilio.auth_token"},
			},
		},
	})
}

func TestAccEmailTemplate_basic(t *testing.T) {
	cfg := func(subject string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_email_template" "test" {
  name    = "verification_code"
  subject = %q
  text    = "Your code is {{code}}."
}
`, subject)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("Your verification code"),
				Check:  resource.TestCheckResourceAttr("atlas_email_template.test", "id", "verification_code"),
			},
			{
				Config: cfg("Your sign-in code is {{code}}"),
				Check:  resource.TestCheckResourceAttr("atlas_email_template.test", "subject", "Your sign-in code is {{code}}"),
			},
			{ResourceName: "atlas_email_template.test", ImportState: true, ImportStateId: "verification_code", ImportStateVerify: true},
		},
	})
}

func TestAccNotificationTemplate_basic(t *testing.T) {
	name := "tf_acc_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	cfg := func(body string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_notification_template" "test" {
  name     = %q
  subject  = "New sign-in on {{device}}"
  body     = %q
  category = "new_device"
}
`, name, body)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("We noticed a sign-in from {{device}}."),
				Check:  resource.TestCheckResourceAttr("atlas_notification_template.test", "category", "new_device"),
			},
			{
				Config: cfg("A new device ({{device}}) signed in."),
				Check:  resource.TestCheckResourceAttr("atlas_notification_template.test", "body", "A new device ({{device}}) signed in."),
			},
			{
				ResourceName:                         "atlas_notification_template.test",
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func TestAccOrgSettingsSchema_basic(t *testing.T) {
	cfg := func(required string) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_org_settings_schema" "test" {
  schema = jsonencode({
    type = "object"
    properties = {
      brand_color   = { type = "string" }
      support_email = { type = "string" }
    }
    required = [%q]
  })
}
`, required)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("support_email"),
				Check:  resource.TestCheckResourceAttrSet("atlas_org_settings_schema.test", "version"),
			},
			{Config: cfg("brand_color")},
			{ResourceName: "atlas_org_settings_schema.test", ImportState: true, ImportStateId: "organization_settings_schema", ImportStateVerify: true},
		},
	})
}

func TestAccOAuthProviderDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "atlas_oauth_provider" "google" { provider_key = "google" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlas_oauth_provider.google", "display_name"),
					resource.TestCheckResourceAttrSet("data.atlas_oauth_provider.google", "redirect_uri"),
				),
			},
		},
	})
}
