# atlas_connection_action

Define or replace a custom action: data only (an input schema, a request template, a response mapping), never code. A definition with a first-party action's name replaces it for this instance.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `PUT /v1/connection_actions/{provider}/{name}` (create and update); `GET /v1/connection_actions/{provider}/{name}` (read); `DELETE /v1/connection_actions/{provider}/{name}` (destroy).

## Example Usage

```hcl
resource "atlas_connection_action" "example" {
  provider_name = "tf-connection-action"
  action_name = "tf-connection-action"
  description = "tf-connection-action"
  request = jsonencode({})
}
```

## Schema

### Required

- `provider_name` (String) The provider name this object belongs to. Changing it forces replacement.
- `action_name` (String) The action name this object belongs to. Changing it forces replacement.
- `description` (String) The `description` field.
- `request` (String, JSON) The `request` field. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift.

### Optional

- `input` (String, JSON) JSON Schema (type object) of the input. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift.
- `response` (String, JSON) The `response` field. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift.
- `scopes` (List of String) The `scopes` field.
- `paginate` (String, JSON) The `paginate` field. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift.
- `mutating` (Boolean) The `mutating` field.

### Read-Only

- `id` (String) Atlas object id.

## Import

The import id is `<provider_name>/<action_name>`.

```sh
terraform import atlas_connection_action.example <provider_name>/<action_name>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
