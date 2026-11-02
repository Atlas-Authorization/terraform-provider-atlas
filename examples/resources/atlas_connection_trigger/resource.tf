# Subscribe to provider events.
resource "atlas_connection_trigger" "example" {
  connection_id = "tf-connection-trigger"
  event = "tf-connection-trigger"
}
