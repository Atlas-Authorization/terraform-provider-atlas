# atlas_instance_config

The singleton configuration of the Atlas instance this secret key belongs to:
its `allowed_origins` CORS list and its `auth_config`. There is exactly one per
instance, so declare a single `atlas_instance_config` resource.

`auth_config` is a **partial patch merged server-side** (never a wholesale
replacement), so manage the whole object you intend to set and expect
`auth_config` to read back the merged result. Destroying this resource only
stops Terraform managing the config; it does not reset the instance.

Backend API (BAPI): `GET /v1/instance` (read),
`PATCH /v1/instance` (create/update — upsert of `allowed_origins` + `auth_config`).

## Example Usage

```hcl
resource "atlas_instance_config" "this" {
  allowed_origins = [
    "https://app.acme.example.com",
    "https://admin.acme.example.com",
  ]

  auth_config = jsonencode({
    session = {
      inactivityTimeout = 1800000
    }
  })
}

output "effective_auth_config" {
  value = atlas_instance_config.this.auth_config_resolved
}
```

## Schema

### Optional

- `allowed_origins` (Set of String) — Exact origins (`scheme://host[:port]`, no
  wildcards, no path) permitted for the widget / JS SDK and OAuth redirect
  validation. Native-app webview origins (`tauri://`, `capacitor://`,
  `ionic://`) are accepted.
- `auth_config` (String) — The instance auth configuration as a JSON object
  string (use `jsonencode(...)`). A partial patch merged into the current config
  server-side; reads back the server-merged result. The provider suppresses the
  diff when your config is a subset of that merged value, so a partial patch
  does not thrash the plan.

### Read-Only

- `id` (String) — The Atlas instance id (singleton key).
- `auth_config_resolved` (String) — The full effective configuration Atlas
  resolved from your `auth_config` plus defaults (read-only).
- `environment` (String) — The instance environment.
- `publishable_key` (String) — The instance publishable key (`pk_...`).
- `frontend_api_host` (String) — The instance Frontend API host (read-only;
  auto-assigned by Atlas).
- `created_at` (Number) — Instance creation time (epoch ms).

## Import

```sh
terraform import atlas_instance_config.this ins_123
```

The id is the instance id; the provider always reads the singleton, so any
placeholder resolves to the same instance on the next refresh.
