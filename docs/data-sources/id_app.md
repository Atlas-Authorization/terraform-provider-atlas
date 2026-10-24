# atlas_id_app (Data Source)

Look up an existing **Sign in with Atlas** app by id. Needs the provider's platform key (`psk_...`). The client secret is never returned.

API: `GET /v1/platform/atlas_id_apps/:id` and `GET /v1/platform/atlas_id_apps/:id/preflight`.

## Example Usage

```hcl
data "atlas_id_app" "acme" {
  id = "oac_123"
}

output "button_origins" {
  value = data.atlas_id_app.acme.button_rendered_origins[*].origin
}
```

## Schema

### Required

- `id` (String) — The app id.

### Read-Only

- `client_id`, `name`, `type`, `homepage_url`, `secret_prefix`
- `redirect_uris`, `allowed_scopes` (List of String)
- `managed_by_instance_id` (String) — The instance that manages the app's grants.
- `managed_via` (String) — `instance_toggle`, `adopted`, or unset.
- `button_rendered_origins` (List of Object) — Page origins the Sign in with Atlas button script ran on in the last 7 days: `origin`, `last_seen_at` (epoch ms), `requests`.
- `created_at`, `updated_at` (Number) — Epoch ms.
