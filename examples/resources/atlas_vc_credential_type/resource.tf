# A mobile driving licence issued as an ISO 18013-5 mdoc.
resource "atlas_vc_credential_type" "mdl" {
  name      = "Driving licence"
  format    = "mso_mdoc"
  vct       = "org.iso.18013.5.1.mDL"
  namespace = "org.iso.18013.5.1"

  claims = [
    { name = "family_name", source = "last_name" },
    { name = "given_name", source = "first_name" },
    { name = "birth_date", source = "public_metadata.birth_date", value_type = "full-date" },
  ]
}

# A membership card as an SD-JWT VC; member_id is always revealed.
resource "atlas_vc_credential_type" "member" {
  name = "Member card"
  vct  = "https://credentials.example.com/member"

  claims = [
    { name = "email", source = "email" },
    { name = "member_id", source = "user_id", disclosable = false },
  ]
}
