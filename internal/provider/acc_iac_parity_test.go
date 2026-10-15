package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Live acceptance tests (TF_ACC=1 + a disposable instance) for the 0.8.0
// additions. Never point these at a production instance.

func TestAccOAuthClient_skipLogoutConfirmation(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-oauth-skip")
	cfg := func(skip bool) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_oauth_client" "test" {
  name                     = %[1]q
  redirect_uris            = ["https://app.example.com/callback"]
  first_party              = true
  skip_logout_confirmation = %[2]t
}

data "atlas_oauth_client" "test" {
  id         = atlas_oauth_client.test.id
  depends_on = [atlas_oauth_client.test]
}
`, name, skip)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_oauth_client.test", "skip_logout_confirmation", "true"),
					resource.TestCheckResourceAttr("data.atlas_oauth_client.test", "skip_logout_confirmation", "true"),
				),
			},
			{
				Config: cfg(false),
				Check:  resource.TestCheckResourceAttr("atlas_oauth_client.test", "skip_logout_confirmation", "false"),
			},
			{
				ResourceName:            "atlas_oauth_client.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret"},
			},
		},
	})
}

// A third-party client asking to skip the logout confirmation fails at plan
// time, before Atlas would answer 422.
func TestAccOAuthClient_skipLogoutConfirmationThirdParty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "atlas_oauth_client" "test" {
  name                     = "tf-acc-third-party"
  redirect_uris            = ["https://app.example.com/callback"]
  skip_logout_confirmation = true
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`needs a first-party client`),
			},
		},
	})
}

func TestAccCookieConsent_settingsLink(t *testing.T) {
	cfg := func(link string) string {
		return providerConfig + `
resource "atlas_cookie_consent" "test" {
  script_url = "https://consent.cookiebot.com/uc.js"
  script_id  = "Cookiebot"
` + link + `
}
`
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg(`
  settings_link = {
    label      = "Cookie settings"
    preset     = "cookiemunch"
    attributes = { "aria-label" = "Open cookie settings" }
  }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_cookie_consent.test", "settings_link.label", "Cookie settings"),
					resource.TestCheckResourceAttr("atlas_cookie_consent.test", "settings_link.rendered_attributes.data-cc", "show-settings"),
					resource.TestCheckResourceAttr("atlas_cookie_consent.test", "settings_link.rendered_attributes.aria-label", "Open cookie settings"),
				),
			},
			{
				Config: cfg(`
  settings_link = {
    label  = "Cookies"
    preset = "cookiebot"
  }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlas_cookie_consent.test", "settings_link.preset", "cookiebot"),
					resource.TestCheckResourceAttr("atlas_cookie_consent.test", "settings_link.rendered_attributes.data-atlas-cmp-open", "cookiebot"),
					resource.TestCheckNoResourceAttr("atlas_cookie_consent.test", "settings_link.attributes"),
				),
			},
			{ResourceName: "atlas_cookie_consent.test", ImportState: true, ImportStateId: "cookie_consent", ImportStateVerify: true},
			{
				// Removing the block removes the link.
				Config: cfg(""),
				Check:  resource.TestCheckNoResourceAttr("atlas_cookie_consent.test", "settings_link"),
			},
		},
	})
}

// force_destroy = true removes the email provider with ?force=true. (A
// development instance keeps the shared sender, so it is never refused; the
// refusal path is covered by the offline harness.)
func TestAccEmailProvider_forceDestroy(t *testing.T) {
	cfg := func(force bool) string {
		return providerConfig + fmt.Sprintf(`
resource "atlas_email_provider" "test" {
  force_destroy = %t
  smtp = {
    host     = "smtp.example.com"
    port     = 587
    username = "mailer"
    password = "not-a-real-password"
  }
}
`, force)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEmailProviderDestroyed,
		Steps: []resource.TestStep{
			{
				Config: cfg(false),
				Check:  resource.TestCheckResourceAttr("atlas_email_provider.test", "force_destroy", "false"),
			},
			{
				Config: cfg(true),
				Check:  resource.TestCheckResourceAttr("atlas_email_provider.test", "force_destroy", "true"),
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

func testAccCheckEmailProviderDestroyed(_ *terraform.State) error {
	_, err := testAccClient().GetMessagingProvider(context.Background(), "email")
	if err == nil {
		return fmt.Errorf("the email provider still exists after destroy")
	}
	if !client.IsNotFound(err) {
		return err
	}
	return nil
}
