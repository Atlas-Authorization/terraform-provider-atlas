# atlas_relay_agent

Register an outbound-only relay agent. The response carries the agent token once; it cannot be read back. An agent can reach only the destinations in allowed_destinations (host, host:port, *.domain).

Scopes: `relay_agents:read` to read, `relay_agents:write` to write.

Backend API: `POST /v1/relay_agents` (create); `GET /v1/relay_agents/{id}` (read); `PATCH /v1/relay_agents/{id}` (update); `DELETE /v1/relay_agents/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_relay_agent" "example" {
  name = "tf-relay-agent"
}
```

## Schema

### Required

- `name` (String) The `name` field.

### Optional

- `organization_id` (String) The `organization_id` field.
- `status` (String) The `status` field. One of: `active`, `revoked`.
- `allowed_destinations` (List of String) host:port destinations the agent may reach.
- `cert_fingerprints` (List of String) Client-certificate pins.

### Read-Only

- `id` (String) Atlas object id.
- `secret_prefix` (String, Sensitive) The `secret_prefix` field. Returned once, on create.
- `online` (Boolean) The `online` field.
- `last_seen_at` (Number) The `last_seen_at` field. (epoch ms)
- `last_ip` (String) The `last_ip` field.
- `agent_version` (String) The `agent_version` field.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `token` (String, Sensitive) On create and rotate only: the agent token, shown once. Returned once, on create.

## Import

The import id is `<id>`.

```sh
terraform import atlas_relay_agent.example <id>
```
