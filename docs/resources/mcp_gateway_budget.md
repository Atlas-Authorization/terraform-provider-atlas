# atlas_mcp_gateway_budget

Add a budget. `metric` is `calls` (a count), `cost` (the sum of each call's cost: `cost_per_call`, or the number found at `cost_arg` in the call arguments, such as an amount) or `rate` (calls per `window_seconds`, required). `limit` is the ceiling. `window_seconds` makes the counter reset in fixed windows; without it a counter never resets (a per-session one ends with the session). `scope` is `agent` (each agent counts separately) or `session` (each session of each agent; the client must echo the `Mcp-Session-Id` from initialize). `agent_id` limits it to one agent and `tool` to one tool. A call is charged against every budget it touches at once, or against none; only a call that is allowed and about to run is charged.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/mcp_gateways/{id}/budgets` (create); `GET /v1/mcp_gateways/{id}/budgets/{budgetId}` (read); `PATCH /v1/mcp_gateways/{id}/budgets/{budgetId}` (update); `DELETE /v1/mcp_gateways/{id}/budgets/{budgetId}` (destroy).

## Example Usage

```hcl
resource "atlas_mcp_gateway_budget" "example" {
  mcp_gateway_id = "tf-mcp-gateway-budget"
}
```

## Schema

### Required

- `mcp_gateway_id` (String) The mcp gateway id this object belongs to. Changing it forces replacement.

### Optional

- `name` (String) The `name` field.
- `scope` (String) The `scope` field. One of: `agent`, `session`.
- `metric` (String) The `metric` field. One of: `calls`, `cost`, `rate`.
- `limit` (Number) The ceiling, in calls or cost units.
- `window_seconds` (Number) Fixed window length in seconds; null never resets. Required for rate.
- `agent_id` (String) The `agent_id` field.
- `tool` (String) The `tool` field.
- `cost_per_call` (Number) cost: what one call costs when cost_arg is not set. Default 1.
- `cost_arg` (String) cost: dot path to a number in the call arguments that is the cost of the call.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `gateway_id` (String) The `gateway_id` field.
- `created_at` (Number) The `created_at` field. (epoch ms)

## Import

The import id is `<mcp_gateway_id>/<id>`.

```sh
terraform import atlas_mcp_gateway_budget.example <mcp_gateway_id>/<id>
```
