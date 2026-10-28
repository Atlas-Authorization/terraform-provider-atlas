# atlas_audit_event_schema

Register a schema for an action: the target types and actor types it allows and a JSON Schema its metadata must satisfy. `version` defaults to the next number for that action; a version is immutable (registering it again is a 409) - change a schema by registering a new version. From then on events with this action are validated on write.

Scopes: `audit_events:read` to read, `audit_events:write` to write.

Backend API: `POST /v1/audit_event_schemas` (create); `GET /v1/audit_event_schemas/{id}` (read); `PATCH /v1/audit_event_schemas/{id}` (update); `DELETE /v1/audit_event_schemas/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_audit_event_schema" "example" {
  action = "tf-audit-event-schema"
}
```

## Schema

### Required

- `action` (String) e.g. invoice.sent Changing it forces replacement.

### Optional

- `version` (Number) The `version` field. Changing it forces replacement.
- `description` (String) The `description` field.
- `target_types` (List of String) The `target_types` field. Changing it forces replacement.
- `actor_types` (List of String) The `actor_types` field. Changing it forces replacement.
- `metadata_schema` (String, JSON) The `metadata_schema` field. A JSON document (use `jsonencode`). Changing it forces replacement.

### Read-Only

- `id` (String) Atlas object id.
- `status` (String) The `status` field. One of: `active`, `deprecated`.
- `created_at` (Number) The `created_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_audit_event_schema.example <id>
```
