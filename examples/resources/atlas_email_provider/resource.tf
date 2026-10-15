# Send email through your own Resend account.
resource "atlas_email_provider" "this" {
  from_address = "Acme <no-reply@acme.example.com>"
  resend = {
    api_key = var.resend_api_key
  }

  # Remove it on destroy even while email sign-in or password reset depend on
  # it (Atlas refuses that otherwise). Leave false unless you mean it.
  force_destroy = false
}

variable "resend_api_key" {
  type      = string
  sensitive = true
}
