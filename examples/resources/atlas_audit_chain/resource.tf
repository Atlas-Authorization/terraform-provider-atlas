# The tamper-evident audit hash chain: turns it on and sets the checkpoint interval.
resource "atlas_audit_chain" "example" {
  enabled = true
}
