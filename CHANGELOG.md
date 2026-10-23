# Changelog

All notable changes to `terraform-provider-atlas` are documented here. The
released version is set from the git tag at build time (see `.goreleaser.yml`);
tag a release `vX.Y.Z` to match an entry below.

## 0.9.0 — 2026-10-24

### Added

- **`atlas_better_auth_bridge`** — the Better Auth session bridge (a singleton per
  instance), for moving from Better Auth without signing anyone out: `enabled`,
  `ends_at` (RFC 3339), `cookie_prefix`, and `live_fallback` (`enabled`,
  `base_url`, the old app's secret as the write-only `secret_wo` with
  `secret_wo_version`, read-only `has_secret`), plus read-only `exchange_path`.
  The secret is never stored in state; it is sent on create and whenever
  `secret_wo_version` changes (a new version without `secret_wo` clears it), and
  removing the `live_fallback` block turns the fallback off and clears it.
  Destroy turns the bridge off. The cookie prefix, window, URL and secret length
  are checked at plan time. Import with `better_auth_bridge`. `secret_wo` needs
  Terraform 1.11 or later.

## 0.8.0

### Added

- **`atlas_oauth_client`**: `skip_logout_confirmation` (Boolean, default
  `false`) — an RP-initiated logout from the client ends the session without
  the "Sign out?" page. First-party clients only: the plan fails with a clear
  error when the client is third-party (`first_party = false`, or unset on a
  new client), before Atlas would answer `422`. The `atlas_oauth_client` data
  source gains `skip_logout_confirmation` (read-only).
- **`atlas_cookie_consent`**: `settings_link` (`label`, `preset`,
  `attributes`, and read-only `rendered_attributes`) — a "Cookie settings"
  link in the hosted pages' footer that reopens the CMP. Presets:
  `cookiemunch`, `cookiebot`, `onetrust`, `custom`. The label length, preset
  and attribute allowlist are checked at plan time.
- **`atlas_email_provider`**: `force_destroy` (Boolean, default `false`) —
  destroy sends `?force=true`, removing the provider even while
  email-dependent features are on. Changing it alone makes no API call; it is
  imported as `false`.

### Changed

- **`atlas_email_provider`**: when Atlas refuses a destroy with
  `409 EMAIL_PROVIDER_IN_USE`, the error names the features that need the
  provider and the two ways forward (turn them off, or set `force_destroy`).

## 0.7.0

### Added

- **`platform_key`** provider attribute (or `ATLAS_PLATFORM_KEY`) for platform
  resources. A configuration that manages only platform resources may omit
  `secret_key`.
- **`atlas_id_app`** (new resource) — Sign in with Atlas apps on the shared
  Atlas issuer: redirect and post-logout redirect URIs, back-/front-channel
  logout URIs, allowed scopes, logo / homepage URLs and `branding` (theme,
  colours, policy links, page `copy` and `design`) as a JSON object compared
  semantically. `rotate_secret_trigger` rotates the client secret in place.
- **`atlas_instance_design`** and **`atlas_organization_design`** (new
  resources) — the hosted sign-in page design as a JSON document, with the
  resolved design and contrast warnings read back.
- **`atlas_cookie_consent`** (new resource) — the cookie-consent banner (any
  CMP) on the hosted pages.
- **`atlas_entitlement_override`** (new resource) — sales-led plan / features /
  limits for a user or organization, without Stripe.
- **`atlas_email_provider`** and **`atlas_sms_provider`** (new resources) —
  bring-your-own messaging providers, one typed block per transport (Resend,
  SES, SMTP, Brevo, Mailgun; Twilio, Vonage, MessageBird, Plivo, Telnyx, Sinch,
  Infobip, ClickSend, Bandwidth, Amazon SNS). Secrets are sensitive and
  write-only.
- **`atlas_instance_preflight`** (new data source) — the instance setup check,
  for use in `check` blocks and preconditions.
- **`atlas_webhook_endpoint`**: `active` is now settable (pause, or switch an
  auto-disabled endpoint back on), `rotate_secret_trigger` rotates the signing
  secret in place, `failing_since` is read back, and the secret is read back
  from Atlas on import.
- **`atlas_oauth_provider`**: `display_name`, `icon_url` and `mode` inputs;
  `credentials` is now optional, so providers without credentials (Steam and
  the `siwe` / `siws` / `siwf` wallet sign-ins) can be configured with
  `settings` alone; `setting_values`, `native`, `atlas_callback`,
  `custom_redirect_uri` and `javascript_origins` are read back. The data
  source gains `icon_url`, `mode`, `native` and `setting_values`.
- **`atlas_organization_policy`**: `end_user_pool` (enabled,
  require_approval, allow_hosted_signup).

### Fixed

- **`atlas_webhook_endpoint`**: an endpoint Atlas switched off after repeated
  delivery failures no longer disappears from state (and is no longer
  re-created on the next apply).
- **`atlas_organization_policy`**: destroying the resource now clears every
  policy key (previously the scalar overrides were left in place).

## 0.6.0

### Added

- **`atlas_notification_category`** (new resource) — manages tenant-defined
  end-user notification categories (`/v1/notification_categories`): a `key` and
  a user-facing `label` shown on the preference toggle. A tenant category is
  always `optional` and cannot shadow a built-in; deleting one still referenced
  by a notification template is rejected (`409`). Import by key.

### Documentation

- Added resource docs for **`atlas_org_settings_schema`**,
  **`atlas_notification_template`** and **`atlas_email_template`** (all shipped
  in 0.5.0 without pages) and for the new **`atlas_notification_category`**,
  plus matching `examples/resources/` HCL.
- Documented the **`atlas_billing_plan.limits`** map attribute (added in 0.5.0)
  on the billing-plan resource page and example.

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
