# atlas_iga_bundle

Create a bundle: several catalog entitlements, request form fields, and a policy (who may ask, guest expiry).

Scopes: `iga:read` to read, `iga:write` to write.

Backend API: `POST /v1/iga/bundles` (create); `GET /v1/iga/bundles/{id}` (read); `PATCH /v1/iga/bundles/{id}` (update); `DELETE /v1/iga/bundles/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_iga_bundle" "example" {
  key = "tf-iga-bundle"
  name = "tf-iga-bundle"
  entitlements = ["example"]
}
```

## Schema

### Required

- `key` (String) The `key` field.
- `name` (String) The `name` field.
- `entitlements` (List of String) Entitlement keys or ids (1 to 25). Not returned by the API, so it is not detected as drift.

### Optional

- `description` (String) The `description` field.
- `catalog` (String) The `catalog` field.
- `form` (String, JSON) The `form` field. A JSON document (use `jsonencode`).
- `policy` (String, JSON) The `policy` field. A JSON document (use `jsonencode`).
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `entitlement_ids` (List of String) The `entitlement_ids` field.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_iga_bundle.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
