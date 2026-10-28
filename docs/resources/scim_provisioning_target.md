# atlas_scim_provisioning_target

Create an outbound SCIM provisioning target, optionally from a preset (preset + preset_params instead of base_url). base_url must be https; the downstream bearer_token is write-only; the cursor starts at the current event so only new users forward (use sync_user to backfill).

Scopes: `scim_provisioning:read` to read, `scim_provisioning:write` to write.

Backend API: `POST /v1/scim_provisioning_targets` (create); `GET /v1/scim_provisioning_targets/{id}` (read); `PATCH /v1/scim_provisioning_targets/{id}` (update); `DELETE /v1/scim_provisioning_targets/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_scim_provisioning_target" "example" {
  name = "tf-scim-provisioning-target"
}
```

## Schema

### Required

- `name` (String) A label for this target.

### Optional

- `kind` (String) scim (default): any SCIM 2.0 server. google_workspace: Directory API users and groups through a service account. microsoft_graph: Entra ID / 365 users, groups and licenses. rest: any HTTP API, described with request templates. ad_writeback: Active Directory through a relay agent. sql: your parameterized statements against a database, through a relay agent. Changing it forces replacement.
- `connector_config` (String, JSON) The kind's non-secret settings, validated per kind. google_workspace: admin_email, domain, org_unit_path. microsoft_graph: tenant_id, client_id, usage_location, managed_licenses, disabled_plans. rest: user/group operations (method, path, body with {{ expression }} placeholders), remote_id_path, auth, headers. ad_writeback: relay_agent_id, url (ldaps), users_dn, groups_dn, bind_dn, upn_suffix. sql: relay_agent_id, driver, host, port, database, user, statements. Any kind: lookups (named maps for expressions). A JSON document (use `jsonencode`).
- `credentials` (String, JSON, Sensitive) Write-only secrets for a non-SCIM kind, encrypted at rest and never returned: google_workspace {client_email, private_key}; microsoft_graph {client_secret}; rest {token}; ad_writeback {bind_password}; sql {password}. On update, send the whole bundle to replace it; omit it to keep the stored one. A JSON document (use `jsonencode`). Write-only: sent to Atlas, never returned, and not detected as drift.
- `base_url` (String) Downstream base URL (kind scim and rest; required there). MUST be https and a public host — private/loopback/link-local/metadata hosts are refused (SSRF guard). Fixed for google_workspace and microsoft_graph, unused for relay-backed kinds.
- `bearer_token` (String, Sensitive) Required for kind scim. Write-only. Encrypted at rest with the per-instance key and never returned (reads report has_bearer_token). Write-only: sent to Atlas, never returned, and not detected as drift.
- `attribute_mapping` (String, JSON) Optional. Attribute path → Atlas source field (primaryEmail, username, id, firstName, lastName) or an expression starting with = (for example "=lower(user.primaryEmail)" or "=contains(user.groups, \"staff\") ? [\"<license sku>\"] : []"). Expressions read user.* (id, username, firstName, lastName, primaryEmail, emails, active, metadata, groups, organizations, roles) and lookups.*. Any attribute beyond userName, externalId and the name parts sets one more attribute on the payload. Omit for the default mapping. A JSON document (use `jsonencode`).
- `deprovision_action` (String) What to do downstream when an Atlas user is banned/soft-deleted. Default deactivate.
- `enabled` (Boolean) Default true.

### Read-Only

- `id` (String) Atlas object id.
- `has_credentials` (Boolean) Whether a write-only credential bundle is stored (non-SCIM kinds). The values are never returned.
- `has_bearer_token` (Boolean) Whether a downstream bearer is configured. The bearer VALUE is never returned.
- `status` (String) The `status` field. One of: `active`, `paused`.
- `cursor` (String) Event cursor; only users after it forward automatically (backfill with sync_user).
- `consecutive_failures` (Number) The `consecutive_failures` field.
- `last_error` (String) The `last_error` field.
- `last_synced_at` (Number) Epoch ms of the last successful push, or null.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_scim_provisioning_target.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
