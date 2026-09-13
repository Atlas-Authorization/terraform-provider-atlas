# Terraform Provider for Atlas

Manage your [Atlas](https://atlas.dev) authentication instance as declarative
infrastructure-as-code. The provider wraps Atlas's secret-key **Backend API**
(the instance-scoped `/v1/*` endpoints an `sk_` key can call), so your OAuth
clients, SSO connections, resource servers, JWT templates, webhook endpoints,
roles, organizations and custom domains live in version control and roll out
through the same review-and-apply flow as the rest of your infrastructure.

Terraform manages **declarative configuration**, not per-user runtime data.
There are deliberately no `user` or `session` resources — those belong to the
Backend API and your application, not to a plan/apply lifecycle.

> **Community provider.** This is an independent, community-maintained provider.
> It is not an official product of, and is not affiliated with or endorsed by,
> Atlas or any hosted Atlas offering. It talks only to the documented public
> secret-key Backend API and authenticates with your instance secret key
> (`sk_...`).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.5, or
  [OpenTofu](https://opentofu.org) >= 1.6
- Go >= 1.22 (only to build from source)
- An Atlas instance secret key (`sk_...`)

## Authentication

The provider reads its configuration from the environment, which is the
preferred way to keep the secret key out of state and version control:

```sh
export ATLAS_SECRET_KEY="sk_live_..."
export ATLAS_API_URL="https://api.atlas.dev"   # optional; this is the default
```

Both can also be set inline in the provider block (`api_url`, `secret_key`),
with the inline value taking precedence over the environment. `secret_key` is
marked sensitive. Every request is sent as `Authorization: Bearer <secret_key>`.

```hcl
provider "atlas" {
  # api_url    = "https://api.atlas.dev"   # or ATLAS_API_URL
  # secret_key = "sk_live_..."             # or ATLAS_SECRET_KEY (preferred)
}
```

## Install from a local build

Until the provider is published to a registry, build it and point Terraform at
the local binary with a dev override:

```sh
cd terraform-provider-atlas
go build -o terraform-provider-atlas .
```

`~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "atlas/atlas" = "/absolute/path/to/terraform-provider-atlas"
  }
  direct {}
}
```

With a dev override in place, skip `terraform init` and run `terraform plan` /
`terraform apply` directly. Once published, consumers instead declare:

```hcl
terraform {
  required_providers {
    atlas = {
      source  = "atlas/atlas"
      version = "~> 0.1"
    }
  }
}
```

## Resources

| Resource                  | CRUD                       | Notes |
|---------------------------|----------------------------|-------|
| `atlas_oauth_client`      | Create / Read / Update / Delete | "Sign in with Atlas" relying party. `client_secret` is computed + sensitive and revealed only on create (empty for a public/PKCE client). |
| `atlas_sso_connection`    | Create / Read / Update / Delete | `oidc` / `saml` / `discourse`. Secrets (`oidc_client_secret`, `saml_idp_certificate`, `discourse_secret`) are write-only; the API never returns them, so presence is tracked via `has_secret` / `has_saml_certificate` / `has_discourse_secret`. `type` is immutable (forces replacement). |
| `atlas_resource_server`   | Create / Read / Update / Delete | Machine-token audience. `identifier` is immutable (forces replacement). |
| `atlas_jwt_template`      | Create / Read / Update / Delete | Custom-claims map. `name` is the key and is immutable (forces replacement). |
| `atlas_webhook_endpoint`  | Create / Read / Delete     | Signing `secret` revealed only on create. The Backend API has **no update route**, so `url` and `enabled_events` force replacement. |
| `atlas_role`              | Create / Read / Update / Delete | Custom org role. `key` is immutable (forces replacement); permissions are applied via a separate endpoint and unknown keys are dropped by the API. |
| `atlas_organization`      | Create / Read / Update / Delete | `created_by` (an existing user id) is required and immutable (forces replacement). |
| `atlas_domain`            | Create / Read / Delete     | Custom instance domain (`fapi` / `accounts`). No update route, so `role` and `host` force replacement. Point a CNAME at `cname_target`, then verify out of band. |

## Data sources

| Data source              | Lookup   |
|--------------------------|----------|
| `atlas_oauth_client`     | by `id`  |
| `atlas_resource_server`  | by `id`  |

## Terraform semantics

- **Sensitive fields** — `secret_key`, `client_secret`, the webhook `secret`,
  and every SSO write-only secret are marked sensitive so they are redacted in
  plan output.
- **Computed vs required** — server-assigned fields (`id`, `client_id`,
  timestamps, `has_*` flags, `cname_target`, …) are computed; identity and
  writable config are required/optional.
- **Drift detection** — every `Read` re-projects the API object onto state, so a
  change made outside Terraform surfaces as a diff on the next plan. Write-only
  secrets are intentionally left untouched on read (the API cannot return them).
- **404 on read → removed from state** — if an object was deleted out of band,
  the next refresh drops it from state and plans its recreation.
- **Import** — every resource supports `terraform import`. Most import by `id`;
  `atlas_jwt_template` imports by its `name`:

  ```sh
  terraform import atlas_oauth_client.dashboard oac_123
  terraform import atlas_jwt_template.supabase supabase
  ```

## Example

See [`examples/main.tf`](./examples/main.tf) for provider configuration plus one
of every resource and data source.

## Building & testing

```sh
make build           # go build -o terraform-provider-atlas .
make vet             # go vet ./...
make fmtcheck        # fails if any file is not gofmt-clean
make test            # unit + (auto-skipped) acceptance tests
```

`make test` (equivalently `go test ./...`) runs the client's request/error
mapping unit tests and compiles the acceptance tests, which **self-skip** unless
`TF_ACC` is set — so the suite is green with no live server.

### Acceptance tests

Framework acceptance tests live in `internal/provider/*_test.go`: a
`resource.Test` case per resource and both data sources, driving real
create / read / update / import / delete round-trips against a live instance.
They serve the provider in-process (`providerserver.NewProtocol6WithError`) —
no separately-built or registry-published binary is needed.

Because they **create and destroy real objects**, run them only against a
disposable test instance, never production:

```sh
export ATLAS_SECRET_KEY="sk_test_..."          # a real instance secret key
export ATLAS_API_URL="https://api.test.example" # that instance's Backend API
export ATLAS_TEST_USER_ID="usr_..."            # only the organization tests need it
make testacc                                    # TF_ACC=1 go test ./... -v -timeout 120m
```

Without `TF_ACC` every `TestAcc*` case is skipped; when `TF_ACC` is set,
`testAccPreCheck` fails fast if `ATLAS_SECRET_KEY` / `ATLAS_API_URL` are missing.
Write-only secrets the API never returns — the OAuth `client_secret`, the SSO
`oidc_client_secret` / `saml_idp_certificate` / `discourse_secret`, and the
webhook `secret` — are excluded from the import round-trip via
`ImportStateVerifyIgnore`. For `atlas_webhook_endpoint` and `atlas_domain` (no
Backend API update route) the tests assert **replacement** (a changed `id`)
rather than in-place update, and include a `_disappears` case that deletes the
object out of band and expects a non-empty follow-up plan.

## License

Mozilla Public License 2.0.
