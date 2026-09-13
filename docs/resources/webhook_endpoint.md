# atlas_webhook_endpoint

A webhook endpoint that receives signed Atlas events. The signing secret
(`whsec_...`) is revealed once, on create, and stored (sensitive) in state. The
Backend API has **no update route**, so any change to `url` or `enabled_events`
replaces the endpoint (and re-issues the secret).

Backend API (BAPI): `POST /v1/webhook_endpoints` (create),
`GET /v1/webhook_endpoints` (list — the read filters this list by id, since there
is no get-by-id route), `DELETE /v1/webhook_endpoints/:id` (delete).

## Example Usage

```hcl
resource "atlas_webhook_endpoint" "events" {
  url            = "https://hooks.example.com/atlas"
  enabled_events = ["user.created", "organization.updated"]
}
```

## Schema

### Required

- `url` (String) — HTTPS URL that receives event deliveries. Immutable — forces replacement.

### Optional

- `enabled_events` (List of String) — Event types to deliver, or `["*"]` for all. Immutable — forces replacement.

### Read-Only

- `id` (String) — Atlas object id.
- `secret` (String, Sensitive) — The signing secret, revealed only on create.
- `active` (Boolean) — Whether the endpoint is currently active.
- `disabled_at` (Number) — When the endpoint was disabled (epoch ms), if it has been.
- `created_at` (Number) — Creation time (epoch ms).

## Import

```sh
terraform import atlas_webhook_endpoint.events whe_123
```

The one-time `secret` is not recovered on import.
