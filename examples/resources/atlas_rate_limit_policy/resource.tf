# Create a rate-limit policy: a limit per window for one kind of subject (API key, user, organization, machine or IP).
resource "atlas_rate_limit_policy" "example" {
  name = "tf-rate-limit-policy"
  subject = "key"
  limit = 1
  window_seconds = 1
}
