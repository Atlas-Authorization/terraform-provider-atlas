# An alias for a permission key: the alias keeps working in checks and role definitions after the permission is renamed.
resource "atlas_permission_alias" "example" {
  alias          = "billing:view"
  permission_key = "invoices:read"
}
