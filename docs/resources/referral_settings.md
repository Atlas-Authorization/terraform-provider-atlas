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
}
```
