# Define or replace a custom action: data only (an input schema, a request template, a response mapping), never code.
resource "atlas_connection_action" "example" {
  provider_name = "tf-connection-action"
  action_name = "tf-connection-action"
  description = "tf-connection-action"
  request = jsonencode({})
}
