# atlas_device_integration

Connect a device source that Atlas pulls on a schedule: Microsoft Intune (Graph, `tenantId` + app client credentials with DeviceManagementManagedDevices.Read.All), Jamf Pro (`baseUrl` + API client), CrowdStrike Falcon (`cloud` + API client with Zero Trust Assessment and Hosts read; `minZtaScore`, `userEmailDomain`), Fleet (`baseUrl` + API token) or Google Chrome / ChromeOS (service account with domain-wide delegation). Credentials are write-only.

Scopes: `device_assurance:read` to read, `device_assurance:write` to write.

Backend API: `POST /v1/device_integrations` (create); `GET /v1/device_integrations/{id}` (read); `PATCH /v1/device_integrations/{id}` (update); `DELETE /v1/device_integrations/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_device_integration" "example" {

}
```

## Schema

### Optional

- `provider_value` (String) The `provider` field. One of: `intune`, `jamf`, `crowdstrike`, `fleet`, `chrome`.
- `name` (String) The `name` field.
- `config` (String, JSON) The `config` field. A JSON document (use `jsonencode`).
- `credentials` (String, JSON, Sensitive) The `credentials` field. A JSON document (use `jsonencode`). Write-only: sent to Atlas, never returned, and not detected as drift.
- `interval_minutes` (Number) The `interval_minutes` field.
- `status` (String) The `status` field. One of: `active`, `paused`.

### Read-Only

- `id` (String) Atlas object id.
- `has_credentials` (Boolean) The `has_credentials` field.
- `next_sync_at` (Number) The `next_sync_at` field. (epoch ms)
- `last_sync_at` (Number) The `last_sync_at` field. (epoch ms)
- `last_status` (String) The `last_status` field.
- `last_error` (String) The `last_error` field.
- `last_stats` (String, JSON) The `last_stats` field. A JSON document (use `jsonencode`).
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_device_integration.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
