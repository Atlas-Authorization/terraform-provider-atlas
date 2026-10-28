# atlas_network_zone

Create a network zone. `name` is unique per instance; `cidrs`, `countries` (ISO alpha-2), `asns` and `anonymizers` are unioned (any match = in the zone). Malformed entries are refused.

Scopes: `network_acls:read` to read, `network_acls:write` to write.

Backend API: `POST /v1/network_zones` (create); `GET /v1/network_zones/{id}` (read); `PATCH /v1/network_zones/{id}` (update); `DELETE /v1/network_zones/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_network_zone" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `description` (String) The `description` field.
- `cidrs` (List of String) The `cidrs` field.
- `countries` (List of String) The `countries` field.
- `asns` (String, JSON) The `asns` field. A JSON document (use `jsonencode`).
- `anonymizers` (List of String) The `anonymizers` field. One of: `tor`, `datacenter`, `vpn`.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_network_zone.example <id>
```
