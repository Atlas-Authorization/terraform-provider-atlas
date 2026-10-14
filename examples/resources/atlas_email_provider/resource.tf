# Send email through your own Resend account.
resource "atlas_email_provider" "this" {
  from_address = "Acme <no-reply@acme.example.com>"
  resend = {
    api_key = var.resend_api_key
  }
}

variable "resend_api_key" {
  type      = string
  sensitive = true
}
