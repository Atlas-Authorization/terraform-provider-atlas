# Create an authenticator policy for a group: `required_factors` (members must pass a second factor, or enrol one), `grace_until` (members with no factor may still sign in until then), and a `passkey` profile (user verification, attestation, attachment, AAGUID allow / block lists) applied when members register passkeys.
resource "atlas_authenticator_policy" "example" {

}
