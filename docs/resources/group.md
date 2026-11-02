# atlas_group

Create an instance group. Emits group.created. → 201.

Scopes: `users:read` to read, `users:write` to write.

Backend API: `POST /v1/groups` (create); `GET /v1/groups/{id}` (read); `PATCH /v1/groups/{id}` (update); `DELETE /v1/groups/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_group" "example" {
  slug = "tf-group"
  name = "tf-group"
}
```

## Schema

### Required

- `slug` (String) Lowercase letters, digits, - and _ (max 63). Unique per instance; usable in place of the id in paths.
- `name` (String) The `name` field.

### Optional

- `description` (String) The `description` field.
- `parent_id` (String) Nest under another group (up to 8 levels, no cycles). Members of a child are also members of its ancestors for the token claim and role mapping.
- `attributes` (String, JSON) Free attributes (max 8 KB). A JSON document (use `jsonencode`).
- `app_role_keys` (List of String) App roles every member (and member of a descendant) holds, merged into app_roles / app_permissions on the token.
- `include_in_token` (Boolean) Add the slug to the `groups` token claim (default true).
- `dynamic_filter` (String, JSON) A user-search filter tree. POST …/reapply (and each SSO sign-in) makes membership match it. A JSON document (use `jsonencode`).
- `sso_match` (String, JSON) { claim, equals | contains | matches }: join at SSO sign-in when the IdP claim (a string or list, e.g. "groups") matches; leave when it stops matching. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `member_count` (Number) The `member_count` field.
- `source` (String) GET /v1/users/:id/groups only: how the user joined.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_group.example <id>
```
