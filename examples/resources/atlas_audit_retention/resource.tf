# Set audit log retention in days (`null` returns to the default).
resource "atlas_audit_retention" "example" {
  retention_days = 1
}
