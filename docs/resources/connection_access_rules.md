# atlas_connection_access_rules

Replace the connection's access rules (at most 100; an empty list opens the connection again). Rules gate credential use: `read_token` (token reads and refreshes), `relay`, `run_actions` and `sign`; a key that no rule names gets 403 with reason `connection_access_denied`. Managing the connection and its rules stays under organizations:read and organizations:write.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `PUT /v1/connections/{id}/access_rules` (create and update); `GET /v1/connections/{id}/access_rules` (read); `DELETE /v1/connections/{id}/access_rules` (destroy).

## Example Usage

```hcl
resource "atlas_connection_access_rules" "example" {
  connection_id = "tf-connection-access-rules"
  rules = jsonencode({})
}
```

## Schema

### Required

- `connection_id` (String) The connection id this object belongs to. Changing it forces replacement.
- `rules` (String, JSON) The `rules` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `restricted` (Boolean) False while the connection has no rule: every key with the scope may use it.

## Import

The import id is `<connection_id>`.

```sh
terraform import atlas_connection_access_rules.example <connection_id>
```
