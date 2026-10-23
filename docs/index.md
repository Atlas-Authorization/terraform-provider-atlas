# Atlas Provider

Manage an [Atlas](https://atlasauth.net) authentication instance as declarative
infrastructure-as-code. The provider wraps Atlas's secret-key **Backend API**
(BAPI) — the instance-scoped `/v1/*` endpoints an `sk_` key can call — so your
OAuth clients, SSO connections, resource servers, JWT templates, webhook
endpoints, roles, organizations (and their policies), custom domains, instance
settings, redirect-URL allowlist and billing plans live in version control and
roll out through the same plan/apply flow as the rest of your infrastructure.

Terraform manages **declarative configuration**, not per-user runtime data.
There are deliberately no `user` or `session` resources — those belong to the
application and the Backend API, not to a plan/apply lifecycle.

> **Official provider.** This is the first-party Atlas Terraform provider,
> published and maintained by Atlas. It talks only to the documented public
> Backend API and authenticates with your instance secret key (`sk_...`).

## Example Usage

```hcl
terraform {
  required_providers {
    atlas = {
      source  = "Atlas-Authorization/atlas"
      version = "~> 0.1"
    }
  }
}

provider "atlas" {
  # api_url    = "https://api.atlasauth.net"  # or ATLAS_API_URL (this is the default)
  # secret_key = "sk_live_..."            # or ATLAS_SECRET_KEY (preferred)
}
```

## Authentication

Every request is sent as `Authorization: Bearer <secret_key>` and is scoped to
the single Atlas instance that key belongs to. Provide credentials via the
environment (preferred, keeps the key out of state and VCS):

```sh
export ATLAS_SECRET_KEY="sk_live_..."
export ATLAS_API_URL="https://api.atlasauth.net"   # optional; this is the default
```

or inline in the provider block. Inline values take precedence over the
environment.

## Schema

### Optional

- `api_url` (String) — Base URL of the Atlas Backend API. Defaults to
  `https://api.atlasauth.net`, or `ATLAS_API_URL` when set.
- `secret_key` (String, Sensitive) — Atlas instance secret key (`sk_...`).
  Prefer `ATLAS_SECRET_KEY` so the key never lands in state or config.
- `platform_key` (String, Sensitive) — Atlas platform key (`psk_...`), needed
  only by platform resources such as `atlas_id_app`. Prefer
  `ATLAS_PLATFORM_KEY`. A configuration that manages only platform resources
  may omit `secret_key`.

## Resources

| Resource | Description |
|----------|-------------|
| [`atlas_oauth_client`](resources/oauth_client.md) | "Sign in with Atlas" relying party (OAuth/OIDC client). |
| [`atlas_oauth_provider`](resources/oauth_provider.md) | Social sign-in provider config (Google, GitHub, ...). |
| [`atlas_sso_connection`](resources/sso_connection.md) | Enterprise SSO connection (OIDC / SAML / Discourse). |
| [`atlas_resource_server`](resources/resource_server.md) | Machine-token API (audience + scopes). |
| [`atlas_jwt_template`](resources/jwt_template.md) | Named custom-claims map injected into tokens. |
| [`atlas_webhook_endpoint`](resources/webhook_endpoint.md) | Signed-event delivery endpoint. |
| [`atlas_role`](resources/role.md) | Custom organization role and its permissions. |
| [`atlas_organization`](resources/organization.md) | Organization (tenant) profile. |
| [`atlas_organization_policy`](resources/organization_policy.md) | Per-org security policy (MFA, SSO, sign-in methods, IP allow-list, session age). |
| [`atlas_domain`](resources/domain.md) | Custom instance domain (`fapi` / `accounts`). |
| [`atlas_instance_config`](resources/instance_config.md) | Singleton instance config: allowed origins + auth config. |
| [`atlas_redirect_url`](resources/redirect_url.md) | Allowlisted OAuth/SSO redirect URL. |
| [`atlas_billing_plan`](resources/billing_plan.md) | End-user billing plan (features, Stripe price, pricing model). |
| [`atlas_org_settings_schema`](resources/org_settings_schema.md) | JSON Schema for organization settings. |
| [`atlas_notification_template`](resources/notification_template.md) | Tenant-defined notification email template. |
| [`atlas_notification_category`](resources/notification_category.md) | Tenant-defined notification category (preference toggle). |
| [`atlas_email_template`](resources/email_template.md) | Override of a built-in transactional email template. |
| [`atlas_entitlement_override`](resources/entitlement_override.md) | Sales-led plan / features / limits for a user or organization, without Stripe. |
| [`atlas_instance_design`](resources/instance_design.md) | Hosted sign-in page design for the instance. |
| [`atlas_organization_design`](resources/organization_design.md) | An organization's design override. |
| [`atlas_cookie_consent`](resources/cookie_consent.md) | Cookie-consent banner (any CMP) on the hosted pages. |
| [`atlas_email_provider`](resources/email_provider.md) | Your own email provider (Resend, SES, SMTP, Brevo, Mailgun). |
| [`atlas_sms_provider`](resources/sms_provider.md) | Your own SMS provider (Twilio, Vonage, …). |
| [`atlas_id_app`](resources/id_app.md) | Sign in with Atlas app (needs `platform_key`). |
| [`atlas_better_auth_bridge`](resources/better_auth_bridge.md) | Better Auth session bridge: migration window, cookie prefix, live fallback (write-only secret). |
| [`atlas_policy`](resources/policy.md) | A policy users accept (terms, privacy): key, title, display, applies_to, reaccept, grace_days. |
| [`atlas_policy_source`](resources/policy_source.md) | External policy source: endpoint, sync mode, write-only signing secret. |
| [`atlas_policy_settings`](resources/policy_settings.md) | Policy settings: the `pol` session claim. |

## Data Sources

| Data source | Description |
|-------------|-------------|
| [`atlas_oauth_client`](data-sources/oauth_client.md) | Look up an OAuth client by id. |
| [`atlas_oauth_provider`](data-sources/oauth_provider.md) | Look up a social sign-in provider by catalog key. |
| [`atlas_resource_server`](data-sources/resource_server.md) | Look up a resource server by id. |
| [`atlas_instance_preflight`](data-sources/instance_preflight.md) | Setup check: everything Atlas can see is broken or worth knowing. |
