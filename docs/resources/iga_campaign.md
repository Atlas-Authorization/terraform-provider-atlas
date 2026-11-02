# atlas_iga_campaign

Create a draft campaign scoped to everything, specific entitlements, or the entitlements for an app, role, group or organization, with reviewers (the person's manager, the entitlement owner, or named users), a due date and reminder days. With `auto_revoke` the access reviewers mark for revocation is removed when the campaign closes.

Scopes: `iga:read` to read, `iga:write` to write.

Backend API: `POST /v1/iga/campaigns` (create); `GET /v1/iga/campaigns/{id}` (read); no update route: every change replaces the object; `DELETE /v1/iga/campaigns/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_iga_campaign" "example" {
  name = "tf-iga-campaign"
}
```

## Schema

### Required

- `name` (String) The `name` field.

### Optional

- `description` (String) The `description` field.
- `scope_type` (String) The `scope_type` field. One of: `all`, `entitlement`, `app`, `role`, `group`, `org`.
- `scope_ids` (List of String) The `scope_ids` field.
- `reviewer_type` (String) The `reviewer_type` field. One of: `manager`, `owner`, `named`.
- `reviewer_ids` (List of String) The `reviewer_ids` field.
- `due_at` (Number) The `due_at` field. (epoch ms)
- `reminder_days` (String, JSON) The `reminder_days` field. A JSON document (use `jsonencode`).
- `auto_revoke` (Boolean) The `auto_revoke` field.
- `stages` (String, JSON) The `stages` field. A JSON document (use `jsonencode`).
- `recurrence` (String, JSON) The `recurrence` field. A JSON document (use `jsonencode`).
- `item_filter` (String, JSON) The `item_filter` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `status` (String) The `status` field. One of: `draft`, `scheduled`, `active`, `closed`.
- `next_run_at` (Number) The `next_run_at` field. (epoch ms)
- `parent_campaign_id` (String) The `parent_campaign_id` field.
- `launched_at` (Number) The `launched_at` field. (epoch ms)
- `closed_at` (Number) The `closed_at` field. (epoch ms)
- `created_at` (Number) Created (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_iga_campaign.example <id>
```
