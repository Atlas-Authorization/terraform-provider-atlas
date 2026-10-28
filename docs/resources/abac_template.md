# atlas_abac_template

Create a policy template: policy source whose scope uses the slots `?principal` and/or `?resource` (`principal == ?principal`, `resource in ?resource`, `resource is Doc in ?resource`). Slots are filled by a link. With an entity schema set, the template is type-checked on save (a slot stands for any entity type).

Scopes: `abac:read` to read, `abac:write` to write.

Backend API: `POST /v1/abac_templates` (create); `GET /v1/abac_templates/{id}` (read); `PATCH /v1/abac_templates/{id}` (update); `DELETE /v1/abac_templates/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_abac_template" "example" {
  name = "tf-abac-template"
  source = "tf-abac-template"
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `source` (String) Policy source with `?principal` and/or `?resource` slots in the scope, e.g. `permit (principal == ?principal, action == Action::"read", resource in ?resource);`.

### Optional

- `description` (String) The `description` field.
- `fga_store_id` (String) The FGA store `fga.check(...)` in the source consults.

### Read-Only

- `id` (String) Atlas object id.
- `version` (Number) Increases each time `source` changes; every link follows it.
- `slots` (List of String) The `slots` field. One of: `?principal`, `?resource`.
- `link_count` (Number) The `link_count` field.
- `links_updated` (Number) On an update that changed `source`: the links that now evaluate the new source.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_abac_template.example <id>
```
