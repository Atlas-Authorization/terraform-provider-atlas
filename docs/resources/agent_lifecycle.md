# atlas_agent_lifecycle

Set an agent's sponsor, backup sponsor and expiry (omitted fields keep their value; null clears). Emits agent.sponsor_changed. When the sponsor's account is gone, banned or locked the backup is promoted automatically; at the expiry the agent's grants are revoked and the agent is revoked, after an agent.expiring warning 14 days ahead.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `PUT /v1/agents/{id}/lifecycle` (create and update); `GET /v1/agents/{id}/lifecycle` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_agent_lifecycle" "example" {
  agent_id = "tf-agent-lifecycle"
}
```

## Schema

### Required

- `agent_id` (String) The agent id this object belongs to. Changing it forces replacement.

### Optional

- `sponsor_user_id` (String) The `sponsor_user_id` field.
- `backup_sponsor_user_id` (String) The `backup_sponsor_user_id` field.
- `expires_at` (Number) Epoch ms or ISO date, in the future.

### Read-Only

- `id` (String) Atlas object id.
- `renewed_at` (Number) The `renewed_at` field. (epoch ms)
- `renewal_count` (Number) The `renewal_count` field.

## Import

The import id is `<agent_id>`.

```sh
terraform import atlas_agent_lifecycle.example <agent_id>
```
