# Pin one subject (an API key, user, organization or machine of this instance, or one IP address) to a policy of its kind (policy_id) or to an inline limit (limit with window_seconds, optional burst).
resource "atlas_rate_limit_override" "example" {
  subject_type = "tf-rate-limit-override"
  subject_id = "tf-rate-limit-override"
}
