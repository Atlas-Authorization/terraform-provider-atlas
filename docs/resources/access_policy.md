# atlas_access_policy

Create a conditional-access policy. `conditions` (applications, userIds, groups, orgs, roles, networkZones {include, exclude}, countries, riskLevels, deviceTrust, authMethods, timeWindow) are ANDed; an empty condition matches anything. `grants` are allow, block, require_mfa, require_acr, require_policy_acceptance or require_action. `stages` limits it to sign_in, authorize and/or token (default all). Further conditions: devicePlatforms, osVersion {min, max}, clientTypes, userAttributes, userRiskLevels, signInRiskLevels and principalIds. Further grants: require_auth_strength (mfa, passwordless, phishing_resistant), require_device_bound_session, require_password_change, sign_in_frequency, max_session_age and persistent_session. `mode: report_only` logs `access_policy.would_block` and never denies. `targets` picks user, machine and/or agent principals (default user). A matching block wins; step-up grants are unioned.

Scopes: `access_policies:read` to read, `access_policies:write` to write.

Backend API: `POST /v1/access_policies` (create); `GET /v1/access_policies/{id}` (read); `PATCH /v1/access_policies/{id}` (update); `DELETE /v1/access_policies/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_access_policy" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `description` (String) The `description` field.
- `enabled` (Boolean) The `enabled` field.
- `priority` (Number) The `priority` field.
- `stages` (List of String) The `stages` field. One of: `sign_in`, `authorize`, `token`.
- `conditions` (String, JSON) The `conditions` field. A JSON document (use `jsonencode`).
- `grants` (String, JSON) The `grants` field. A JSON document (use `jsonencode`).
- `mode` (String) The `mode` field. One of: `enforce`, `report_only`.
- `targets` (List of String) The `targets` field. One of: `user`, `machine`, `agent`.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_access_policy.example <id>
```
