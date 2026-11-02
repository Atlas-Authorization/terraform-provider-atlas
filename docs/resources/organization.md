# atlas_organization

An organization (tenant). `created_by` optionally seats an existing user as the
first admin; without it the organization is created by your backend with no
members yet. Terraform manages the org's declarative profile, not its runtime
membership.

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

# Created by your backend, keyed by your own tenant id.
resource "atlas_organization" "globex" {
  name        = "Globex"
  slug        = "globex"
  external_id = "tenant_42"

  public_metadata  = jsonencode({ tier = "gold" })
  private_metadata = jsonencode({ crm_id = "c_981" })

  role_mappings = {
    "org:admin"  = "owner"
    "org:member" = "member"
  }
}
```

## Schema

### Required

- `name` (String) — Organization name.
- `slug` (String) — URL-safe slug.

### Optional

- `created_by` (String) — User id seated as the initial admin. Omit it for an organization created by your backend, which then has no members yet. Immutable — changing it forces replacement.
- `max_allowed_memberships` (Number) — Optional seat cap. `0` or omitted means unlimited.
- `external_id` (String) — Your own id for this customer (1-255 printable characters, no spaces), unique per instance. Removing it clears it.
- `public_metadata` (String) — Public metadata as a JSON object (use `jsonencode`); replaces the stored object. Left unset, Terraform does not manage it and reads back whatever is stored.
- `private_metadata` (String, Sensitive) — Backend-only metadata as a JSON object. Atlas never returns it, so Terraform keeps the configured value and cannot detect drift; removing it clears the stored object.
- `role_mappings` (Map of String) — Atlas role key to your product's name for it (e.g. `"org:admin" = "owner"`); a member's mapped name rides in the `org_role_mapped` token claim. At most 50 roles. Removing it clears the mapping.

### Read-Only

- `id` (String) — Atlas object id.
- `image_url` (String) — Organization logo URL.
- `created_at` (Number) — Creation time (epoch ms).
- `updated_at` (Number) — Last update time (epoch ms).

## Import

```sh
terraform import atlas_organization.acme org_123
```
