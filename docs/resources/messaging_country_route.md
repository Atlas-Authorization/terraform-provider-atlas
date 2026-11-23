# atlas_messaging_country_route

Route SMS to a country through a provider. `provider_id` is the id of your SMS provider or one of its fallback providers; the other providers still back it up when failover is on. Replaces an earlier route for the country.

Scopes: `messaging:read` to read, `messaging:write` to write.

Backend API: `PUT /v1/messaging/country_routes` (create); `GET /v1/messaging/country_routes` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/messaging/country_routes/{country}` (destroy).

## Example Usage

```hcl
resource "atlas_messaging_country_route" "example" {
  country = "tf-messaging-country-route"
  provider_id = "tf-messaging-country-route"
}
```

## Schema

### Required

- `country` (String) The `country` field.
- `provider_id` (String) The `provider_id` field.

### Read-Only

- `id` (String) Atlas object id.
- `channel` (String) The `channel` field.
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_messaging_country_route.example <id>
```
