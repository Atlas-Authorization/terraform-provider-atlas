# atlas_cloud_access_target

Add a cloud access target. An AWS or Google Cloud target becomes a launcher tile that signs people in to that account and role (AWS: the Role, RoleSessionName, SessionDuration and PrincipalTag attributes) or workforce pool; assignment limits who sees it. Azure stores a federated-credential trust recipe. Config: aws: account_id (12 digits), role_name, saml_provider_name, session_duration (900-43200 s), session_tags {tag key: user field}, transitive_tag_keys, partition (aws, aws-us-gov, aws-cn). gcp: pool_id, provider_id (a Workforce Identity Federation SAML provider). azure: app_object_id, subject, credential_name.

Scopes: `app_catalog:read` to read, `app_catalog:write` to write.

Backend API: `POST /v1/cloud_access/targets` (create); `GET /v1/cloud_access/targets/{id}` (read); `PATCH /v1/cloud_access/targets/{id}` (update); `DELETE /v1/cloud_access/targets/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_cloud_access_target" "example" {
  provider_value = "aws"
  name = "tf-cloud-access-target"
  config = jsonencode({})
}
```

## Schema

### Required

- `provider_value` (String) The `provider` field. One of: `aws`, `gcp`, `azure`. Changing it forces replacement.
- `name` (String) The tile label, for example "Production / Admin".
- `config` (String, JSON) aws: account_id (12 digits), role_name, saml_provider_name, session_duration (900-43200 s), session_tags {tag key: user field}, transitive_tag_keys, partition (aws, aws-us-gov, aws-cn). gcp: pool_id, provider_id (a Workforce Identity Federation SAML provider). azure: app_object_id, subject, credential_name. A JSON document (use `jsonencode`).

### Optional

- `assignment` (String, JSON) Same shape as an installed app's assignment. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `launch_url` (String) IdP-initiated sign-in to the cloud (AWS and Google Cloud).
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_cloud_access_target.example <id>
```
