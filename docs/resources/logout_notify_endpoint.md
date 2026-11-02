# atlas_logout_notify_endpoint

Register a logout fan-out endpoint. It is POSTed `logout_token=<jwt>` (application/x-www-form-urlencoded; OIDC Back-Channel Logout: iss, sub, sid, events, signed with the instance JWKS) whenever an Atlas session ends, plus one with sub and no sid when the user is banned, deleted or deprovisioned. With oauth_client_id it follows one OAuth client instead (that client's sign-in ending, incl. /oauth2/session/end and grant revocation). Retried on network errors, 5xx, 408 and 429; redirects are not followed. The URL must be public https.

Scopes: `logout_notify_endpoints:read` to read, `logout_notify_endpoints:write` to write.

Backend API: `POST /v1/logout_notify_endpoints` (create); `GET /v1/logout_notify_endpoints/{id}` (read); `PATCH /v1/logout_notify_endpoints/{id}` (update); `DELETE /v1/logout_notify_endpoints/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_logout_notify_endpoint" "example" {
  url = "https://example.com/url"
}
```

## Schema

### Required

- `url` (String) Public https endpoint.

### Optional

- `oauth_client_id` (String) A client_id of this instance to follow; omit for instance-wide. Changing it forces replacement.
- `audience` (String) The logout_token `aud`. Default: the client_id when following a client, else the instance id.
- `description` (String) The `description` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_logout_notify_endpoint.example <id>
```
