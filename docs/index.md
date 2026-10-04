# Atlas Provider

Manage an [Atlas](https://atlasauth.net) authentication instance as declarative
infrastructure-as-code. The provider wraps Atlas's secret-key **Backend API**
(BAPI) — the instance-scoped `/v1/*` endpoints an `sk_` key can call — so your
OAuth clients, SSO connections, resource servers, JWT templates, webhook
endpoints, roles, organizations and custom domains live in version control and
roll out through the same plan/apply flow as the rest of your infrastructure.

Terraform manages **declarative configuration**, not per-user runtime data.
There are deliberately no `user` or `session` resources — those belong to the
application and the Backend API, not to a plan/apply lifecycle.

> **Community provider.** This is an independent, community-maintained provider.
> It is not an official product of, and is not affiliated with or endorsed by,
> Atlas or any hosted Atlas offering. It talks only to the documented public
> Backend API.

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
| [`atlas_domain`](resources/domain.md) | Custom instance domain (`fapi` / `accounts`). |

## Data Sources

| Data source | Description |
|-------------|-------------|
| [`atlas_oauth_client`](data-sources/oauth_client.md) | Look up an OAuth client by id. |
| [`atlas_oauth_provider`](data-sources/oauth_provider.md) | Look up a social sign-in provider by catalog key. |
| [`atlas_resource_server`](data-sources/resource_server.md) | Look up a resource server by id. |
