# atlas_webhook_endpoint

A webhook endpoint that receives signed Atlas events. The signing secret
(`whsec_...`) is stored (sensitive) in state: it is returned on create, and on
import it is read back from Atlas. Changing `url`, `enabled_events` or `active`
updates the endpoint **in place** and **preserves the signing secret**, so event
deliveries keep verifying. Change `rotate_secret_trigger` to rotate the secret
in place.

Atlas switches an endpoint off after repeated delivery failures. It then stays
in state with `active = false` (and `failing_since` / `disabled_at` set); set
`active = true` to switch it back on.

Backend API (BAPI): `POST /v1/webhook_endpoints` (create),
`GET /v1/webhook_endpoints` (list — the read filters this list by id, since there
is no get-by-id route), `PATCH /v1/webhook_endpoints/:id` (in-place update — does
not rotate the secret), `POST /v1/webhook_endpoints/:id/rotate_secret`
(rotation), `POST /v1/webhook_endpoints/:id/reveal_secret` (import),
`DELETE /v1/webhook_endpoints/:id` (delete).

## Example Usage

```hcl
resource "atlas_webhook_endpoint" "events" {
  url            = "https://hooks.example.com/atlas"
  enabled_events = ["user.created", "organization.updated"]

  # Change this value to rotate the signing secret.
  rotate_secret_trigger = "2026-10"
}

# A signing secret you choose (Terraform 1.11+), created switched off.
resource "atlas_webhook_endpoint" "partner" {
  url               = "https://partner.example.com/hooks/atlas"
  enabled_events    = ["user.created"]
  active            = false
  secret_wo         = var.partner_webhook_secret # whsec_ + 32-256 characters
  secret_wo_version = 1                          # bump to rotate to a new secret_wo
}
```

## Schema

### Required

- `url` (String) — HTTPS URL that receives event deliveries. Updatable in place (the secret is preserved).

### Optional

- `enabled_events` (List of String) — Event types to deliver, or `["*"]` for all. Updatable in place (the secret is preserved).
- `active` (Boolean) — Whether the endpoint receives deliveries. `false` creates it switched off (or pauses it); `true` switches an auto-disabled endpoint back on.
- `rotate_secret_trigger` (String) — Any value; changing it rotates the signing secret in place. Deliveries signed with the old secret stop verifying at once, so update your verifier.
- `secret_wo` (String, Sensitive, Write-only) — A signing secret you choose: `whsec_` (optionally prefixed `atlas_live_` / `atlas_test_`) followed by 32-256 characters of `A-Za-z0-9+/=_-`. Sent on create and whenever `secret_wo_version` changes (which rotates to it in place). Leave it out to let Atlas generate the secret. Needs Terraform 1.11 or later.
- `secret_wo_version` (Number) — Any number; changing it rotates the signing secret in place to `secret_wo` (or to a new Atlas-generated one when `secret_wo` is unset).

### Read-Only

- `id` (String) — Atlas object id.
- `secret` (String, Sensitive) — The signing secret (from create, rotation, or read back on import).
- `failing_since` (Number) — Start of the current run of failed deliveries (epoch ms), if any.
- `disabled_at` (Number) — When the endpoint was disabled (epoch ms), if it has been.
- `created_at` (Number) — Creation time (epoch ms).

## Import

```sh
terraform import atlas_webhook_endpoint.events whe_123
```

The `secret` is read back from Atlas on import (an audited reveal).

## New in 0.11

`description`, `event_naming`, `external`, `signature_versions`, `alert_contacts`, `destination` + `destination_credentials` (EventBridge, Pub/Sub, Cloudflare Queues; set `url` or `destination`), `external_ref`, `metadata`, and read-only `health`, `last_success_at`, `consecutive_failures`, `last_failing_since`.
