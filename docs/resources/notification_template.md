# atlas_notification_template

A tenant-authored end-user notification template. `POST /v1/notifications`
resolves a template by `name` ahead of the built-ins, substituting
`{{variable}}` placeholders from the send request's `data`. The `category` ties
the template to the per-user notification-preference gate — it may be a built-in
category (`new_device`, `unattended_access`, `credential_change`) or one you
declare with `atlas_notification_category`.

Backend API (BAPI): `POST /v1/notification_templates` (create),
`GET /v1/notification_templates/:name` (read), `PUT /v1/notification_templates/:name`
(update — `name` is the URL key and immutable),
`DELETE /v1/notification_templates/:name` (delete). Scopes:
`notifications:read` / `notifications:write`.

## Example Usage

```hcl
resource "atlas_notification_template" "new_device" {
  name     = "new_device"
  subject  = "New sign-in on {{device}}"
  body     = "Hi {{name}}, we noticed a sign-in from {{device}} at {{time}}."
  category = "new_device"
}
```

## Schema

### Required

- `name` (String) — Template name, unique per instance; used to reference the
  template when sending. Immutable — changing it forces replacement. Also the
  import id.
- `subject` (String) — Notification subject. May contain `{{variable}}`
  placeholders.
- `body` (String) — Notification body. May contain `{{variable}}` placeholders.
- `category` (String) — The notification category gating the template against
  the user's preferences. A built-in (`new_device`, `unattended_access`,
  `credential_change`) or a tenant category created with
  `atlas_notification_category`.

### Read-Only

- `id` (String) — Atlas object id.
- `created_at` (Number) — Creation time (epoch ms).
- `updated_at` (Number) — Last update time (epoch ms).

## Import

```sh
terraform import atlas_notification_template.new_device new_device
```

The import id is the template `name`, not the object id.
