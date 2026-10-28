# Create or replace a custom client policy profile (rules: PKCE, PAR, sender constraint, signed request object, allowed auth methods and signing algorithms, https-only redirects, allowed claims).
resource "atlas_oauth_client_profile" "example" {
  name = "tf-oauth-client-profile"
}
