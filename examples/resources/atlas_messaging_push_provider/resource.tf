# The push sender used for push MFA and in-app sign-in approval.
resource "atlas_messaging_push_provider" "example" {
  transport = "expo"
  config    = jsonencode({ access_token = var.expo_access_token })
}
