# atlas_analytics_alert_rule

Create an alert rule: tell me when a sign-in health metric, or a key or organization's API usage against its baseline, crosses a threshold over a trailing window. It must have email_contacts, a webhook_url or ssf. A webhook signing secret is returned once.

Scopes: `analytics:read` to read, `analytics:write` to write.

Backend API: `POST /v1/analytics/alert_rules` (create); `GET /v1/analytics/alert_rules/{id}` (read); `PATCH /v1/analytics/alert_rules/{id}` (update); `DELETE /v1/analytics/alert_rules/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_analytics_alert_rule" "example" {
  name = "tf-analytics-alert-rule"
  metric = "sign_in_success_rate"
  comparator = "lt"
  threshold = 1
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `metric` (String) Percent metrics are 0-100; failed_sign_ins is a count; sign_in_p95_seconds is seconds; api_usage_spike is the ratio of a key or organization's request rate in the window to its average hourly rate over the 7 days before (for example 10 = ten times its usual volume), and min_sample is the fewest requests in the window that may fire it. One of: `sign_in_success_rate`, `failed_passkey_ratio`, `failed_sign_ins`, `sign_in_p95_seconds`, `api_usage_spike`.
- `comparator` (String) The `comparator` field. One of: `lt`, `gt`.
- `threshold` (Number) The `threshold` field.

### Optional

- `window_minutes` (Number) Trailing window, 5 to 1440 (default 60).
- `min_sample` (Number) Ratios and latency are not evaluated on fewer observations (default 20).
- `cooldown_minutes` (Number) Minimum minutes between notifications while the rule stays firing (default 60).
- `email_contacts` (List of String) The `email_contacts` field.
- `webhook_url` (String) A public HTTPS endpoint. The body is signed like any Atlas webhook; the signing secret is returned once, on create.
- `enabled` (Boolean) The `enabled` field.
- `subject_type` (String) api_usage_spike only: watch secret keys, end-user API keys or organizations (null = all). One of: `api_key`, `end_user_api_key`, `organization`, `null`.
- `subject_id` (String) api_usage_spike only: watch one key or organization (needs subject_type).
- `ssf` (Boolean) Also raise a Shared Signals event (an opaque subject naming the key or organization) on every enabled stream that asks for it.

### Read-Only

- `id` (String) Atlas object id.
- `has_webhook_secret` (Boolean) The `has_webhook_secret` field.
- `state` (String) The `state` field. One of: `ok`, `firing`.
- `last_value` (Number) The `last_value` field.
- `last_evaluated_at` (Number) The `last_evaluated_at` field. (epoch ms)
- `last_triggered_at` (Number) The `last_triggered_at` field. (epoch ms)
- `last_error` (String) The `last_error` field.
- `created_at` (Number) The `created_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_analytics_alert_rule.example <id>
```
