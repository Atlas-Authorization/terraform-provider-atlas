# atlas_policy_translation

Add or replace the text of a version in one language (content_markdown or content_html, sanitised like the version itself). GET /v1/client/policies?locale= then serves it, and an acceptance records the hash of the text that was shown. A published version can gain languages but a language already there cannot change (409); a draft can change freely. 400 for an external policy.

Scopes: `instance:write` to read, `instance:write` to write.

Backend API: `PUT /v1/policies/{id}/versions/{versionId}/translations/{locale}` (create and update); no read route: refresh keeps the last applied values; `DELETE /v1/policies/{id}/versions/{versionId}/translations/{locale}` (destroy).

## Example Usage

```hcl
resource "atlas_policy_translation" "example" {
  policy_id = "tf-policy-translation"
  version_id = "tf-policy-translation"
  locale = "tf-policy-translation"
}
```

## Schema

### Required

- `policy_id` (String) The policy id this object belongs to. Changing it forces replacement.
- `version_id` (String) The version id this object belongs to. Changing it forces replacement.
- `locale` (String) The locale this object belongs to. Changing it forces replacement.

### Optional

- `title` (String) The `title` field.
- `content_markdown` (String) The `content_markdown` field.
- `content_html` (String) The `content_html` field.

### Read-Only

- `id` (String) Atlas object id.
- `policy` (String) The `policy` field.
- `version` (String) The `version` field.
- `content_sha256` (String) The `content_sha256` field.
- `created_at` (Number) Epoch ms.
- `updated_at` (Number) Epoch ms.

## Import

The import id is `<policy_id>/<version_id>/<locale>`.

```sh
terraform import atlas_policy_translation.example <policy_id>/<version_id>/<locale>
```
