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

## Site and partner link

- `site_id`: the orchestrator-side site id a case is for. Unset, Atlas falls back to the site of a matching integration or outbound partner link; a case with none is abandoned with an error naming `site_id`.
- `partner_link_id`: the id of the partner link the orchestrator's owner granted this instance. Setting it makes Atlas call the orchestrator with that grant and turns on `use_partner_link`. Removing it clears it.
- `cancel_deletion_on_sign_in`: cancel a pending account deletion when the person signs in again (the instance setting `compliance.retention.cancelDeletionOnSignIn`). Unset keeps the current setting.
