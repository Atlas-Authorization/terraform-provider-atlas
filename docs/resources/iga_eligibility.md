# atlas_iga_eligibility

Make a person eligible for an entitlement: they hold nothing until they activate it. The entitlement needs `activation_rules`. `expires_at` ends the eligibility.

Scopes: `iga:read` to read, `iga:write` to write.

Backend API: `POST /v1/iga/eligibilities` (create); `GET /v1/iga/eligibilities` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/iga/eligibilities/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_iga_eligibility" "example" {
  user_id = "tf-iga-eligibility"
  entitlement = "tf-iga-eligibility"
}
```

## Schema

### Required

- `user_id` (String) The `user_id` field.
- `entitlement` (String) The `entitlement` field. Not returned by the API, so it is not detected as drift.

### Optional

- `expires_at` (Number) The `expires_at` field. (epoch ms)

### Read-Only

- `id` (String) Atlas object id.
- `entitlement_id` (String) The `entitlement_id` field.
- `created_by` (String) The `created_by` field.
- `created_at` (Number) Created (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_iga_eligibility.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
