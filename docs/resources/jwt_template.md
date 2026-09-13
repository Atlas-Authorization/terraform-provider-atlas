# atlas_jwt_template

A JWT template: a named map of custom claims injected into issued tokens. The
`name` is the key and is immutable; only the `claims` map is updatable.

Backend API (BAPI): `POST /v1/jwt_templates` (create),
`GET /v1/jwt_templates/:id` (read), `PATCH /v1/jwt_templates/:id` (update),
`DELETE /v1/jwt_templates/:id` (delete). The `:id` path segment is the template
name.

## Example Usage

```hcl
resource "atlas_jwt_template" "supabase" {
  name = "supabase"
  claims = {
    role         = "authenticated"
    user_email   = "{{user.primary_email}}"
    "aud"        = "authenticated"
  }
}
```

## Schema

### Required

- `name` (String) — Template name — the key. Immutable; changing it forces replacement.
- `claims` (Map of String) — Claim name to value-template map. Reserved system claims are rejected by the API.

### Read-Only

- `id` (String) — Equal to the template name.

## Import

```sh
terraform import atlas_jwt_template.supabase supabase
```
