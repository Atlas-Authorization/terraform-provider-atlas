# atlas_partner_link

An inbound partner link: a credential that lets another product call this instance with exactly the granted scopes.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_partner_link" "cookiemunch" {
  partner = "cookiemunch"
  scopes  = ["privacy:execute"]
}
```
