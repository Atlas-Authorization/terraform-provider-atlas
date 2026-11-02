# atlas_abac_token_mapping

Set the claim mapping used by `authorize_token`: the principal type and id claim, claims that become parent entities (org, role, groups), claims copied as attributes, and trusted issuers with their JWKS URLs.

Scopes: `abac:read` to read, `abac:write` to write.

Backend API: `PUT /v1/abac/token_mapping` (create and update); `GET /v1/abac/token_mapping` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_abac_token_mapping" "example" {

}
```

## Schema

### Optional

- `principal_type` (String) The `principal_type` field.
- `id_claim` (String) The `id_claim` field.
- `parents` (String, JSON) The `parents` field. A JSON document (use `jsonencode`).
- `attribute_claims` (List of String) The `attribute_claims` field.
- `trusted_issuers` (String, JSON) The `trusted_issuers` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_abac_token_mapping.example abac_token_mapping
```
