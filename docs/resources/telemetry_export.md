# atlas_telemetry_export

Configure OTLP/HTTP export of gateway tool calls and relay calls as spans and metrics (atlas.calls, atlas.call.duration) to your collector. The endpoint must be https and public; headers (for example an API token) are sealed and never returned. Delivery is batched and best effort.

Scopes: `instance:read` to read, `instance:write` to write.

Backend API: `PUT /v1/telemetry_export` (create and update); `GET /v1/telemetry_export` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_telemetry_export" "example" {

}
```

## Schema

### Optional

- `enabled` (Boolean) The `enabled` field.
- `endpoint` (String) The `endpoint` field.
- `headers` (String, JSON, Sensitive) Up to 10 request headers; null clears them; omit to keep. A JSON document (use `jsonencode`). Write-only: sent to Atlas, never returned, and not detected as drift.
- `service_name` (String) The `service_name` field.
- `traces` (Boolean) The `traces` field.
- `metrics` (Boolean) The `metrics` field.

### Read-Only

- `id` (String) Atlas object id.
- `has_headers` (Boolean) Headers (collector tokens) are write-only.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_telemetry_export.example telemetry_export
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
