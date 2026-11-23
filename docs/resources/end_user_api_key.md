# atlas_end_user_api_key

An end-user API key for a user or organization. With `credential = "key_pair"` the key has no shared secret: Atlas keeps only the public key and the holder signs short-lived assertions with the private key. Destroy revokes the key.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

A key pair generated in Terraform. Atlas only ever sees the public key, so no private key from Atlas lands in state (the `tls_private_key` resource keeps its own in state, as usual):

```hcl
resource "tls_private_key" "reports" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P256"
}

resource "atlas_end_user_api_key" "reports" {
  subject_type = "organization"
  subject_id   = "org_123"
  name         = "reports exporter"
  scopes       = ["reports:read"]
  credential   = "key_pair"
  audience     = "https://api.example.com"
  public_key   = tls_private_key.reports.public_key_pem
}

output "reports_key_id" {
  value = atlas_end_user_api_key.reports.key_id
}
```

Let Atlas generate the pair. The PKCS#8 `private_key` is returned once and is kept, marked sensitive, in Terraform state; anyone who can read the state can read it. Prefer bringing your own `public_key` as above.

```hcl
resource "atlas_end_user_api_key" "generated" {
  subject_type = "user"
  subject_id   = "user_123"
  credential   = "key_pair"
}

output "private_key" {
  value     = atlas_end_user_api_key.generated.private_key
  sensitive = true
}
```

A shared-secret key (the default, `credential = "secret"`) exposes its plaintext `secret`, also kept in state.

## Notes

- `subject_type`, `subject_id`, `credential`, `audience` and `public_key` cannot change: changing any of them replaces the key. `name` and `scopes` update in place.
- To rotate, run `terraform apply -replace=atlas_end_user_api_key.<name>`.
- Needs the `api_keys:write` scope.
- Sign assertions with the key id (`key_id`) as `kid` and `iss`; see https://atlasauth.net/docs/api-key-pairs.
