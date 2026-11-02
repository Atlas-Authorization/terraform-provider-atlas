# Create a proxy app: a hostname, optional upstream URL, ordered path/method rules requiring a role, permission or entitlement, public paths, and header/token injection.
resource "atlas_proxy_app" "example" {
  name = "tf-proxy-app"
  hostname = "https://example.com/hostname"
  upstream_url = "https://example.com/upstream_url"
}
