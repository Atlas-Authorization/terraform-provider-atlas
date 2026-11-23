# Store the secret Atlas verifies this provider's callbacks with (your Mailgun webhook signing key, SendGrid public key, Resend webhook secret, Vonage signature secret, MessageBird signing key, or the SNS topic ARNs to accept).
resource "atlas_message_receiver" "example" {
  provider_name = "mailgun"
  secret = "example-signing-key"
}
