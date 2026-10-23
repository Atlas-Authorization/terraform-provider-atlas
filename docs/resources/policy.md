# atlas_policy

A policy your users accept: terms of service, a privacy policy, any key. The
policy asks nobody anything until a version is published.

**Versions are content, so they are not managed by this resource.** Publish them
from the dashboard, the Backend API or SDKs (`policies.versions`), or serve them
from an external system with [`atlas_policy_source`](policy_source.md).

Destroying the resource deletes a policy that was never published and *retires*
one with a published version (its acceptances are kept as evidence). A retired
policy is treated as gone: the next plan re-creates it, and applying revives the
retired policy that has the same `key`.

API: `POST /v1/policies`, `GET`/`PATCH`/`DELETE /v1/policies/{id}`.

## Example Usage

```hcl
resource "atlas_policy" "tos" {
  key        = "tos"
  title      = "Terms of service"
  display    = "scroll"
  applies_to = "sign_in_and_sign_up"
  reaccept   = "grace"
  grace_days = 14
}
```

## Schema

### Required

- `key` (String) — Slug, unique per instance: lowercase letters, digits, `_` and `-` (for example `tos`, `privacy`). It becomes the `policy` of every consent record. Immutable — changing it forces replacement.
- `title` (String) — The name users see (up to 200 characters).

### Optional

- `display` (String) — How the policy is presented: `checkbox` (default), `scroll` (an unskippable reader; accepting needs scrolled-to-end) or `link` (clickwrap).
- `required` (Boolean) — Whether users must accept it to sign up. Default `true`.
- `applies_to` (String) — `sign_up` (default): asked once at sign-up. `sign_in_and_sign_up`: existing users also accept new versions, per `reaccept`.
- `reaccept` (String) — When a new version takes effect for users who accepted an earlier one: `none` (default), `next_sign_in` (the next sign-in is held until they accept) or `grace` (they may keep signing in for `grace_days`). Only meaningful with `applies_to = "sign_in_and_sign_up"`.
- `grace_days` (Number) — The grace window in days (1-365). Required with `reaccept = "grace"`, and only allowed with it.

### Read-Only

- `id` (String) — Atlas object id (`pol_…`).
- `status` (String) — `active`.

The provider checks `key`, `title`, the enumerations and `grace_days` at plan time.

## Import

```sh
terraform import atlas_policy.tos tos   # by key or by id
```
