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
- `role_mappings` (String, JSON) Maps the SCIM user roles attribute, and SCIM group membership, to the member's organization role. roles: primary value first; a value that is itself a role key applies without a mapping. Groups: when a member joins or leaves a group (or a group is renamed or deleted), the first mapping, in this order, that names one of their groups by displayName or externalId sets the role, so list the most privileged first. Only members SCIM manages follow (seated by SCIM; not a role set through the API or dashboard, and never the organization's creator). A member whose roles or groups map to nothing keeps their role (org:member when first seated). A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `prefix` (String) Recognisable prefix; never the whole token.
- `token` (String, Sensitive) The bearer token + SCIM base URL — returned ONCE, on create only. Store it now. Returned once, on create.
- `scim_base_url` (String) On create: the SCIM 2.0 base URL to give the identity provider (on the instance's frontend API host), e.g. https://auth.example.com/scim/v2.
- `warnings` (String, JSON) On create: non-blocking notes. `no_sso_connection` = the organization has no SSO connection yet, so provisioned users cannot sign in through the IdP until one exists. A JSON document (use `jsonencode`).
- `last_used_at` (Number) The `last_used_at` field. (epoch ms)
- `expires_at` (Number) The `expires_at` field. (epoch ms)
- `revoked_at` (Number) The `revoked_at` field. (epoch ms)
- `created_at` (Number) Created (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_scim_token.example <id>
```
