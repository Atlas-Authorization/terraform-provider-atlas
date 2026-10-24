# atlas_secret_key

A scoped runtime secret key. The secret is returned once and kept (sensitive) in state; destroy revokes the key.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_secret_key" "ci" {
  name            = "ci"
  scopes          = ["users:read", "webhooks:write"]
  owner           = "platform@example.com"
  expires_at      = "2027-06-01T00:00:00Z"
  allowed_ips     = ["203.0.113.0/24"]
  allowed_origins = ["https://app.example.com"]
  external_ref    = "ci-key"
  metadata        = { team = "platform" }
}
```
