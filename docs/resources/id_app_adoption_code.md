# atlas_id_app_adoption_code

Issue a one-time code that lets an instance of the account adopt this app with its own secret key.

Scopes: `platform:atlas_id_apps:read` to read, `platform:atlas_id_apps:write` to write.

Backend API: `POST /v1/platform/atlas_id_apps/{id}/adoption_codes` (create); `GET /v1/platform/atlas_id_apps/{id}/adoption_codes` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/platform/atlas_id_apps/{id}/adoption_codes/{codeId}` (destroy).

## Example Usage

```hcl
resource "atlas_id_app_adoption_code" "example" {
  atlas_id_app_id = "tf-atlas-id-app-adoption-code"
}
```

## Schema

### Required

- `atlas_id_app_id` (String) The atlas id app id this object belongs to. Changing it forces replacement.

### Optional

- `instance_id` (String) The one instance of the account that may redeem the code. Not returned by the API, so it is not detected as drift.
- `expires_in_seconds` (Number) The `expires_in_seconds` field. Not returned by the API, so it is not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `app_id` (String) The `app_id` field.
- `code` (String) Shown once, when issued.
- `target_instance_id` (String) The `target_instance_id` field.
- `status` (String) The `status` field. One of: `open`, `redeemed`, `expired`. A JSON document (use `jsonencode`).
- `expires_at` (Number) The `expires_at` field. (epoch ms)
- `redeemed_at` (Number) The `redeemed_at` field. (epoch ms)
- `redeemed_by_instance_id` (String) The `redeemed_by_instance_id` field.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `note` (String) The `note` field.

## Import

The import id is `<atlas_id_app_id>/<id>`.

```sh
terraform import atlas_id_app_adoption_code.example <atlas_id_app_id>/<id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
