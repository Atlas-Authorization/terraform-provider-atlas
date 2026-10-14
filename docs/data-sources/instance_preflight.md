# atlas_instance_preflight (Data Source)

Checks your instance setup in one call and lists everything Atlas can see is
broken or worth knowing: social providers enabled without credentials, OAuth
clients that cannot complete a sign-in, custom domains not serving yet, email
sign-in with no email provider, and more. Each finding says what is wrong and
how to fix it.

Use it with a `check` block (a warning on every plan) or a `precondition`
(fail the plan) to catch setup mistakes before they reach users.

API: `GET /v1/instance/preflight`.

## Example Usage

```hcl
data "atlas_instance_preflight" "this" {}

check "atlas_setup" {
  assert {
    condition = data.atlas_instance_preflight.this.ok
    error_message = join("\n", [
      for f in data.atlas_instance_preflight.this.findings :
      "${f.severity}: ${f.message} Fix: ${f.fix}" if f.severity == "error"
    ])
  }
}
```

## Schema

### Read-Only

- `id` (String) — Always `preflight`.
- `ok` (Boolean) — True when there are no error-severity findings.
- `checked_at` (Number) — When the check ran (epoch ms).
- `error_count` (Number) — Number of error-severity findings.
- `warning_count` (Number) — Number of warning-severity findings.
- `findings` (List of Object) — Ordered error, then warning, then info. (see
  [below for nested schema](#nestedatt--findings))

<a id="nestedatt--findings"></a>
### Nested Schema for `findings`

Read-Only:

- `severity` (String) — `error`, `warning` or `info`.
- `code` (String) — Stable identifier, e.g. `oauth_provider_missing_credentials`.
- `message` (String) — What is wrong.
- `fix` (String) — What to do about it.
- `subject` (Map of String) — Ids of what the finding is about (provider, client_id, domain_id, host, …).
