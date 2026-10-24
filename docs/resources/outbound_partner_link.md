# atlas_outbound_partner_link

The credential a partner granted this instance, so Atlas can call the partner.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_outbound_partner_link" "cookiemunch" {
  partner       = "cookiemunch"
  client_id     = var.cm_client_id
  client_secret = var.cm_client_secret
  scopes        = ["dsar:write"]
}
```
