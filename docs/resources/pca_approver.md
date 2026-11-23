# atlas_pca_approver

Settings of a passkey approver that already exists: its quorum, display name and status. An approver is created when its first member enrols a passkey on an enrolment link (`POST /v1/pca/approvers/enrolments`), a ceremony that happens outside Terraform; this resource then manages its settings. Destroying the resource leaves the approver as it is.

Scopes: `pca:read` to read, `pca:write` to write.

Backend API: `PATCH /v1/pca/approvers/{id}` (create and update); `GET /v1/pca/approvers/{id}` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_pca_approver" "example" {
  approver_id = "pcaapp_example"
  quorum      = 1
}
```

## Schema

### Required

- `approver_id` (String) The id of the existing approver. Changing it forces replacement.

### Optional

- `quorum` (Number) Distinct members whose passkey approval is required (at most the active members).
- `display_name` (String) The approver display name.
- `status` (String) One of: `active`, `disabled`. A disabled approver cannot approve.

### Read-Only

- `id` (String) Atlas object id.
- `principal_pub` (String) The principal public key the approver is bound to.
- `subject_type` (String) What the principal is bound to: user, org or machine.
- `subject_id` (String) The id of that user, org or machine.
- `members` (String, JSON) The approver members, their devices and channels, as a JSON document.
- `created_at` (String) Creation time.

## Import

The import id is `<approver_id>`.

```sh
terraform import atlas_pca_approver.example <approver_id>
```
