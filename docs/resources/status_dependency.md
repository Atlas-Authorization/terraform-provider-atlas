# atlas_status_dependency

An upstream Statuspage-compatible status page this instance depends on.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_status_dependency" "email" {
  name = "Email provider"
  url  = "https://status.email-provider.example"
}
```
