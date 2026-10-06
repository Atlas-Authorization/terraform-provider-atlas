# atlas_instance_config

The singleton configuration of the Atlas instance this secret key belongs to:
its `allowed_origins` CORS list and its `auth_config`. There is exactly one per
instance, so declare a single `atlas_instance_config` resource.

`auth_config` is a **partial patch merged server-side** (never a wholesale
replacement). State keeps exactly the patch you wrote (round-tripped verbatim),
so a partial config produces a clean plan with no phantom diff; the full
server-merged result is exposed separately in the computed `auth_config_resolved`.
Destroying this resource only stops Terraform managing the config; it does not
reset the instance.

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

  # Native-app passkeys, configured with a typed block. It is folded into the
  # auth_config patch under `nativeApps`, so the FAPI host serves the matching
  # apple-app-site-association and assetlinks.json files and the WebAuthn
  # ceremony accepts the derived native app origins.
  native_apps = {
    apple_app_ids = ["LB4397Q8XJ.com.acme.app"]
    android_apps = [
      {
        package_name             = "com.acme.app"
        sha256_cert_fingerprints = ["AB:CD:EF:...:12:34"]
      },
    ]
  }
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
  server-side. State holds exactly the patch you wrote (round-tripped verbatim),
  not the server-merged object, so a partial patch does not thrash the plan. The
  full effective configuration is exposed separately as `auth_config_resolved`.
- `native_apps` (Attributes) — Typed native-app passkey association, folded into
  the `auth_config` patch under `nativeApps` before it is sent (so you configure
  it with typed HCL instead of raw JSON). A value here **takes precedence** over
  any `nativeApps` embedded in the raw `auth_config` JSON. Atlas serves the
  matching `/.well-known/apple-app-site-association` and
  `/.well-known/assetlinks.json` on the instance Frontend API host
  (`frontend_api_host`). Like `auth_config`, it is an Optional partial patch kept
  verbatim in state and is never refreshed from the server; leave it unset to
  manage `nativeApps` via raw `auth_config` (or not at all). (see [below for
  nested schema](#nestedatt--native_apps))

<a id="nestedatt--native_apps"></a>
### Nested Schema for `native_apps`

Optional:

- `apple_app_ids` (List of String) — iOS/macOS app ids in `<TeamID>.<bundleId>`
  form (e.g. `LB4397Q8XJ.com.acme.app`). Emitted as `appleAppIds`.
- `android_apps` (Attributes List) — Android apps allowed to assert passkeys.
  Emitted as `androidApps`. (see [below for nested
  schema](#nestedatt--native_apps--android_apps))

<a id="nestedatt--native_apps--android_apps"></a>
### Nested Schema for `native_apps.android_apps`

Required:

- `package_name` (String) — The Android application id / package name (e.g.
  `com.acme.app`). Emitted as `packageName`.
- `sha256_cert_fingerprints` (List of String) — SHA-256 signing-certificate
  fingerprints (colon-separated hex, from keytool or the Play signing key).
  Emitted as `sha256CertFingerprints`.

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
