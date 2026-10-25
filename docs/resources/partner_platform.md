# atlas_partner_platform

This instance as a **partner platform** of another product: a product that signs the events it delivers to you (for example consent events) with a published Ed25519 key set. Atlas verifies each delivery against the partner's key set and applies it to your users. One resource per partner. A preset partner fills in the defaults on create. The partner token is **write-only** (`token_wo`): it is stored once, encrypted, and never returned or kept in state. `token_wo` needs Terraform 1.11 or later.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_partner_platform" "consent" {
  partner         = "consentco"
  platform_id     = "atlas"
  jwks_url        = "https://consentco.example/.well-known/jwks.json"
  signature_header = "x-consentco-signature"
  allowed_sources = ["site-123"]

  token_wo         = var.partner_token
  token_wo_version = 1
}
```

## The partner token

`token_wo` is sent on create and whenever `token_wo_version` changes (a later send replaces the stored token). Leave both out to register the platform first and store the token later. A stored token cannot be cleared except by destroying the resource, which removes the platform and its token. `token_set` and `token_prefix` (the first characters, to tell tokens apart) show that a token is held; the token itself is never available.

`api_base_url` is sent together with the token; it takes effect when the token is sent.

## Registering with the partner

`receiver_url` (or `receiver_url_template` when the partner registers per site) is where the partner delivers events for this instance; give it to the partner. `registration_status` reads `registered` once the partner has registered it. `last_delivery_at` and `last_delivery_error` show the latest delivery.

## Import

```
terraform import atlas_partner_platform.consent consentco
```

The token cannot be imported; set `token_wo` and `token_wo_version` to store it.
