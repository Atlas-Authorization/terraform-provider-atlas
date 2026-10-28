# Create a service account for the LDAPS interface: base_dn (dc=/o=/ou= parts, unique across Atlas), bind_cn, allowed_cidrs (empty = nobody may connect), allow_user_bind and size_limit.
resource "atlas_ldap_server_client" "example" {
  name = "tf-ldap-server-client"
  base_dn = "tf-ldap-server-client"
  bind_cn = "tf-ldap-server-client"
}
