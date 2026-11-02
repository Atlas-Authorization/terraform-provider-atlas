resource "atlas_iga_sod_rule" "approve_vs_pay" {
  name          = "Approving and paying invoices"
  entitlement_a = atlas_iga_entitlement.invoice_approver.id
  entitlement_b = atlas_iga_entitlement.invoice_payer.id
  mode          = "require_approval"
  approver      = jsonencode({ type = "group", ids = ["risk"], mode = "any" })
}
