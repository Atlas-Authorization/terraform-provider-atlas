# atlas_proxy_app

Create a proxy app: a hostname, optional upstream URL, ordered path/method rules requiring a role, permission or entitlement, public paths, and header/token injection. The hostname is unique across the deployment.

Scopes: `proxy_apps:read` to read, `proxy_apps:write` to write.

Backend API: `POST /v1/proxy_apps` (create); `GET /v1/proxy_apps/{id}` (read); `PATCH /v1/proxy_apps/{id}` (update); `DELETE /v1/proxy_apps/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_proxy_app" "example" {
  name = "tf-proxy-app"
  hostname = "https://example.com/hostname"
  upstream_url = "https://example.com/upstream_url"
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `hostname` (String) The `hostname` field.
- `upstream_url` (String) The `upstream_url` field.

### Optional

- `rules` (String, JSON) Path / method rules requiring a role, permission or entitlement. A JSON document (use `jsonencode`).
- `public_paths` (List of String) The `public_paths` field.
- `inject_headers` (Boolean) The `inject_headers` field.
- `inject_token` (Boolean) The `inject_token` field.
- `allow_agent_tokens` (Boolean) Also let an active agent through with its machine token (for example to front an MCP gateway). Its scopes count as the permissions rules check. Default false.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_proxy_app.example <id>
```
