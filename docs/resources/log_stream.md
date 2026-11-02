# atlas_log_stream

Create a log stream. destination secrets (api_key / token / auth headers) are write-only; the cursor starts at the current event so only new events forward.

Scopes: `log_streams:read` to read, `log_streams:write` to write.

Backend API: `POST /v1/log_streams` (create); `GET /v1/log_streams/{id}` (read); `PATCH /v1/log_streams/{id}` (update); `DELETE /v1/log_streams/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_log_stream" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `type` (String) The `type` field. One of: `http`, `datadog`, `splunk`, `s3`, `gcs`, `azure_blob`, `aws_eventbridge`, `kinesis_firehose`, `cloudwatch_logs`, `elastic`, `sumo_logic`, `new_relic`, `azure_event_hubs`, `sentinel`, `gcp_pubsub`, `bigquery`, `snowflake`, `loki`, `otlp`, `kafka`. Changing it forces replacement.
- `destination` (String, JSON) The `destination` field. A JSON document (use `jsonencode`).
- `event_filter` (List of String) The `event_filter` field.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `status` (String) The `status` field.
- `cursor` (String) The `cursor` field.
- `consecutive_failures` (Number) The `consecutive_failures` field.
- `last_error` (String) The `last_error` field.
- `last_delivered_at` (Number) The `last_delivered_at` field. (epoch ms)
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_log_stream.example <id>
```
