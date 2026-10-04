# An allowlisted OAuth/SSO redirect target. The API normalises the stored value
# (bare scheme://host[:port]); supply an already-normalised URL to avoid a diff.
resource "atlas_redirect_url" "app_callback" {
  url = "https://app.acme.example.com/callback"
}
