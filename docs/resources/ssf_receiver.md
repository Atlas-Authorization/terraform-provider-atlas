# atlas_ssf_receiver

Register a transmitter. Its SETs are POSTed (`application/secevent+jwt`) to the returned `receive_endpoint` and verified against `jwks_uri`, `issuer` and the optional `audience`. `actions` maps an event type URI to `revoke_sessions`, `refresh_claims` (invalidate minted tokens so the next refresh carries fresh claims, sessions stay), `ban` or `ignore`; unmapped events are ignored. The default maps session-revoked, credential-change, credential-compromise and device-compliance-change (only when the new status is not-compliant) to `revoke_sessions`, token-claims-change to `refresh_claims` and account-disabled to `ban`.

Scopes: `ssf:read` to read, `ssf:write` to write.

Backend API: `POST /v1/ssf/receivers` (create); `GET /v1/ssf/receivers` (read: the list is scanned for the id); `PATCH /v1/ssf/receivers/{id}` (update); `DELETE /v1/ssf/receivers/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_ssf_receiver" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `issuer` (String) The `issuer` field. Changing it forces replacement.
- `jwks_uri` (String) The `jwks_uri` field.
- `audience` (String) The `audience` field.
- `actions` (String, JSON) The `actions` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.

## Import

The import id is `<id>`.

```sh
terraform import atlas_ssf_receiver.example <id>
```
