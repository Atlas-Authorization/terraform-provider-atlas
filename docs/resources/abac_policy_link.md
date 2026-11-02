# atlas_abac_policy_link

Link a template to entities. Every slot the template uses needs its entity (`principal` for `?principal`, `resource` for `?resource`); one it does not use is refused. The link is evaluated together with the other enabled policies, always from the template’s current source, and is type-checked against the entity schema when one is set.

Scopes: `abac:read` to read, `abac:write` to write.

Backend API: `POST /v1/abac_policy_links` (create); `GET /v1/abac_policy_links/{id}` (read); `PATCH /v1/abac_policy_links/{id}` (update); `DELETE /v1/abac_policy_links/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_abac_policy_link" "example" {
  template_id = "tf-abac-policy-link"
  name = "tf-abac-policy-link"
}
```

## Schema

### Required

- `template_id` (String) The `template_id` field.
- `name` (String) Shown as the deciding policy in decisions.

### Optional

- `principal` (String, JSON) An entity such as `{ "type": "User", "id": "alice" }`. A JSON document (use `jsonencode`).
- `resource` (String, JSON) An entity such as `{ "type": "User", "id": "alice" }`. A JSON document (use `jsonencode`).
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `template_name` (String) The `template_name` field.
- `template_version` (Number) The template version the link currently follows (always the latest).
- `source` (String) The policy source the link evaluates: the template with its slots filled.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_abac_policy_link.example <id>
```
