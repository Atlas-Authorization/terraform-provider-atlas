# atlas_iga_birthright_rule

A joiner/mover rule: everyone matching `match` holds the listed entitlements,
and loses them when they stop matching. Evaluated when users, memberships and
groups change, and re-checked periodically so directory-sync changes are caught.

Backend API (BAPI): `POST /v1/iga/birthright_rules`, `PATCH /v1/iga/birthright_rules/:id`,
`DELETE /v1/iga/birthright_rules/:id` (and the list for reads). The API key needs
the `iga:read` and `iga:write` scopes.

## Example Usage

```hcl
resource "atlas_iga_birthright_rule" "engineering" {
  name              = "Engineering gets the wiki"
  match             = jsonencode({ group = "engineering" })
  entitlement_ids   = [atlas_iga_entitlement.wiki.id]
  revoke_on_unmatch = true
}
```

## Argument Reference

- `name` (Required) Display name.
- `match` (Required) Conditions as JSON; every one given must hold: `group` (slug), `organization_id` with optional `org_role`, `app_role`, `email_domain` (verified addresses only).
- `entitlement_ids` (Required) Ids of the entitlements the rule gives.
- `revoke_on_unmatch` (Optional) Take the access back when a person stops matching. Defaults to `true`.
- `enabled` (Optional) Defaults to `true`.

## Attribute Reference

- `id` The rule id.

## Import

```shell
terraform import atlas_iga_birthright_rule.engineering igb_…
```
