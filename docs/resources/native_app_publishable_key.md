# atlas_native_app_publishable_key

Restrict which registered native apps may use the publishable key (an empty list means any). Destroying the resource leaves the restriction in place; set `app_ids = []` to lift it.

Scopes: `native_apps:read` to read, `native_apps:write` to write.

Backend API: `PUT /v1/native_apps/publishable_key` (create and update); `GET /v1/native_apps/publishable_key` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_native_app_publishable_key" "example" {
  app_ids = ["nativeApp_example"]
}
```

## Schema

### Required

- `app_ids` (List of String) Ids of the registered native apps that may use the publishable key; an empty list removes the restriction.

### Read-Only

- `id` (String) Atlas object id.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_native_app_publishable_key.example native_app_publishable_key
```
