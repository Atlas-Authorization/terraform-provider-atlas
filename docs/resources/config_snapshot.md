# atlas_config_snapshot

Save the instance's configuration as it is now, with an optional label. Secrets are removed from the copy (their paths are listed in redacted_paths). The worker also saves a snapshot automatically after configuration changes, so you can go back to a state you never named. Needs instance:write. Honours Idempotency-Key. Audited.

Scopes: `instance:read` to read, `instance:write` to write.

Backend API: `POST /v1/instance/config_snapshots` (create); `GET /v1/instance/config_snapshots/{id}` (read); no update route: every change replaces the object; `DELETE /v1/instance/config_snapshots/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_config_snapshot" "example" {

}
```

## Schema

### Optional

- `label` (String) The `label` field.

### Read-Only

- `id` (String) Atlas object id.
- `version` (Number) 1, 2, 3 ... per instance.
- `source` (String) The `source` field. One of: `manual`, `auto`, `pre_restore`, `cli`, `clone`.
- `sha256` (String) Digest of the stored configuration and origins; two snapshots with the same digest are identical.
- `redacted_paths` (List of String) Where a secret was removed.
- `created_by` (String) The `created_by` field.
- `created_at` (Number) Epoch ms.
- `auth_config` (String, JSON) The configuration (only on the single-snapshot read). A JSON document (use `jsonencode`).
- `allowed_origins` (List of String) The allowed origins (only on the single-snapshot read).

## Import

The import id is `<id>`.

```sh
terraform import atlas_config_snapshot.example <id>
```
