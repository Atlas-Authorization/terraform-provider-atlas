# atlas_sms_template

Set an SMS body template. A body dropping the code placeholder is refused.

Scopes: `sms_templates:read` to read, `sms_templates:write` to write.

Backend API: `POST /v1/sms_templates` (create); `GET /v1/sms_templates/{name}` (read); `PATCH /v1/sms_templates/{name}` (update); `DELETE /v1/sms_templates/{name}` (destroy).

## Example Usage

```hcl
resource "atlas_sms_template" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field. Changing it forces replacement.
- `body` (String) The `body` field.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `default` (String) The `default` field.
- `customised` (Boolean) The `customised` field.
- `variables` (List of String) The `variables` field.

## Import

The import id is `<id>`.

```sh
terraform import atlas_sms_template.example <id>
```
