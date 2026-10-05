# Changelog

All notable changes to `terraform-provider-atlas` are documented here. The
released version is set from the git tag at build time (see `.goreleaser.yml`);
tag a release `vX.Y.Z` to match an entry below.

## 0.5.0

### Added

- **`atlas_billing_plan.limits`** — an optional map of limit name to number
  (e.g. `{ max_devices = 5 }`): the numeric caps a plan grants, alongside
  `features`. Unset keeps the current limits; `{}` clears them.
- **`atlas_org_settings_schema`** (new resource) — manages the per-instance
  org-settings JSON Schema registry (`/v1/organization_settings_schema`), with a
  computed `version`. Destroy only removes it from state (the API has no delete).
- **`atlas_notification_template`** (new resource) — manages tenant
  notification templates (`/v1/notification_templates`): name, subject, body
  and category. Import by name.
- **`atlas_email_template`** (new resource) — manages the copy override
  (subject/text) of a single built-in email template through the dedicated
  `/v1/email_templates/:name` endpoint, which validates `{{placeholders}}` at
  save time. Destroy reverts to the built-in copy. Import by template name.

## 0.4.0

### Changed

- **`atlas_webhook_endpoint` now updates `url` and `enabled_events` IN PLACE.**
  Both attributes were previously `RequiresReplace`, so any change destroyed and
  recreated the endpoint and RE-ISSUED its signing secret, breaking delivery
  verification. The Backend API now exposes `PATCH /v1/webhook_endpoints/:id`,
  which updates the endpoint without rotating the secret. The provider calls it
  from a real `Update` and carries the create-time `secret` forward from prior
  state (the PATCH response never returns it), so changing the URL or the event
  list no longer rotates the secret or re-subscribes.

### Fixed

- **`atlas_instance_config.auth_config` no longer causes a phantom diff or an
  "inconsistent result after apply".** `auth_config` is a partial patch, but
  Atlas stores and returns the full server-merged object. The provider now keeps
  in state exactly the patch you wrote (round-tripped verbatim) rather than the
  merged value, so `plan` == `apply` with no drift. The v0.3.0 subset plan
  modifier — which kept the prior state in the plan but then clashed with the
  merged value read back on apply, producing Terraform's "Provider produced
  inconsistent result after apply" error — has been removed. The full merged
  configuration remains available through the computed `auth_config_resolved`.

## 0.3.0

### Fixed

- **`atlas_billing_plan.amount` documentation corrected.** The docs claimed
  `amount` was the "smallest currency unit (cents)". Atlas stores `amount` as an
  opaque free-form string verbatim — it is not parsed, validated or converted.
  The attribute, resource docs and client comment now say so and tell you to use
  whatever units you display (e.g. `12.00` for $12, or `1200` for cents).
- **`atlas_instance_config.auth_config` no longer shows a perpetual diff.**
  `auth_config` is a partial patch, but Atlas returns the full server-merged
  object, so the refreshed state never byte-matched the config literal. A new
  plan modifier suppresses the diff when your config is a subset of the merged
  value already in state, and still shows a diff for a genuine change.

### Added

- **`atlas_instance_config.auth_config_resolved`** (computed, read-only) exposes
  the full effective configuration Atlas resolved from your `auth_config` plus
  defaults, while `auth_config` continues to hold your partial patch.
