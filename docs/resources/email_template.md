# atlas_email_template

The copy override for ONE built-in transactional email template (for example
`verification_code` or `magic_link`). Atlas ships a default subject and body for
each system email; this resource overrides the `subject` and/or plain-text
`text` for a single template. Atlas validates `{{placeholders}}` at save time,
rejecting an override that introduces an unknown variable. Destroying the
resource reverts the template to its built-in copy.

Backend API (BAPI): `GET /v1/email_templates` (list — the read filters this
list by `name`, since there is no get-by-name route),
`PUT /v1/email_templates/:name` (save the override — validates
`{{placeholders}}`), `DELETE /v1/email_templates/:name` (revert to the built-in
copy).

## Example Usage

```hcl
resource "atlas_email_template" "verification" {
  name    = "verification_code"
  subject = "Your {{app_name}} verification code"
  text    = "Your code is {{code}}. It expires in {{ttl_minutes}} minutes."
}
```

## Schema

### Required

- `name` (String) — The built-in template to override, e.g.
  `verification_code`, `magic_link`, `password_reset`. Immutable — changing it
  forces replacement. Also the import id.

### Optional

- `subject` (String) — Override subject line (max 200 characters). May contain
  `{{variable}}` placeholders.
- `text` (String) — Override plain-text body (max 5000 characters). May contain
  `{{variable}}` placeholders.

### Read-Only

- `id` (String) — Equals `name`.

## Import

```sh
terraform import atlas_email_template.verification verification_code
```

The import id is the template `name`.
