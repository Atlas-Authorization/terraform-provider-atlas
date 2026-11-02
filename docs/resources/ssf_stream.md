# atlas_ssf_stream

Create a stream. `delivery_method` is `push` (we POST each signed SET to `endpoint_url`, with the optional `authorization_header`) or `poll` (the receiver pulls from the returned `poll_endpoint` with the one-time `poll_token`). `events_requested` limits the events sent (empty = all supported). SETs are signed with the instance signing key.

Scopes: `ssf:read` to read, `ssf:write` to write.

Backend API: `POST /v1/ssf/streams` (create); `GET /v1/ssf/streams/{id}` (read); `PATCH /v1/ssf/streams/{id}` (update); `DELETE /v1/ssf/streams/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_ssf_stream" "example" {

}
```

## Schema

### Optional

- `delivery_method` (String) The `delivery_method` field. One of: `push`, `poll`. Changing it forces replacement.
- `endpoint_url` (String) The `endpoint_url` field.
- `authorization_header` (String) The `authorization_header` field.
- `audience` (String) The `audience` field. Changing it forces replacement.
- `description` (String) The `description` field.
- `events_requested` (List of String) The `events_requested` field.
- `default_subjects` (String) The `default_subjects` field. One of: `ALL`, `NONE`. Changing it forces replacement.

### Read-Only

- `id` (String) Atlas object id.

## Import

The import id is `<id>`.

```sh
terraform import atlas_ssf_stream.example <id>
```
