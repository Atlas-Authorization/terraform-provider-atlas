resource "atlas_iga_birthright_rule" "engineering" {
  name              = "Engineering gets the wiki"
  match             = jsonencode({ group = "engineering" })
  entitlement_ids   = [atlas_iga_entitlement.wiki.id]
  revoke_on_unmatch = true
}
