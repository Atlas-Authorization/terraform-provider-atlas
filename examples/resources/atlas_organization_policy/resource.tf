# A per-organization security policy: each field tightens the instance policy.
resource "atlas_organization_policy" "acme" {
  organization_id         = atlas_organization.acme.id
  require_mfa             = true
  sso_required            = true
  allowed_sign_in_methods = ["password", "oauth_google", "saml"]
  ip_allowlist            = ["203.0.113.0/24"]
  max_session_age_seconds = 3600
}
