# atlas_iga_entitlement

A requestable item in the identity-governance catalog: a role, organization
membership, group, resource role, installed-app assignment or custom item, with
an owner, a risk level and approval rules.

Backend API (BAPI): `POST /v1/iga/entitlements` (create), `GET /v1/iga/entitlements/:id`
(read), `PATCH /v1/iga/entitlements/:id` (update), `DELETE /v1/iga/entitlements/:id`
(delete). The API key needs the `iga:read` and `iga:write` scopes.

## Example Usage

```hcl
resource "atlas_iga_entitlement" "prod_admin" {
  key           = "prod-admin"
  name          = "Production admin"
  kind          = "role"
  target        = jsonencode({ role_key = "admin" })
  owner_user_id = "user_01J8EXAMPLE"
  risk_level    = "high"

  approval_rules = jsonencode({
    steps = [
      { type = "manager", mode = "any", timeout_seconds = 86400, escalate_to = { type = "owner" } },
      { type = "group", ids = ["secops"], mode = "all" },
    ]
  })

  max_duration_seconds = 28800
}
```

## Argument Reference

- `key` (Required, forces replacement) Stable key: lowercase letters, digits and `. _ : -`.
- `name` (Required) Display name.
- `description` (Optional) What the access gives.
- `kind` (Required, forces replacement) `role`, `org_membership`, `group`, `resource_role`, `app_assignment` or `custom`.
- `target` (Required) What it points at, as JSON: `role` `{ role_key }`, `org_membership` `{ organization_id, role }`, `group` `{ group }`, `resource_role` `{ resource_id, role }`, `app_assignment` `{ app_id }`; free-form for `custom`. The target must exist.
- `owner_user_id` (Optional) The owning user. Required when an approval step or escalation uses `owner`.
- `risk_level` (Optional) `low` (default), `medium`, `high` or `critical`.
- `approval_rules` (Optional) `{ steps = [...] }`; up to five steps, each `type` (`manager`, `owner`, `group` with one slug in `ids`, `users` with user ids), `mode` (`any` or `all`), optional `timeout_seconds` and `escalate_to`. No steps means approved at once.
- `max_duration_seconds` (Optional) The longest a grant may last; omit for until revoked.
- `requestable` (Optional) Whether users can request it themselves. Defaults to `true`.
- `require_justification` (Optional) Whether a request must state a reason. Defaults to `true`.

## Attribute Reference

- `id` The entitlement id.

## Import

```shell
terraform import atlas_iga_entitlement.prod_admin ige_…
```
