# atlas_data_key_provider

Customer-managed encryption for the instance: enables it on first apply (the KMS key must already be set with the key provider) and sets the data-key cache window. Destroying the resource leaves encryption on: removing it crypto-shreds everything encrypted under the data keys, so it is never done by terraform destroy.

Scopes: `vault:manage` to read, `vault:manage` to write.

Backend API: `PUT /v1/instance/data_key_provider` (create and update); `GET /v1/instance/data_key_provider` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_data_key_provider" "example" {

}
```

## Schema

### Optional

- `cache_window_seconds` (Number) The `cache_window_seconds` field.

### Read-Only

- `id` (String) Atlas object id.
- `status` (String, JSON) The full object the API returned, as a JSON document.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_data_key_provider.example data_key_provider
```
