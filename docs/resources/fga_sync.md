# atlas_fga_sync

Configure the managed sync that turns organization memberships, org roles and groups into FGA tuples, so a model can say `viewer: member from organization` without mirroring anything by hand. Opt-in per instance; nothing is written until `enabled` is true with a `store_id`. Enabling starts from the current end of the change feed and queues a backfill. The sync only ever removes tuples it wrote; a tuple that existed first is left alone. Tuples the store's model has no relation for are skipped and counted. A store or a tuple name cannot change while the sync owns tuples (409): disable it and `POST /v1/fga/sync/purge` first. Emits nothing itself; its writes emit fga.tuple_written and fga.tuple_deleted.

Scopes: `fga:read` to read, `fga:write` to write.

Backend API: `PUT /v1/fga/sync` (create and update); `GET /v1/fga/sync` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_fga_sync" "example" {

}
```

## Schema

### Optional

- `enabled` (Boolean) Turn the sync on or off. Turning it on needs `store_id`, starts from the current end of the change feed and queues a backfill. Turning it off keeps the tuples.
- `store_id` (String) The FGA store the tuples are written to. Cannot change while the sync owns tuples in the current store: disable and purge first.
- `sync_memberships` (Boolean) Write `user:<id> member organization:<id>` for every organization membership.
- `sync_org_roles` (Boolean) Write `user:<id> <role> organization:<id>` for every membership, one relation per role key.
- `sync_groups` (Boolean) Write group membership (`user:<id> member group:<key>`) and group nesting (`group:<child>#member member group:<parent>`).
- `user_type` (String) The FGA type of a user. Default `user`.
- `org_type` (String) The FGA type of an organization. Default `organization`.
- `group_type` (String) The FGA type of a group. Default `group`.
- `member_relation` (String) The relation for membership. Default `member`.
- `role_relation_prefix` (String) Prepended to the role key to make the relation (`org:admin` becomes `org_admin`). Default empty.
- `group_key` (String) Which group attribute names the group object. Default `slug`.

### Read-Only

- `id` (String) Atlas object id.
- `last_synced_at` (Number) The `last_synced_at` field. (epoch ms)
- `last_error` (String) The `last_error` field.
- `managed` (String, JSON) The `managed` field. A JSON document (use `jsonencode`).
- `backfill` (String, JSON) The `backfill` field. A JSON document (use `jsonencode`).
- `last_run` (String, JSON) The `last_run` field. A JSON document (use `jsonencode`).

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_fga_sync.example fga_sync
```
