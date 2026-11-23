# atlas_pca_action_reviewer

A reviewer of agent action changes. Creating it makes the reviewer record and a single-use enrolment link (valid 24 hours); the reviewer enrols a passkey on that page, outside Terraform. Destroying the resource unpins the reviewer key and revokes its passkeys.

Scopes: `pca:read` to read, `pca:write` to write.

Backend API: `POST /v1/pca/action-registry/reviewers/enrolments` (create); `GET /v1/pca/action-registry/reviewers` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/pca/action-registry/reviewers/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_pca_action_reviewer" "example" {
  display_name = "Security reviewer"
  email        = "reviewer@example.com"
}
```

## Schema

### Optional

- `display_name` (String) The reviewer name.
- `email` (String) The reviewer email.
- `user_id` (String) The instance user the reviewer is.

### Read-Only

- `id` (String) Atlas object id.
- `enrolment_url` (String, Sensitive) The single-use enrolment link. Returned once, on create.
- `kid` (String) The reviewer key id, once enrolled.
- `pk` (String) The reviewer public key, once enrolled.
- `status` (String) The reviewer status.
- `created_at` (String) Creation time.

## Import

The import id is `<id>`.

```sh
terraform import atlas_pca_action_reviewer.example <id>
```
