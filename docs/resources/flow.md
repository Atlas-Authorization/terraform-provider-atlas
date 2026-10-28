# atlas_flow

Create a sign-in journey (version 1, active unless activate is false). Steps run top to bottom once a sign-in has passed its factors: refusals (deny, require_method, require_mfa), required actions the user must finish first, Action calls that may refuse, session claims and the redirect after sign-in. Action steps must name workflow_step Actions. → 201.

Scopes: `actions:read` to read, `actions:write` to write.

Backend API: `POST /v1/flows` (create); `GET /v1/flows/{id}` (read); `PATCH /v1/flows/{id}` (update); `DELETE /v1/flows/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_flow" "example" {
  name = "tf-flow"
  definition = jsonencode({})
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `definition` (String, JSON) The `definition` field. A JSON document (use `jsonencode`).

### Optional

- `description` (String) The `description` field.
- `is_default` (Boolean) The journey for every sign-in without an application-specific one (only one is default).
- `applications` (List of String) OAuth client ids whose sign-ins use this journey instead of the default.
- `enabled` (Boolean) The `enabled` field.
- `activate` (Boolean) Make the new version active (default true). Not returned by the API, so it is not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `active_version` (Number) The `active_version` field.
- `latest_version` (Number) The `latest_version` field.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_flow.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
