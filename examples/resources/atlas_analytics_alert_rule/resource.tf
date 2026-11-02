# Create an alert rule: tell me when a sign-in health metric, or a key or organization's API usage against its baseline, crosses a threshold over a trailing window.
resource "atlas_analytics_alert_rule" "example" {
  name = "tf-analytics-alert-rule"
  metric = "sign_in_success_rate"
  comparator = "lt"
  threshold = 1
}
