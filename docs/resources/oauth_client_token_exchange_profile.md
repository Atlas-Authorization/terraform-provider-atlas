# atlas_oauth_client_token_exchange_profile

Register a custom subject_token_type (a urn: or https URI that is not a standard Atlas or IETF token type) for a confidential client. An exchange with it runs your `custom_token_exchange` Action, which validates the token and names the user to issue tokens for.

Scopes: `oauth_clients:read` to read, `oauth_clients:write` to write.

Backend API: `POST /v1/oauth_clients/{id}/token_exchange_profiles` (create); `GET /v1/oauth_clients/{id}/token_exchange_profiles` (read: the list is scanned for the id); `PATCH /v1/oauth_clients/{id}/token_exchange_profiles/{profileId}` (update); `DELETE /v1/oauth_clients/{id}/token_exchange_profiles/{profileId}` (destroy).

## Example Usage

```hcl
resource "atlas_oauth_client_token_exchange_profile" "example" {
  client_id = "tf-oauth-client-token-exchange-profile"
}
```

## Schema

### Required

- `client_id` (String) The client id this object belongs to. Changing it forces replacement.

### Optional

- `name` (String) The `name` field.
- `subject_token_type` (String) The `subject_token_type` field. Changing it forces replacement.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<client_id>/<id>`.

```sh
terraform import atlas_oauth_client_token_exchange_profile.example <client_id>/<id>
```
