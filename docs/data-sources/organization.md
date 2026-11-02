# atlas_organization (data source)

An organization in the instance, looked up by id.

Every attribute except `id` is computed. Scope: `organizations:read`.

## Example Usage

```hcl
data "atlas_organization" "example" {
  id = "<id>"
}
```

## Schema

### Required

- `id` (String) The id to look up.

### Read-Only

- `name` (String) The `name` field.
- `slug` (String) The `slug` field.
- `image_url` (String) The `image_url` field.
- `parent_organization_id` (String) The `parent_organization_id` field.
- `public_metadata` (String, JSON) The `public_metadata` field. A JSON document (use `jsonencode`).
- `max_allowed_memberships` (Number) The `max_allowed_memberships` field.
- `created_by` (String) The user seated at creation; null for an organization your backend created with no creator.
- `external_id` (String) Your own id for this customer; unique per instance.
- `role_mappings` (String, JSON) Role key → your product's name for it (e.g. {"org:admin": "owner"}); a member's mapped name rides in the org_role_mapped token claim. A JSON document (use `jsonencode`).
- `owner_invitation` (String, JSON) On create with owner_email only: the invitation sent to that address. A JSON document (use `jsonencode`).
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).
- `members` (String, JSON) The `members` field. A JSON document (use `jsonencode`).
- `entitlements` (String, JSON) The org's effective entitlement (plan/features/limits), or null when none. A JSON document (use `jsonencode`).
- `sso` (String, JSON) The organization's enterprise sign-in setup. A JSON document (use `jsonencode`).
