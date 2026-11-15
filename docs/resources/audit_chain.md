# atlas_audit_chain

The tamper-evident audit hash chain: turns it on and sets the checkpoint interval. Destroying the resource leaves the chain on, because the chain cannot be switched off without breaking verification of the rows already chained.

Scopes: `audit:read` to read, `audit:write` to write.

Backend API: `PUT /v1/audit/chain` (create and update); `GET /v1/audit/chain` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_audit_chain" "example" {
  enabled = true
}
```

## Schema

### Required

- `enabled` (Boolean) Must be true.

### Optional

- `checkpoint_interval_seconds` (Number) Default 3600.

### Read-Only

- `id` (String) Atlas object id.
- `enabled_at` (Number) Audit rows written before this are not chained. (epoch ms)
- `head_seq` (Number) The sequence number of the newest link.
- `head_hash` (String) The `head_hash` field.
- `genesis_hash` (String) The value the first link chains from.
- `last_checkpoint_seq` (Number) The `last_checkpoint_seq` field.
- `last_checkpoint_at` (Number) Epoch ms.
- `unchained_rows` (Number) Audit rows old enough to be chained that no link covers: 0 on a healthy chain.
- `algorithm` (String, JSON) The tags the hashes use, for an independent verifier. A JSON document (use `jsonencode`).
- `jwks_uri` (String) Where the keys that sign checkpoints are published.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_audit_chain.example audit_chain
```
