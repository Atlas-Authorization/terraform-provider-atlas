# Create the Stripe Connect (Express) account a user or organization is paid out to, on YOUR Stripe account.
resource "atlas_billing_connect_account" "example" {
  subject_type = "user"
  subject_id = "tf-billing-connect-account"
  idempotency_key = "tf-billing-connect-account"
}
