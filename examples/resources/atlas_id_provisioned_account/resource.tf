# A reviewer login for a directory or app-store submission. The account works
# with this app only. The password is write-only (Terraform 1.11 or later):
# it is sent to Atlas and never kept in state. Change password_wo_version to
# rotate it. Leave password_wo out to let Atlas generate one (it lands in the
# sensitive generated_password).
resource "atlas_id_app" "product" {
  name          = "Product"
  redirect_uris = ["https://product.example/auth/callback"]

  # Optional: let +atlas_test addresses on a domain you verified sign in with
  # the fixed code 424242, for this app only.
  test_mode                         = false
  provisioned_verified_domains_only = true
}

resource "atlas_id_provisioned_account" "reviewer" {
  app_id = atlas_id_app.product.id
  email  = "reviewer@product.example"
  label  = "app-store reviewer"

  # Reviewers that cannot receive codes. Anyone with the password can sign in
  # to the app as this account.
  two_factor_exempt = true

  password_wo         = var.reviewer_password
  password_wo_version = 1
}
