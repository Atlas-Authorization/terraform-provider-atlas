# atlas_region

Data residency for the instance: the pinned serving region, the failover region and FAPI isolation. Destroying the resource leaves the setting in place.

Scopes: `instance:read` to read, `instance:write` to write.

Backend API: `PUT /v1/region` (create and update); `GET /v1/region` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_region" "example" {
  pin = true
}
```

## Schema

### Optional

- `pin` (Boolean) true pins the instance to the region of the stack answering. false is refused.
- `failover_region` (String) A region id from the region table, different from the instance's own; null clears it. The instance must be pinned.
- `fapi_isolation` (String) null follows the deployment default. One of: `off`, `enforce`, `null`.

### Read-Only

- `id` (String) Atlas object id.
- `region` (String) The region the instance is pinned to, or null when unpinned.
- `pinned` (Boolean) The `pinned` field.
- `serving_region` (String) The region of the stack that answered. Null on a single-region deployment.
- `regions_enabled` (Boolean) Whether this deployment runs with regions. When false nothing is pinned or checked.
- `effective_fapi_isolation` (String) What applies now: when `enforce`, publishable-key requests are served only on the instance's own Frontend API host. One of: `off`, `enforce`.
- `regions` (String, JSON) The `regions` field. A JSON document (use `jsonencode`).

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_region.example region
```
