# atlas_workflow

Create a lifecycle workflow. Steps are notify, disable, delete, invite, add_to_org, remove_from_org, add_to_team, remove_from_team, add_to_group, remove_from_group (an instance group by id or slug) and call_action (an Action with the `workflow_step` trigger, which may be an `http` Action), plus the graph steps: if, switch, for_each, wait, set_variable, http_request, call_connector, create_ticket, create_iga_request, start_campaign, revoke_grant, approval and kv. Expressions use `{{ user.email }}`, `event.*`, `input.*`, `vars.*` and `steps.<id>.output`. Each subject runs once per workflow; the first failing step stops the run unless it sets `on_error: continue` or a `retry`.

Scopes: `actions:read` to read, `actions:write` to write.

Backend API: `POST /v1/workflows` (create); `GET /v1/workflows/{id}` (read); `PATCH /v1/workflows/{id}` (update); `DELETE /v1/workflows/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_workflow" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `description` (String) The `description` field.
- `enabled` (Boolean) The `enabled` field.
- `trigger` (String, JSON) The `trigger` field. A JSON document (use `jsonencode`).
- `steps` (String, JSON) The `steps` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `next_run_at` (Number) A cron workflow's next slot, Unix ms; null for any other trigger.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_workflow.example <id>
```
