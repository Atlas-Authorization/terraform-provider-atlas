# atlas_machine

Create a named machine (M2M) identity; returns the machine secret (mk_) ONCE.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/machines` (create); `GET /v1/machines/{id}` (read); `PATCH /v1/machines/{id}` (update); `DELETE /v1/machines/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_machine" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `owner_user_id` (String) The `owner_user_id` field. Changing it forces replacement.
- `data` (String, JSON) The `data` field. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift. Changing it forces replacement.

### Read-Only

- `id` (String) Atlas object id.
- `organization_id` (String) Owning org, when org-scoped.
- `status` (String) Stored device lifecycle state (unchanged by the owner's state). One of: `active`, `pending`, `revoked`.
- `owner_usable` (Boolean) Whether the owning user is usable (not banned, locked, or deleted). True when the machine has no owner. When false the machine cannot mint or verify tokens even while its stored status stays 'active'.
- `effective_status` (String) Derived status: 'active' only when the machine is active AND its owner is usable; 'owner_suspended' when active but the owner is not; otherwise the raw status.
- `device_key` (String) The stable device self-identifier supplied at enrolment (null when none). Lets a fleet manager correlate this record with the device it re-enrolled across tokens.
- `public_key_jkt` (String) RFC 7638 thumbprint of the device public key.
- `metadata` (String, JSON) Device metadata (hostname, os, app_version, last_seen_at). A JSON document (use `jsonencode`).
- `enrolled_at` (Number) When the device completed enrolment (epoch ms).
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).
- `secret` (String, Sensitive) Reveal-once mk_ secret (create/rotate-only). Returned once, on create.

## Import

The import id is `<id>`.

```sh
terraform import atlas_machine.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
