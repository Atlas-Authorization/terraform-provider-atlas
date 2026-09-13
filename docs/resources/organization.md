# atlas_organization

An organization (tenant). Creation requires an existing user id as `created_by`,
who becomes the first admin — Terraform manages the org's declarative profile,
not its runtime membership.

Backend API (BAPI): `POST /v1/organizations` (create),
`GET /v1/organizations/:id` (read), `PATCH /v1/organizations/:id` (update),
`DELETE /v1/organizations/:id` (delete).

## Example Usage

```hcl
resource "atlas_organization" "acme" {
  name                    = "Acme Inc"
  slug                    = "acme"
  created_by              = "usr_123"
  max_allowed_memberships = 50
}
```

## Schema

### Required

- `name` (String) — Organization name.
- `slug` (String) — URL-safe slug.
- `created_by` (String) — User id of the initial admin. Immutable — changing it forces replacement.

### Optional

- `max_allowed_memberships` (Number) — Optional seat cap. `0` or omitted means unlimited.

### Read-Only

- `id` (String) — Atlas object id.
- `image_url` (String) — Organization logo URL.
- `created_at` (Number) — Creation time (epoch ms).
- `updated_at` (Number) — Last update time (epoch ms).

## Import

```sh
terraform import atlas_organization.acme org_123
```
