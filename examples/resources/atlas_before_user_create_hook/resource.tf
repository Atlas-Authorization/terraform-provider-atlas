# Call your service before every sign-up; refuse the sign-up when it is unreachable.
resource "atlas_before_user_create_hook" "this" {
  enabled                   = true
  url                       = "https://app.example.com/atlas/before-user-create"
  signing_secret_wo         = var.hook_signing_secret
  signing_secret_wo_version = 1
  timeout_ms                = 2000
  fail_closed               = true
}
