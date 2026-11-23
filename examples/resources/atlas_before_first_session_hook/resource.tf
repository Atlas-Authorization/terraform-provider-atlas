# The `before_first_session` hook: called once per user, before their first session token, so account linking can stamp `external_id` on the user.
resource "atlas_before_first_session_hook" "example" {
  enabled        = true
  url            = "https://example.com/hooks/first-session"
  signing_secret = var.hook_signing_secret
}
