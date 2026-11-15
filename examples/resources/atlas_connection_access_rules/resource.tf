# Replace the connection's access rules (at most 100; an empty list opens the connection again).
resource "atlas_connection_access_rules" "example" {
  connection_id = "tf-connection-access-rules"
  rules = jsonencode({})
}
