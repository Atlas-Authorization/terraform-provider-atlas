# A Sign in with Atlas app. Needs the provider's platform_key (psk_...).
resource "atlas_id_app" "acme" {
  name          = "Acme"
  redirect_uris = ["https://acme.example.com/auth/callback"]

  post_logout_redirect_uris           = ["https://acme.example.com/"]
  backchannel_logout_uri              = "https://acme.example.com/auth/backchannel-logout"
  backchannel_logout_session_required = true
  allowed_scopes                      = ["openid", "profile", "email", "offline_access"]

  branding = jsonencode({
    theme        = "dark"
    accent_color = "#4f46e5"
    copy = {
      sign_in_title = "Sign in to {app}"
    }
  })

  # Change to rotate the client secret.
  rotate_secret_trigger = "2026-10"
}
