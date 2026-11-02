# Sign people in from a short-lived JWT your own system signs, seated in an organization.
resource "atlas_jwt_connection" "acme" {
  organization_id = atlas_organization.acme.id
  status          = "active"
  jwks_uri        = "https://auth.acme.example.com/.well-known/jwks.json"
  issuer          = "https://auth.acme.example.com"
  audience        = "atlas"
  max_age_seconds = 120
}

# Send people to: <login_url>?token=<jwt>&redirect_url=<your app URL>
output "jwt_login_url" {
  value = atlas_jwt_connection.acme.login_url
}
