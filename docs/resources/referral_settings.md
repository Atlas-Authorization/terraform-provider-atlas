# atlas_referral_settings

Referral capture and attribution settings (singleton). *_wo secrets need Terraform 1.11+.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_referral_settings" "this" {
  enabled         = true
  provider_id     = "referalli"
  program_id      = "prog_123"
  context_cookies = ["ref", "_rl"]
  consent_provider = "atlasconsent"

  # Linked from the "Refer a friend" panel on the hosted account page.
  partner_portal_url = "https://partners.example.com"

  # Count the first Sign in with Atlas sign-in to an app this instance manages as a referred sign-up.
  report_linked_atlas_id_signups = true
}
```

`partner_portal_url` (String) is the https partner portal where a referrer manages
their partnership. `report_linked_atlas_id_signups` (Boolean, default `false`)
reports the first sign-in of an Atlas ID account to a Sign in with Atlas app this
instance manages as a sign-up; the result arrives as `user.referral_attributed`
with `source = "sign_in_with_atlas"`.
