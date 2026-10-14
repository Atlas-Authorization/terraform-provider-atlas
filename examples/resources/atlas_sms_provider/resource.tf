# Send SMS through your own Twilio account.
resource "atlas_sms_provider" "this" {
  twilio = {
    account_sid = var.twilio_account_sid
    auth_token  = var.twilio_auth_token
    from        = "+15555550100"
  }
}

variable "twilio_auth_token" {
  type      = string
  sensitive = true
}
