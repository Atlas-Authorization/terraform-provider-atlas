# atlas_notification_category

A tenant-defined end-user notification category. The `key` identifies the
category; the `label` is shown on the end-user's notification-preference toggle.
A tenant category is always optional — it cannot shadow a built-in category
(`new_device`, `unattended_access`, `credential_change`) — and is what a
`atlas_notification_template.category` points at to gate the template against
the user's preferences. Deleting a category still referenced by a template is
rejected with `409`; re-categorise or delete the template first.

Backend API (BAPI): `POST /v1/notification_categories` (create),
`GET /v1/notification_categories` (list — the read filters this list by `key`,
since there is no get-by-key route), `PUT /v1/notification_categories/:key`
(update the label), `DELETE /v1/notification_categories/:key` (delete — `409`
if a template still uses it). Scopes: `notifications:read` / `notifications:write`.

## Example Usage

```hcl
resource "atlas_notification_category" "product_updates" {
  key   = "product_updates"
  label = "Product updates"
}

# A template can then gate itself against the category's preference toggle.
resource "atlas_notification_template" "release_note" {
  name     = "release_note"
  subject  = "What's new in {{product}}"
  body      = "Hi {{name}}, here is what shipped this week."
  category = atlas_notification_category.product_updates.key
}
```

## Schema

### Required

- `key` (String) — Category key: lowercase letters, digits and underscores,
  starting with a letter (max 64). Unique per instance and immutable — changing
  it forces replacement. Also the import id.
- `label` (String) — User-facing label (1–120 characters) shown on the
  end-user's preference toggle.

### Read-Only

- `id` (String) — Atlas object id.
- `optional` (Boolean) — Always `true`: a tenant category is opt-out-able and
  never gates a built-in.
- `created_at` (Number) — Creation time (epoch ms).

## Import

```sh
terraform import atlas_notification_category.product_updates product_updates
```

The import id is the category `key`, not the object id.
