# atlas_jwt_connection

A trusted-JWT connection: your own system signs a short-lived JWT for a person
and sends them to `login_url?token=<jwt>&redirect_url=<your app URL>`. Atlas
checks the signature against `jwks_uri`, the `iss`, `aud` and `exp`, the max age,
and that the `jti` was never used before, then signs the person in and seats them
in `organization_id`. An `active` connection needs `jwks_uri`, `issuer` and
`audience`. Needs a secret key with the `sso_connections:write` scope.

Backend API (BAPI): `POST /v1/jwt_connections` (create),
`GET /v1/jwt_connections/:id` (read), `PATCH /v1/jwt_connections/:id` (update),
`DELETE /v1/jwt_connections/:id` (delete).

## Example Usage

```hcl
resource "atlas_jwt_connection" "acme" {
  organization_id = atlas_organization.acme.id
  status          = "active"
  jwks_uri        = "https://auth.acme.example.com/.well-known/jwks.json"
  issuer          = "https://auth.acme.example.com"
  audience        = "atlas"
  max_age_seconds = 120
}

# Send people to: <login_url>?token=<jwt>&redirect_url=<your app URL>
output "jwt_login_url" {
  value = atlas_jwt_connection.acme.login_url
}
```

## Schema

### Optional

- `organization_id` (String) — The organization people who sign in this way are seated in. Removing it clears it.
- `status` (String) — `draft` (default), `active` or `disabled`.
- `jwks_uri` (String) — An https URL serving the signer's JSON Web Key Set (public RS / PS / ES / EdDSA keys; HS keys are ignored).
- `issuer` (String) — The exact `iss` the tokens carry.
- `audience` (String) — An `aud` value the tokens must carry.
- `email_claim` (String) — The claim carrying the email address (default `email`). `email_verified: true` lets the sign-in link to an existing account with that verified email.
- `max_age_seconds` (Number) — The longest a token may live (`exp` - `iat`), 30-3600 seconds (default 300).
- `default_role_id` (String) — The organization role a newly seated member gets (else `org:member`). Removing it clears it.

`jwks_uri`, `issuer`, `audience`, `email_claim` and `max_age_seconds` cannot be
cleared once set; removing one from the configuration keeps the stored value.

### Read-Only

- `id` (String) — The connection id.
- `login_url` (String) — Send the person here with `?token=<jwt>&redirect_url=<your app URL>`.
- `created_at` (Number) — Creation time (epoch ms).
- `updated_at` (Number) — Last update time (epoch ms).

## Import

```sh
terraform import atlas_jwt_connection.acme jwt_123
```
