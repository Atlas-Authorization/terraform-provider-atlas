# atlas_scim_token

Mint a SCIM provisioning token for an organization. The token is revealed exactly once. scopes ["read"] makes it read-only; role_mappings maps the SCIM user roles attribute to organization roles.

Scopes: `scim:read` to read, `scim:write` to write.

Backend API: `POST /v1/scim_tokens` (create); `GET /v1/scim_tokens` (read: the list is scanned for the id); `PATCH /v1/scim_tokens/{id}` (update); `POST /v1/scim_tokens/{id}/revoke` (destroy).

## Example Usage

```hcl
resource "atlas_scim_token" "example" {
  organization_id = "tf-scim-token"
}
```

## Schema

### Required

- `organization_id` (String) The organization this token provisions into. Changing it forces replacement.

### Optional

- `name` (String) A human label for the token.
- `connection_id` (String) Optional — bind to an SSO connection. Changing it forces replacement.
- `scopes` (List of String) ["read"] makes a read-only token (list and fetch; every write is refused with 403). Omitted or null = read and write.
- `role_mappings` (String, JSON) Maps the SCIM user roles attribute to the member's organization role (primary value first). A value that is itself a role key applies without a mapping. A user whose roles map to nothing keeps their role (org:member when first seated). A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `prefix` (String) Recognisable prefix; never the whole token.
- `token` (String, Sensitive) The bearer token + SCIM base URL — returned ONCE, on create only. Store it now. Returned once, on create.
- `last_used_at` (Number) The `last_used_at` field. (epoch ms)
- `expires_at` (Number) The `expires_at` field. (epoch ms)
- `revoked_at` (Number) The `revoked_at` field. (epoch ms)
- `created_at` (Number) Created (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_scim_token.example <id>
```
