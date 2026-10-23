# atlas_policy_source

The external policy source — a singleton per instance. Connect the system that
serves your policy documents (a CMS, a docs platform, your own service). Atlas
fetches `GET {endpoint_url}/policies`, signed with the signing secret, and in
`webhook` mode your system tells Atlas about a new version with a signed
`policy.published` / `policy.retired` event sent to `webhook_url`.

Each apply sets every field this resource manages. Destroying the resource
disconnects the source; the policies and their history stay. If the source is
disconnected outside Terraform, the next plan re-creates it.

Your system verifies Atlas's request and signs its events with the helpers in the
Atlas server SDKs (`verifyPolicyFeedRequest` / `signPolicySourceEvent`).

API: `GET`/`PUT`/`DELETE /v1/instance/policy_source`.

## Example Usage

```hcl
variable "policy_source_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "atlas_policy_source" "this" {
  enabled      = true
  endpoint_url = "https://cms.example.com/atlas"
  sync         = "webhook"

  signing_secret_wo         = var.policy_source_secret # write-only, never in state
  signing_secret_wo_version = 1                        # bump to send a new secret
}
```

## Schema

### Required

- `endpoint_url` (String) — The source's base URL, without query or fragment: `https` (`http` only for localhost on a development instance). Atlas calls `GET {endpoint_url}/policies`.

### Optional

- `enabled` (Boolean) — Whether Atlas syncs from the source. Enabling needs `endpoint_url` and a signing secret. Default `false`.
- `sync` (String) — `webhook` (default): your source sends signed events to `webhook_url` and Atlas re-fetches. `poll`: Atlas fetches every `poll_interval_minutes`.
- `poll_interval_minutes` (Number) — Poll mode: minutes between fetches (5-1440). Atlas's default when unset.
- `signing_secret_wo` (String, Sensitive, [Write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)) — The signing secret shared with your source (16-512 characters; a `whsec_` secret works). Never stored in state or returned by Atlas. It is sent on create and whenever `signing_secret_wo_version` changes; change the version to rotate it, or change the version with `signing_secret_wo` unset to clear it. Needs Terraform 1.11 or later.
- `signing_secret_wo_version` (Number) — Any number; changing it sends `signing_secret_wo` again (or clears the secret when `signing_secret_wo` is unset).

### Read-Only

- `id` (String) — Always `policy_source`.
- `signing_secret_set` (Boolean) — Whether Atlas holds a signing secret (never the secret).
- `webhook_url` (String) — Where your source sends `policy.published` / `policy.retired` events.
- `protocol_version` (String) — The policy source protocol version Atlas speaks.
- `last_synced_at` (Number) — When the last sync finished (epoch ms); unset before the first.
- `last_sync_ok` (Boolean) — Whether the last sync succeeded.
- `last_sync_error` (String) — Why the last sync failed, if it did.

The provider checks `endpoint_url`, `sync`, `poll_interval_minutes` and the secret length at plan time.

## Import

```sh
terraform import atlas_policy_source.this policy_source
```

An imported source has no `signing_secret_wo_version`; set one together with
`signing_secret_wo` to replace the stored secret.
