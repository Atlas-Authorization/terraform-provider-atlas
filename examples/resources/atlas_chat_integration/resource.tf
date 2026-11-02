# Connect a Slack workspace or a Teams bot using your own app.
resource "atlas_chat_integration" "example" {
  provider_value = "slack"
  name = "tf-chat-integration"
  credentials = jsonencode({})
}
