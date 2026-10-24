# atlas_privacy_orchestrator

The data-rights orchestrator (singleton). api_key_wo needs Terraform 1.11+.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_privacy_orchestrator" "this" {
  provider_id            = "cookiemunch"
  api_key_wo             = var.orchestrator_key
  api_key_wo_version     = 1
  sign_requests          = true
  completion_webhook_url = "https://app.example.com/privacy/done"
}
```
